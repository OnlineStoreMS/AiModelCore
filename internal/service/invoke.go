package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"aimodelcore/internal/dto"
	"aimodelcore/internal/model"
	"aimodelcore/internal/pkg/openai"
	"aimodelcore/internal/pkg/secret"
)

func (s *Service) Chat(ctx context.Context, in dto.ChatRequest) (*dto.ChatResponse, error) {
	if in.TenantID == 0 || len(in.Messages) == 0 {
		return nil, ErrBadRequest
	}
	capName := strings.TrimSpace(in.Capability)
	if capName == "" {
		capName = "chat"
	}
	if capName != "chat" && capName != "vision" {
		return nil, ErrBadRequest
	}
	modelRowID := in.ModelID
	if name := strings.TrimSpace(in.Model); name != "" {
		named, err := s.repo.ResolveModelByName(in.TenantID, in.ConnectionID, name, capName)
		if err != nil {
			return nil, mapNotFound(err)
		}
		modelRowID = named.ID
	}
	ctx, traffic := openai.WithTraffic(ctx)
	started := time.Now()
	conn, endpoint, apiKey, err := s.resolve(in.TenantID, in.ConnectionID, modelRowID, capName, in.ProviderCode)
	if err != nil {
		if strings.TrimSpace(in.ProviderCode) == "private" && errors.Is(err, ErrNoModel) {
			return nil, errors.New("没有可用的私有识图模型，请在 AI 大模型管理中心接入本地模型，能力勾选识图并设为默认")
		}
		return nil, err
	}
	if code := strings.TrimSpace(in.ProviderCode); code != "" && conn.ProviderCode != code {
		return nil, errors.New("没有可用的私有识图模型，请在 AI 大模型管理中心接入本地模型，能力勾选识图并设为默认")
	}
	req := openai.ChatRequest{Model: endpoint.ModelID, MaxTokens: in.MaxTokens, Temperature: in.Temperature}
	if in.Thinking != nil && !*in.Thinking && conn.ProviderCode == "deepseek" {
		req.Thinking = &openai.ThinkingMode{Type: "disabled"}
	}
	for _, m := range in.Messages {
		role := strings.TrimSpace(m.Role)
		if role == "" {
			role = "user"
		}
		req.Messages = append(req.Messages, openai.ChatMessage{Role: role, Content: m.Content})
	}
	ctx = proxyContext(ctx, conn, endpoint)
	result, callErr := s.llm.Chat(ctx, conn.BaseURL, apiKey, nil, req)
	s.record(in.TenantID, conn, endpoint, in.CallerApp, capName, started, resultUsage(result), traffic, callErr)
	if callErr != nil {
		return nil, callErr
	}
	return &dto.ChatResponse{
		Content: result.Content, Model: result.Model,
		ConnectionID: conn.ID, ConnectionName: conn.Name,
		PromptTokens: result.Usage.PromptTokens, CompletionTokens: result.Usage.CompletionTokens, TotalTokens: result.Usage.TotalTokens,
	}, nil
}

func (s *Service) Image(ctx context.Context, in dto.ImageRequest) (*dto.ImageResponse, error) {
	if in.TenantID == 0 || strings.TrimSpace(in.Prompt) == "" {
		return nil, ErrBadRequest
	}
	modelRowID := in.ModelID
	if name := strings.TrimSpace(in.Model); name != "" {
		named, err := s.repo.ResolveModelByName(in.TenantID, in.ConnectionID, name, "image")
		if err != nil {
			return nil, mapNotFound(err)
		}
		modelRowID = named.ID
	}
	ctx, traffic := openai.WithTraffic(ctx)
	started := time.Now()
	conn, endpoint, apiKey, err := s.resolve(in.TenantID, in.ConnectionID, modelRowID, "image", "")
	if err != nil {
		return nil, err
	}
	ctx = proxyContext(ctx, conn, endpoint)
	var result *openai.ImageResult
	var callErr error
	mode := "generate"
	switch {
	case conn.ProviderCode == "dashscope" && hasImageSource(in):
		ref, refErr := dashscopeImageRef(in)
		if refErr != nil {
			return nil, refErr
		}
		mode = "edit"
		result, callErr = s.llm.EditImageGenerations(ctx, conn.BaseURL, apiKey, nil, endpoint.ModelID, in.Prompt, in.Size, ref)
	case conn.ProviderCode == "xai" && hasImageSource(in):
		ref, refErr := dashscopeImageRef(in)
		if refErr != nil {
			return nil, refErr
		}
		mode = "edit"
		result, callErr = s.llm.EditImageJSON(ctx, conn.BaseURL, apiKey, nil, endpoint.ModelID, in.Prompt, in.Size, ref)
	default:
		image, filename, err := decodeImage(in)
		if err != nil {
			return nil, err
		}
		if len(image) > 0 {
			mode = "edit"
			result, callErr = s.llm.EditImage(ctx, conn.BaseURL, apiKey, nil, endpoint.ModelID, in.Prompt, in.Size, filename, image)
		} else {
			result, callErr = s.llm.GenerateImage(ctx, conn.BaseURL, apiKey, nil, endpoint.ModelID, in.Prompt, in.Size)
		}
	}
	s.record(in.TenantID, conn, endpoint, in.CallerApp, "image", started, openai.Usage{}, traffic, callErr)
	if callErr != nil {
		return nil, callErr
	}
	return &dto.ImageResponse{
		URL: result.URL, B64JSON: result.B64JSON, Model: endpoint.ModelID,
		ConnectionID: conn.ID, ConnectionName: conn.Name, Mode: mode,
	}, nil
}

