package dto

import "encoding/json"

type ConnectionInput struct {
	Name         string `json:"name"`
	ProviderCode string `json:"providerCode"`
	BaseURL      string `json:"baseUrl"`
	APIKey       string `json:"apiKey"`
	Enabled      *int   `json:"enabled"`
	Remark       string `json:"remark"`
}

type ConnectionView struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	ProviderCode string `json:"providerCode"`
	ProviderName string `json:"providerName"`
	BaseURL      string `json:"baseUrl"`
	APIKeyHint   string `json:"apiKeyHint"`
	HasAPIKey    bool   `json:"hasApiKey"`
	Enabled      int    `json:"enabled"`
	Remark       string `json:"remark"`
}

type ModelInput struct {
	ConnectionID uint64   `json:"connectionId"`
	ModelID      string   `json:"modelId"`
	DisplayName  string   `json:"displayName"`
	Capabilities []string `json:"capabilities"`
	IsDefault    *int     `json:"isDefault"`
	Enabled      *int     `json:"enabled"`
	Remark       string   `json:"remark"`
}

type ModelView struct {
	ID             uint64   `json:"id"`
	ConnectionID   uint64   `json:"connectionId"`
	ConnectionName string   `json:"connectionName"`
	ProviderCode   string   `json:"providerCode"`
	ModelID        string   `json:"modelId"`
	DisplayName    string   `json:"displayName"`
	Capabilities   []string `json:"capabilities"`
	IsDefault      int      `json:"isDefault"`
	Enabled        int      `json:"enabled"`
	Remark         string   `json:"remark"`
}

type TestResult struct {
	OK      bool     `json:"ok"`
	Message string   `json:"message"`
	Models  []string `json:"models,omitempty"`
}

type Dashboard struct {
	Connections        int64  `json:"connections"`
	EnabledConnections int64  `json:"enabledConnections"`
	Models             int64  `json:"models"`
	TodayCalls         int64  `json:"todayCalls"`
	TodayTokens        int64  `json:"todayTokens"`
	TodayErrors        int64  `json:"todayErrors"`
	TodayRequestBytes  int64  `json:"todayRequestBytes"`
	TodayResponseBytes int64  `json:"todayResponseBytes"`
	TodayAvgLatencyMS  int64  `json:"todayAvgLatencyMs"`
	TodayMaxLatencyMS  int64  `json:"todayMaxLatencyMs"`
	BalanceReady       bool   `json:"balanceReady"`
	BalanceNote        string `json:"balanceNote"`
}

type ChatRequest struct {
	TenantID     uint64        `json:"tenantId"`
	CallerApp    string        `json:"callerApp"`
	Capability   string        `json:"capability"`
	ConnectionID uint64        `json:"connectionId"`
	ModelID      uint64        `json:"modelId"`
	Messages     []ChatMessage `json:"messages"`
	Temperature  *float64      `json:"temperature"`
	MaxTokens    int           `json:"maxTokens"`
	Model        string        `json:"model"`
	Thinking     *bool         `json:"thinking"`
}

type ChatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type ChatResponse struct {
	Content          string `json:"content"`
	Model            string `json:"model"`
	ConnectionID     uint64 `json:"connectionId"`
	ConnectionName   string `json:"connectionName"`
	PromptTokens     int    `json:"promptTokens"`
	CompletionTokens int    `json:"completionTokens"`
	TotalTokens      int    `json:"totalTokens"`
}

type ImageRequest struct {
	TenantID     uint64 `json:"tenantId"`
	CallerApp    string `json:"callerApp"`
	ConnectionID uint64 `json:"connectionId"`
	ModelID      uint64 `json:"modelId"`
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	Size         string `json:"size"`
	ImageURL     string `json:"imageUrl"`
	ImageBase64  string `json:"imageBase64"`
}

type PlaygroundInput struct {
	Capability     string `json:"capability"`
	Prompt         string `json:"prompt"`
	System         string `json:"system"`
	ImageURL       string `json:"imageUrl"`
	ImageBase64    string `json:"imageBase64"`
	AudioBase64    string `json:"audioBase64"`
	AudioName      string `json:"audioName"`
	Language       string `json:"language"`
	Thinking       *bool  `json:"thinking"`
	Duration       int    `json:"duration"`
	AspectRatio    string `json:"aspectRatio"`
	Resolution     string `json:"resolution"`
	VideoRequestID string `json:"videoRequestId"`
}

type PlaygroundResult struct {
	OK        bool            `json:"ok"`
	Endpoint  string          `json:"endpoint,omitempty"`
	Status    int             `json:"status,omitempty"`
	LatencyMS int             `json:"latencyMs,omitempty"`
	Pending   bool            `json:"pending"`
	RequestID string          `json:"requestId,omitempty"`
	Progress  int             `json:"progress,omitempty"`
	Phase     string          `json:"phase,omitempty"`
	Text      string          `json:"text,omitempty"`
	Reasoning string          `json:"reasoning,omitempty"`
	ImageURL  string          `json:"imageUrl,omitempty"`
	ImageData string          `json:"imageData,omitempty"`
	VideoURL  string          `json:"videoUrl,omitempty"`
	Raw       json.RawMessage `json:"raw,omitempty"`
	Error     string          `json:"error,omitempty"`
}

type ImageResponse struct {
	URL            string `json:"url,omitempty"`
	B64JSON        string `json:"b64Json,omitempty"`
	Model          string `json:"model"`
	ConnectionID   uint64 `json:"connectionId"`
	ConnectionName string `json:"connectionName"`
	Mode           string `json:"mode"`
}
