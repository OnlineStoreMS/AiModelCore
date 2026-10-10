package service

import (
	"context"
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
	started := time.Now()
	conn, endpoint, apiKey, err := s.resolve(in.TenantID, in.ConnectionID, in.ModelID, capName)
	if err != nil {
		return nil, err
	}
	req := openai.ChatRequest{Model: endpoint.ModelID, MaxTokens: in.MaxTokens, Temperature: in.Temperature}
	for _, m := range in.Messages {
		role := strings.TrimSpace(m.Role)
		if role == "" {
			role = "user"
		}
		req.Messages = append(req.Messages, openai.ChatMessage{Role: role, Content: m.Content})
	}
	result, callErr := s.llm.Chat(ctx, conn.BaseURL, apiKey, nil, req)
	s.record(in.TenantID, conn, endpoint, in.CallerApp, capName, started, resultUsage(result), callErr)
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
	started := time.Now()
	conn, endpoint, apiKey, err := s.resolve(in.TenantID, in.ConnectionID, in.ModelID, "image")
	if err != nil {
		return nil, err
	}
	image, filename, err := decodeImage(in)
	if err != nil {
		return nil, err
	}
	var result *openai.ImageResult
	var callErr error
	mode := "generate"
	if len(image) > 0 {
		mode = "edit"
		result, callErr = s.llm.EditImage(ctx, conn.BaseURL, apiKey, nil, endpoint.ModelID, in.Prompt, in.Size, filename, image)
	} else {
		result, callErr = s.llm.GenerateImage(ctx, conn.BaseURL, apiKey, nil, endpoint.ModelID, in.Prompt, in.Size)
	}
	s.record(in.TenantID, conn, endpoint, in.CallerApp, "image", started, openai.Usage{}, callErr)
	if callErr != nil {
		return nil, callErr
	}
	return &dto.ImageResponse{
		URL: result.URL, B64JSON: result.B64JSON, Model: endpoint.ModelID,
		ConnectionID: conn.ID, ConnectionName: conn.Name, Mode: mode,
	}, nil
}

func (s *Service) resolve(tenantID, connectionID, modelRowID uint64, capability string) (*model.Connection, *model.ModelEndpoint, string, error) {
	endpoint, err := s.repo.ResolveModel(tenantID, connectionID, modelRowID, capability)
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
	if strings.TrimSpace(apiKey) == "" {
		return nil, nil, "", ErrNoKey
	}
	return conn, endpoint, apiKey, nil
}

func (s *Service) record(tenantID uint64, conn *model.Connection, endpoint *model.ModelEndpoint, caller, capability string, started time.Time, usage openai.Usage, callErr error) {
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
	_ = s.repo.AddUsage(&model.UsageRecord{
		TenantID: tenantID, ConnectionID: conn.ID, ModelRowID: endpoint.ID, ModelName: endpoint.ModelID,
		CallerApp: caller, Capability: capability,
		PromptTokens: usage.PromptTokens, CompletionTokens: usage.CompletionTokens, TotalTokens: usage.TotalTokens,
		Status: status, ErrorMessage: msg, LatencyMS: int(time.Since(started).Milliseconds()),
	})
}

func resultUsage(result *openai.ChatResult) openai.Usage {
	if result == nil {
		return openai.Usage{}
	}
	return result.Usage
}
