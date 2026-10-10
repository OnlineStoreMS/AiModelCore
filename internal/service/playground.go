package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"aimodelcore/internal/dto"
	"aimodelcore/internal/pkg/openai"
	"aimodelcore/internal/pkg/secret"
)

var videoRequestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func (s *Service) Playground(ctx context.Context, tenantID, modelRowID uint64, in dto.PlaygroundInput) (*dto.PlaygroundResult, error) {
	endpoint, err := s.repo.GetModel(tenantID, modelRowID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	conn, err := s.repo.GetConnection(tenantID, endpoint.ConnectionID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if conn.Enabled != 1 {
		return nil, errors.New("请先启用这条接入")
	}
	apiKey, err := secret.Decrypt(s.secret, conn.APIKeyEnc)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrNoKey
	}
	capName := strings.ToLower(strings.TrimSpace(in.Capability))
	if capName == "" {
		capName = firstCap(endpoint.Capabilities)
	}
	if capName == "" || !hasCap(endpoint.Capabilities, capName) {
		return nil, errors.New("该模型没有此能力")
	}

	ctx, traffic := openai.WithTraffic(ctx)
	started := time.Now()
	var result *dto.PlaygroundResult
	var callErr error
	switch capName {
	case "chat", "vision":
		result, callErr = s.playChat(ctx, conn.ProviderCode, conn.BaseURL, apiKey, endpoint.ModelID, capName, in)
	case "image":
		result, callErr = s.playImage(ctx, conn.BaseURL, apiKey, endpoint.ModelID, in)
	case "video":
		result, callErr = s.playVideo(ctx, conn.ProviderCode, conn.BaseURL, apiKey, endpoint.ModelID, in)
	case "voice":
		result, callErr = s.playVoice(ctx, conn.ProviderCode, conn.BaseURL, apiKey, endpoint.ModelID, in)
	default:
		return nil, errors.New("这个能力还没有原生体验")
	}
	if callErr != nil && (result == nil || result.Endpoint == "") {
		return nil, callErr
	}
	if result != nil && result.Endpoint != "" {
		recErr := callErr
		if recErr == nil && !result.OK && result.Error != "" {
			recErr = errors.New(result.Error)
		}
		s.record(tenantID, conn, endpoint, "playground", capName, started, resultUsageFromRaw(result.Raw), traffic, recErr)
	}
	if callErr != nil {
		result.OK = false
		if result.Error == "" {
			result.Error = callErr.Error()
		}
	}
	return result, nil
}

func (s *Service) playChat(ctx context.Context, provider, base, apiKey, modelID, capName string, in dto.PlaygroundInput) (*dto.PlaygroundResult, error) {
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		return nil, errors.New("请填写要发送的内容")
	}
	imageRef := ""
	if capName == "vision" {
		ref, err := visionImage(in.ImageURL, in.ImageBase64)
		if err != nil {
			return nil, err
		}
		if ref == "" {
			return nil, errors.New("识图需要一张图片")
		}
		imageRef = ref
	}
	url := openai.Endpoint(base, "chat/completions")
	payload := chatPayload(provider, modelID, strings.TrimSpace(in.System), prompt, imageRef, in.Thinking)
	started := time.Now()
	status, raw, err := s.llm.Call(ctx, http.MethodPost, url, apiKey, payload)
	if err != nil {
		return &dto.PlaygroundResult{Endpoint: url}, err
	}
	return presentChat(url, status, raw, started), nil
}

func (s *Service) playImage(ctx context.Context, base, apiKey, modelID string, in dto.PlaygroundInput) (*dto.PlaygroundResult, error) {
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		return nil, errors.New("请填写画面描述")
	}
	url := openai.Endpoint(base, "images/generations")
	payload := map[string]any{"model": modelID, "prompt": prompt}
	started := time.Now()
	status, raw, err := s.llm.Call(ctx, http.MethodPost, url, apiKey, payload)
	if err != nil {
		return &dto.PlaygroundResult{Endpoint: url}, err
	}
	out := presentUpstream(url, status, raw, started)
	if out.OK {
		out.ImageURL, out.ImageData = extractImage(raw)
	}
	return out, nil
}

func (s *Service) playVideo(ctx context.Context, provider, base, apiKey, modelID string, in dto.PlaygroundInput) (*dto.PlaygroundResult, error) {
	if provider != "xai" {
		return nil, errors.New("当前只接通了 xAI 的视频生成接口")
	}
	if id := strings.TrimSpace(in.VideoRequestID); id != "" {
		if !videoRequestIDPattern.MatchString(id) {
			return nil, errors.New("视频任务号不正确")
		}
		url := openai.Endpoint(base, "videos/"+id)
		started := time.Now()
		status, raw, err := s.llm.Call(ctx, http.MethodGet, url, apiKey, nil)
		if err != nil {
			return &dto.PlaygroundResult{Endpoint: url, RequestID: id}, err
		}
		return presentVideo(url, status, raw, started, id), nil
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		return nil, errors.New("请填写视频描述")
	}
	duration := in.Duration
	if duration <= 0 {
		duration = 5
	}
	if duration > 15 {
		duration = 15
	}
	aspect := strings.TrimSpace(in.AspectRatio)
	if aspect == "" {
		aspect = "16:9"
	}
	resolution := strings.TrimSpace(in.Resolution)
	if resolution == "" {
		resolution = "720p"
	}
	url := openai.Endpoint(base, "videos/generations")
	payload := map[string]any{
		"model": modelID, "prompt": prompt,
		"duration": duration, "aspect_ratio": aspect, "resolution": resolution,
	}
	started := time.Now()
	status, raw, err := s.llm.Call(ctx, http.MethodPost, url, apiKey, payload)
	if err != nil {
		return &dto.PlaygroundResult{Endpoint: url}, err
	}
	return presentVideo(url, status, raw, started, ""), nil
}

