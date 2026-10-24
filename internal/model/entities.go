package model

import "time"

type Connection struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	ProviderCode string    `gorm:"size:64;not null;index" json:"providerCode"`
	BaseURL      string    `gorm:"size:512;not null" json:"baseUrl"`
	HTTPProxy    string    `gorm:"size:512" json:"httpProxy"`
	APIKeyEnc    string    `gorm:"size:2048" json:"-"`
	APIKeyHint   string    `gorm:"size:16" json:"apiKeyHint"`
	Enabled      int       `gorm:"not null;default:1" json:"enabled"`
	Remark       string    `gorm:"size:512" json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (Connection) TableName() string { return "ai_connections" }

type ModelEndpoint struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	ConnectionID uint64    `gorm:"index;not null" json:"connectionId"`
	ModelID      string    `gorm:"size:128;not null" json:"modelId"`
	DisplayName  string    `gorm:"size:128" json:"displayName"`
	Capabilities string    `gorm:"size:128;not null" json:"capabilities"`
	HTTPProxy    string    `gorm:"size:512" json:"httpProxy"`
	IsDefault    int       `gorm:"not null;default:0" json:"isDefault"`
	Enabled      int       `gorm:"not null;default:1" json:"enabled"`
	Remark       string    `gorm:"size:512" json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (ModelEndpoint) TableName() string { return "ai_models" }

type UsageRecord struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	TenantID         uint64    `gorm:"index;not null" json:"tenantId"`
	ConnectionID     uint64    `gorm:"index" json:"connectionId"`
	ModelRowID       uint64    `gorm:"index" json:"modelRowId"`
	ModelName        string    `gorm:"size:128" json:"modelName"`
	CallerApp        string    `gorm:"size:64;index" json:"callerApp"`
	Capability       string    `gorm:"size:32;index" json:"capability"`
	PromptTokens     int       `json:"promptTokens"`
	CompletionTokens int       `json:"completionTokens"`
	TotalTokens      int       `json:"totalTokens"`
	Status           string    `gorm:"size:16;index" json:"status"`
	ErrorMessage     string    `gorm:"size:512" json:"errorMessage"`
	RequestBytes     int64     `json:"requestBytes"`
	ResponseBytes    int64     `json:"responseBytes"`
	LatencyMS        int       `json:"latencyMs"`
	CreatedAt        time.Time `gorm:"index" json:"createdAt"`
}

func (UsageRecord) TableName() string { return "ai_usage_records" }

// BalanceSnapshot 预留厂商余额快照，当前不主动拉取。
type BalanceSnapshot struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	ConnectionID uint64    `gorm:"index;not null" json:"connectionId"`
	Currency     string    `gorm:"size:16" json:"currency"`
	Amount       string    `gorm:"size:64" json:"amount"`
	Source       string    `gorm:"size:32" json:"source"`
	Note         string    `gorm:"size:256" json:"note"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (BalanceSnapshot) TableName() string { return "ai_balance_snapshots" }