func (s *Service) resolve(tenantID, connectionID, modelRowID uint64, capability, providerCode string) (*model.Connection, *model.ModelEndpoint, string, error) {
	endpoint, err := s.repo.ResolveModel(tenantID, connectionID, modelRowID, capability, providerCode)
	if err != nil {
		return nil, nil, "", mapNotFound(err)
	}
	if endpoint == nil || !strings.Contains(","+endpoint.Capabilities+",", ","+capability+",") {
		return nil, nil, "", ErrNoModel
	}
	conn, err := s.repo.GetConnection(tenantID, endpoint.ConnectionID)
	if err != nil {
		return nil, nil, "", mapNotFound(err)
	}
	if conn.Enabled != 1 || endpoint.Enabled != 1 {
		return nil, nil, "", ErrNoModel
	}
	apiKey, err := secret.Decrypt(s.secret, conn.APIKeyEnc)
	if err != nil {
		return nil, nil, "", err
	}
	if strings.TrimSpace(apiKey) == "" && conn.ProviderCode != "private" {
		return nil, nil, "", ErrNoKey
	}
	return conn, endpoint, apiKey, nil
}

func (s *Service) record(tenantID uint64, conn *model.Connection, endpoint *model.ModelEndpoint, caller, capability string, started time.Time, usage openai.Usage, traffic *openai.Traffic, callErr error) {
	status := "ok"
	msg := ""
	if callErr != nil {
		status = "error"
		msg = callErr.Error()
		if len(msg) > 500 {
			msg = msg[:500]
		}
	}
	if caller == "" {
		caller = "unknown"
	}
	var requestBytes, responseBytes int64
	if traffic != nil {
		requestBytes = traffic.RequestBytes
		responseBytes = traffic.ResponseBytes
	}
	_ = s.repo.AddUsage(&model.UsageRecord{
		TenantID: tenantID, ConnectionID: conn.ID, ModelRowID: endpoint.ID, ModelName: endpoint.ModelID,
		CallerApp: caller, Capability: capability,
		PromptTokens: usage.PromptTokens, CompletionTokens: usage.CompletionTokens, TotalTokens: usage.TotalTokens,
		Status: status, ErrorMessage: msg,
		RequestBytes: requestBytes, ResponseBytes: responseBytes,
		LatencyMS: int(time.Since(started).Milliseconds()),
	})
}

func hasImageSource(in dto.ImageRequest) bool {
	return strings.TrimSpace(in.ImageURL) != "" || strings.TrimSpace(in.ImageBase64) != ""
}

func dashscopeImageRef(in dto.ImageRequest) (string, error) {
	if u := strings.TrimSpace(in.ImageURL); u != "" {
		if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "data:") {
			return u, nil
		}
		return "", errors.New("图片地址无效")
	}
	raw := strings.TrimSpace(in.ImageBase64)
	if raw == "" {
		return "", errors.New("请提供图片")
	}
	if strings.HasPrefix(raw, "data:") {
		return raw, nil
	}
	return "data:image/png;base64," + raw, nil
}

func resultUsage(result *openai.ChatResult) openai.Usage {
	if result == nil {
		return openai.Usage{}
	}
	return result.Usage
}