func (s *Service) playVoice(ctx context.Context, provider, base, apiKey, modelID string, in dto.PlaygroundInput) (*dto.PlaygroundResult, error) {
	if provider != "xai" {
		return nil, errors.New("当前只接通了 xAI 的语音接口")
	}
	if !strings.Contains(modelID, "transcribe") {
		return &dto.PlaygroundResult{
			OK:    false,
			Error: "这个模型的原生接口是实时语音对话（WebSocket），体验窗不能代替一轮通话。转录模型可以上传音频，直接调用 /v1/stt。",
		}, nil
	}
	file, filename, err := decodeAudio(in.AudioBase64, in.AudioName)
	if err != nil {
		return nil, err
	}
	if len(file) == 0 {
		return nil, errors.New("请上传一段音频")
	}
	lang := strings.TrimSpace(in.Language)
	if lang == "" {
		lang = "zh"
	}
	url := openai.Endpoint(base, "stt")
	started := time.Now()
	status, raw, err := s.llm.PostFile(ctx, url, apiKey, "file", filename, file, map[string]string{
		"model": modelID, "language": lang, "format": "true",
	})
	if err != nil {
		return &dto.PlaygroundResult{Endpoint: url}, err
	}
	out := presentUpstream(url, status, raw, started)
	if out.OK {
		out.Text = extractTextField(raw)
	}
	return out, nil
}

func chatPayload(provider, modelID, system, prompt, imageRef string, thinking *bool) map[string]any {
	messages := make([]any, 0, 2)
	if system != "" {
		messages = append(messages, map[string]string{"role": "system", "content": system})
	}
	var content any = prompt
	if imageRef != "" {
		content = []any{
			map[string]string{"type": "text", "text": prompt},
			map[string]any{"type": "image_url", "image_url": map[string]string{"url": imageRef}},
		}
	}
	messages = append(messages, map[string]any{"role": "user", "content": content})
	payload := map[string]any{
		"model": modelID, "messages": messages, "stream": false,
	}
	if useThinking(provider, thinking) {
		payload["thinking"] = map[string]string{"type": "enabled"}
		payload["reasoning_effort"] = "high"
	} else if thinking != nil && !*thinking && provider == "deepseek" {
		payload["thinking"] = map[string]string{"type": "disabled"}
	}
	return payload
}

func useThinking(provider string, thinking *bool) bool {
	if thinking != nil {
		return *thinking
	}
	return provider == "deepseek"
}

func presentChat(url string, status int, raw []byte, started time.Time) *dto.PlaygroundResult {
	out := presentUpstream(url, status, raw, started)
	if !out.OK {
		return out
	}
	out.Text, out.Reasoning = extractChat(raw)
	return out
}

func presentVideo(url string, status int, raw []byte, started time.Time, fallbackID string) *dto.PlaygroundResult {
	out := presentUpstream(url, status, raw, started)
	id, phase, progress, videoURL := extractVideo(raw)
	if id == "" {
		id = fallbackID
	}
	out.RequestID = id
	out.Phase = phase
	out.Progress = progress
	out.VideoURL = videoURL
	if out.OK && videoURL == "" && id != "" && phase != "failed" && phase != "done" {
		out.Pending = true
	}
	if phase == "failed" {
		out.OK = false
		out.Pending = false
		if out.Error == "" {
			out.Error = "视频生成失败"
		}
	}
	return out
}

func presentUpstream(url string, status int, raw []byte, started time.Time) *dto.PlaygroundResult {
	out := &dto.PlaygroundResult{
		OK:        status < 300,
		Endpoint:  url,
		Status:    status,
		LatencyMS: int(time.Since(started).Milliseconds()),
		Raw:       redactRaw(raw),
	}
	if !out.OK {
		out.Error = upstreamMessage(raw)
		if out.Error == "" {
			out.Error = fmt.Sprintf("上游 HTTP %d", status)
		}
	}
	return out
}

