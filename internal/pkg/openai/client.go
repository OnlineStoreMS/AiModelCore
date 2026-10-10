package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"
)

type Client struct {
	HTTP *http.Client
}

func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 120 * time.Second}}
}

type ChatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResult struct {
	Content string
	Model   string
	Usage   Usage
}

type ImageResult struct {
	URL     string
	B64JSON string
}

func Endpoint(base, suffix string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	suffix = strings.TrimLeft(suffix, "/")
	if base == "" {
		return ""
	}
	last := path.Base(base)
	if isAPIRoot(last) {
		return base + "/" + suffix
	}
	return base + "/v1/" + suffix
}

func isAPIRoot(last string) bool {
	switch strings.ToLower(last) {
	case "v1", "v2", "v3", "v4", "openai":
		return true
	default:
		return strings.HasPrefix(strings.ToLower(last), "v1")
	}
}

func (c *Client) ListModels(ctx context.Context, base, apiKey string, headers map[string]string) ([]string, error) {
	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	status, raw, err := c.do(ctx, http.MethodGet, Endpoint(base, "models"), apiKey, headers, nil, &body)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, errStatus(status, raw, "该地址未提供 /models")
	}
	if status >= 300 {
		return nil, errStatus(status, raw, "列举模型失败")
	}
	ids := make([]string, 0, len(body.Data))
	for _, item := range body.Data {
		if item.ID != "" {
			ids = append(ids, item.ID)
		}
	}
	return ids, nil
}

func (c *Client) Chat(ctx context.Context, base, apiKey string, headers map[string]string, req ChatRequest) (*ChatResult, error) {
	var body struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	status, raw, err := c.do(ctx, http.MethodPost, Endpoint(base, "chat/completions"), apiKey, headers, req, &body)
	if err != nil {
		return nil, err
	}
	if status >= 300 {
		return nil, errStatus(status, raw, "对话调用失败")
	}
	if len(body.Choices) == 0 {
		return nil, fmt.Errorf("模型未返回内容")
	}
	return &ChatResult{
		Content: textContent(body.Choices[0].Message.Content),
		Model:   firstNonEmpty(body.Model, req.Model),
		Usage:   body.Usage,
	}, nil
}

func (c *Client) GenerateImage(ctx context.Context, base, apiKey string, headers map[string]string, model, prompt, size string) (*ImageResult, error) {
	payload := map[string]string{"model": model, "prompt": prompt}
	if size != "" {
		payload["size"] = size
	}
	return c.imageJSON(ctx, Endpoint(base, "images/generations"), apiKey, headers, payload)
}

func (c *Client) EditImage(ctx context.Context, base, apiKey string, headers map[string]string, model, prompt, size, filename string, image []byte) (*ImageResult, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("image", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(image); err != nil {
		return nil, err
	}
	_ = w.WriteField("prompt", prompt)
	_ = w.WriteField("model", model)
	if size != "" {
		_ = w.WriteField("size", size)
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, Endpoint(base, "images/edits"), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	applyAuth(req, apiKey, headers)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 300 {
		return nil, errStatus(resp.StatusCode, raw, "图片编辑失败")
	}
	return parseImage(raw)
}

func (c *Client) imageJSON(ctx context.Context, url, apiKey string, headers map[string]string, payload any) (*ImageResult, error) {
	status, raw, err := c.do(ctx, http.MethodPost, url, apiKey, headers, payload, nil)
	if err != nil {
		return nil, err
	}
	if status >= 300 {
		return nil, errStatus(status, raw, "图片生成失败")
	}
	return parseImage(raw)
}

func parseImage(raw []byte) (*ImageResult, error) {
	var body struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("解析图片响应失败")
	}
	if len(body.Data) == 0 {
		return nil, fmt.Errorf("模型未返回图片")
	}
	return &ImageResult{URL: body.Data[0].URL, B64JSON: body.Data[0].B64JSON}, nil
}

func (c *Client) do(ctx context.Context, method, url, apiKey string, headers map[string]string, payload any, out any) (int, []byte, error) {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		body = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	applyAuth(req, apiKey, headers)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if out != nil && resp.StatusCode < 300 && len(raw) > 0 {
		_ = json.Unmarshal(raw, out)
	}
	return resp.StatusCode, raw, nil
}

func applyAuth(req *http.Request, apiKey string, headers map[string]string) {
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "OSMS-AiModelCore/0.1")
	for k, v := range headers {
		if strings.TrimSpace(k) != "" {
			req.Header.Set(k, v)
		}
	}
}

func textContent(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Text != "" {
				b.WriteString(p.Text)
			}
		}
		return b.String()
	}
	return string(raw)
}

func errStatus(status int, raw []byte, fallback string) error {
	msg := compact(raw)
	if msg == "" {
		msg = fallback
	}
	return fmt.Errorf("上游 HTTP %d：%s", status, msg)
}

func compact(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 400 {
		s = s[:400]
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