func extractChat(raw []byte) (text, reasoning string) {
	var body struct {
		Choices []struct {
			Message struct {
				Content          json.RawMessage `json:"content"`
				ReasoningContent json.RawMessage `json:"reasoning_content"`
				Reasoning        json.RawMessage `json:"reasoning"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &body); err != nil || len(body.Choices) == 0 {
		return "", ""
	}
	msg := body.Choices[0].Message
	text = rawText(msg.Content)
	reasoning = rawText(msg.ReasoningContent)
	if reasoning == "" {
		reasoning = rawText(msg.Reasoning)
	}
	return text, reasoning
}

func extractImage(raw []byte) (url, data string) {
	var body struct {
		Data []struct {
			URL string `json:"url"`
			B64 string `json:"b64_json"`
		} `json:"data"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", ""
	}
	if body.URL != "" {
		url = body.URL
	}
	if len(body.Data) > 0 {
		if body.Data[0].URL != "" {
			url = body.Data[0].URL
		}
		if body.Data[0].B64 != "" {
			data = imageDataURL(body.Data[0].B64)
		}
	}
	return url, data
}

func extractVideo(raw []byte) (id, phase string, progress int, videoURL string) {
	var body struct {
		RequestID string  `json:"request_id"`
		Status    string  `json:"status"`
		Progress  float64 `json:"progress"`
		Video     struct {
			URL string `json:"url"`
		} `json:"video"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", "", 0, ""
	}
	return body.RequestID, body.Status, int(body.Progress), body.Video.URL
}

func extractTextField(raw []byte) string {
	var body struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return ""
	}
	return body.Text
}

func resultUsageFromRaw(raw json.RawMessage) openai.Usage {
	var body struct {
		Usage openai.Usage `json:"usage"`
	}
	_ = json.Unmarshal(raw, &body)
	return body.Usage
}

func upstreamMessage(raw []byte) string {
	var body struct {
		Error   json.RawMessage `json:"error"`
		Message string          `json:"message"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return compactText(raw)
	}
	if len(body.Error) > 0 && string(body.Error) != "null" {
		var text string
		if err := json.Unmarshal(body.Error, &text); err == nil && text != "" {
			return text
		}
		var obj struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body.Error, &obj); err == nil && obj.Message != "" {
			return obj.Message
		}
	}
	if body.Message != "" {
		return body.Message
	}
	return compactText(raw)
}

func redactRaw(raw []byte) json.RawMessage {
	if len(bytesTrim(raw)) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		b, _ := json.Marshal(compactText(raw))
		return b
	}
	redactValue(v)
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func redactValue(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if s, ok := val.(string); ok && len(s) > 400 && (k == "b64_json" || strings.Contains(strings.ToLower(k), "base64")) {
				t[k] = fmt.Sprintf("[省略 %d 字符]", len(s))
				continue
			}
			redactValue(val)
		}
	case []any:
		for _, item := range t {
			redactValue(item)
		}
	}
}

func visionImage(imageURL, imageBase64 string) (string, error) {
	if raw := strings.TrimSpace(imageBase64); raw != "" {
		if strings.HasPrefix(raw, "data:image/") {
			if len(raw) > 12<<20 {
				return "", errors.New("图片过大")
			}
			return raw, nil
		}
		b, err := decodeB64(raw)
		if err != nil {
			return "", errors.New("图片内容无法解析")
		}
		if len(b) > 8<<20 {
			return "", errors.New("图片过大")
		}
		return imageDataURL(base64.StdEncoding.EncodeToString(b)), nil
	}
	url := strings.TrimSpace(imageURL)
	if url == "" {
		return "", nil
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", errors.New("图片地址仅支持 http 或 https")
	}
	return url, nil
}

func decodeAudio(raw, name string) ([]byte, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, "", nil
	}
	if i := strings.Index(raw, ","); i >= 0 && strings.Contains(raw[:i], "base64") {
		raw = raw[i+1:]
	}
	b, err := decodeB64(raw)
	if err != nil {
		return nil, "", errors.New("音频内容无法解析")
	}
	if len(b) > 15<<20 {
		return nil, "", errors.New("音频过大")
	}
	filename := path.Base(strings.TrimSpace(name))
	if filename == "" || filename == "." || filename == "/" {
		filename = "audio.mp3"
	}
	return b, filename, nil
}

func decodeB64(raw string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(raw)
	if err == nil {
		return b, nil
	}
	return base64.RawStdEncoding.DecodeString(raw)
}

func imageDataURL(b64 string) string {
	if strings.HasPrefix(b64, "data:") {
		return b64
	}
	mime := "image/png"
	switch {
	case strings.HasPrefix(b64, "/9j/"):
		mime = "image/jpeg"
	case strings.HasPrefix(b64, "UklGR"):
		mime = "image/webp"
	}
	return "data:" + mime + ";base64," + b64
}

func rawText(raw json.RawMessage) string {
	raw = bytesTrim(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.Text)
		}
		return b.String()
	}
	return string(raw)
}

func hasCap(caps, cap string) bool {
	return strings.Contains(","+caps+",", ","+cap+",")
}

func firstCap(caps string) string {
	for _, part := range strings.Split(caps, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return ""
}

func compactText(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 400 {
		s = s[:400]
	}
	return s
}

func bytesTrim(raw []byte) []byte {
	return []byte(strings.TrimSpace(string(raw)))
}
