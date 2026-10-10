package service

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"aimodelcore/internal/dto"
	"aimodelcore/internal/model"
	"aimodelcore/internal/pkg/openai"
	"aimodelcore/internal/pkg/secret"
	"aimodelcore/internal/provider"
	"aimodelcore/internal/repo"

	"gorm.io/gorm"
)

var (
	ErrNotFound   = errors.New("记录不存在")
	ErrBadRequest = errors.New("参数不正确")
	ErrNoModel    = errors.New("没有可用模型，请先在 AI 大模型管理中心配置并启用")
	ErrNoKey      = errors.New("该接入尚未填写密钥")
)

type Service struct {
	repo   *repo.Repo
	secret string
	llm    *openai.Client
}

func New(r *repo.Repo, secretKey string) *Service {
	return &Service{repo: r, secret: secretKey, llm: openai.New()}
}

func (s *Service) Providers() []provider.Item { return provider.Catalog() }

func (s *Service) ListConnections(tenantID uint64) ([]dto.ConnectionView, error) {
	rows, err := s.repo.ListConnections(tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ConnectionView, 0, len(rows))
	for _, row := range rows {
		out = append(out, s.viewConn(row))
	}
	return out, nil
}

func (s *Service) CreateConnection(tenantID uint64, in dto.ConnectionInput) (*dto.ConnectionView, error) {
	row, err := s.buildConnection(tenantID, nil, in, true)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveConnection(row); err != nil {
		return nil, err
	}
	view := s.viewConn(*row)
	return &view, nil
}

func (s *Service) UpdateConnection(tenantID, id uint64, in dto.ConnectionInput) (*dto.ConnectionView, error) {
	existing, err := s.repo.GetConnection(tenantID, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	row, err := s.buildConnection(tenantID, existing, in, false)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveConnection(row); err != nil {
		return nil, err
	}
	view := s.viewConn(*row)
	return &view, nil
}

func (s *Service) DeleteConnection(tenantID, id uint64) error {
	return mapNotFound(s.repo.DeleteConnection(tenantID, id))
}

func (s *Service) TestConnection(ctx context.Context, tenantID, id uint64) (*dto.TestResult, error) {
	row, err := s.repo.GetConnection(tenantID, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	apiKey, err := secret.Decrypt(s.secret, row.APIKeyEnc)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return &dto.TestResult{OK: false, Message: "尚未填写密钥"}, nil
	}
	models, err := s.llm.ListModels(ctx, row.BaseURL, apiKey, nil)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "404") || strings.Contains(msg, "/models") {
			return &dto.TestResult{OK: true, Message: "地址可访问，但厂商未提供模型列表，保存后可直接调用"}, nil
		}
		return &dto.TestResult{OK: false, Message: msg}, nil
	}
	if len(models) > 20 {
		models = models[:20]
	}
	return &dto.TestResult{OK: true, Message: "连接成功", Models: models}, nil
}

func (s *Service) SuggestModels(tenantID, connectionID uint64) (int, error) {
	conn, err := s.repo.GetConnection(tenantID, connectionID)
	if err != nil {
		return 0, mapNotFound(err)
	}
	item, ok := provider.Find(conn.ProviderCode)
	if !ok || len(item.Models) == 0 {
		return 0, nil
	}
	added := 0
	for _, m := range item.Models {
		exists, err := s.repo.ModelExists(tenantID, connectionID, m.ModelID)
		if err != nil {
			return added, err
		}
		if exists {
			continue
		}
		row := &model.ModelEndpoint{
			TenantID: tenantID, ConnectionID: connectionID,
			ModelID: m.ModelID, DisplayName: m.DisplayName,
			Capabilities: strings.Join(normalizeCaps(m.Capabilities), ","),
			Enabled:      1,
		}
		if err := s.repo.SaveModel(row); err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}

func (s *Service) ListModels(tenantID, connectionID uint64) ([]dto.ModelView, error) {
	rows, err := s.repo.ListModels(tenantID, connectionID)
	if err != nil {
		return nil, err
	}
	names := map[uint64]model.Connection{}
	conns, err := s.repo.ListConnections(tenantID)
	if err != nil {
		return nil, err
	}
	for _, c := range conns {
		names[c.ID] = c
	}
	out := make([]dto.ModelView, 0, len(rows))
	for _, row := range rows {
		conn := names[row.ConnectionID]
		out = append(out, viewModel(row, conn))
	}
	return out, nil
}

func (s *Service) CreateModel(tenantID uint64, in dto.ModelInput) (*dto.ModelView, error) {
	row, err := s.buildModel(tenantID, nil, in, true)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveModel(row); err != nil {
		return nil, err
	}
	conn, _ := s.repo.GetConnection(tenantID, row.ConnectionID)
	view := viewModel(*row, deref(conn))
	return &view, nil
}

func (s *Service) UpdateModel(tenantID, id uint64, in dto.ModelInput) (*dto.ModelView, error) {
	existing, err := s.repo.GetModel(tenantID, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	row, err := s.buildModel(tenantID, existing, in, false)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveModel(row); err != nil {
		return nil, err
	}
	conn, _ := s.repo.GetConnection(tenantID, row.ConnectionID)
	view := viewModel(*row, deref(conn))
	return &view, nil
}

func (s *Service) DeleteModel(tenantID, id uint64) error {
	return mapNotFound(s.repo.DeleteModel(tenantID, id))
}

func (s *Service) Dashboard(tenantID uint64) (*dto.Dashboard, error) {
	total, enabled, err := s.repo.CountConnections(tenantID)
	if err != nil {
		return nil, err
	}
	models, err := s.repo.CountModels(tenantID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	usage, err := s.repo.TodayUsage(tenantID, start)
	if err != nil {
		return nil, err
	}
	return &dto.Dashboard{
		Connections:        total,
		EnabledConnections: enabled,
		Models:             models,
		TodayCalls:         usage.Calls,
		TodayTokens:        usage.Tokens,
		TodayErrors:        usage.Errors,
		BalanceReady:       false,
		BalanceNote:        "厂商余额与账单接口已预留，当前只统计本平台发起的调用量。",
	}, nil
}

func (s *Service) ListUsage(tenantID uint64) ([]model.UsageRecord, error) {
	return s.repo.ListUsage(tenantID, 50)
}

func (s *Service) buildConnection(tenantID uint64, existing *model.Connection, in dto.ConnectionInput, creating bool) (*model.Connection, error) {
	name := strings.TrimSpace(in.Name)
	code := strings.TrimSpace(in.ProviderCode)
	if name == "" || code == "" {
		return nil, ErrBadRequest
	}
	item, ok := provider.Find(code)
	if !ok {
		return nil, ErrBadRequest
	}
	base := strings.TrimSpace(in.BaseURL)
	if base == "" {
		base = item.DefaultBaseURL
	}
	if base == "" {
		return nil, errors.New("请填写接口地址")
	}
	row := &model.Connection{TenantID: tenantID, Name: name, ProviderCode: code, BaseURL: base, Remark: strings.TrimSpace(in.Remark), Enabled: 1}
	if existing != nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
		row.APIKeyEnc = existing.APIKeyEnc
		row.APIKeyHint = existing.APIKeyHint
		row.Enabled = existing.Enabled
	}
	if in.Enabled != nil {
		row.Enabled = *in.Enabled
	}
	key := strings.TrimSpace(in.APIKey)
	if key != "" {
		enc, err := secret.Encrypt(s.secret, key)
		if err != nil {
			return nil, err
		}
		row.APIKeyEnc = enc
		row.APIKeyHint = secret.Hint(key)
	} else if creating {
		return nil, errors.New("请填写密钥")
	}
	return row, nil
}

func (s *Service) buildModel(tenantID uint64, existing *model.ModelEndpoint, in dto.ModelInput, creating bool) (*model.ModelEndpoint, error) {
	modelID := strings.TrimSpace(in.ModelID)
	if modelID == "" || in.ConnectionID == 0 {
		return nil, ErrBadRequest
	}
	if _, err := s.repo.GetConnection(tenantID, in.ConnectionID); err != nil {
		return nil, mapNotFound(err)
	}
	caps := normalizeCaps(in.Capabilities)
	if len(caps) == 0 {
		return nil, errors.New("请至少选择一种能力")
	}
	display := strings.TrimSpace(in.DisplayName)
	if display == "" {
		display = modelID
	}
	row := &model.ModelEndpoint{
		TenantID: tenantID, ConnectionID: in.ConnectionID, ModelID: modelID,
		DisplayName: display, Capabilities: strings.Join(caps, ","), Enabled: 1,
		Remark: strings.TrimSpace(in.Remark),
	}
	if existing != nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
		row.IsDefault = existing.IsDefault
		row.Enabled = existing.Enabled
	}
	if in.IsDefault != nil {
		row.IsDefault = *in.IsDefault
	}
	if in.Enabled != nil {
		row.Enabled = *in.Enabled
	}
	if creating && in.Enabled == nil {
		row.Enabled = 1
	}
	return row, nil
}

func (s *Service) viewConn(row model.Connection) dto.ConnectionView {
	name := row.ProviderCode
	if item, ok := provider.Find(row.ProviderCode); ok {
		name = item.Name
	}
	return dto.ConnectionView{
		ID: row.ID, Name: row.Name, ProviderCode: row.ProviderCode, ProviderName: name,
		BaseURL: row.BaseURL, APIKeyHint: row.APIKeyHint, HasAPIKey: row.APIKeyEnc != "",
		Enabled: row.Enabled, Remark: row.Remark,
	}
}

func viewModel(row model.ModelEndpoint, conn model.Connection) dto.ModelView {
	caps := []string{}
	if row.Capabilities != "" {
		caps = strings.Split(row.Capabilities, ",")
	}
	return dto.ModelView{
		ID: row.ID, ConnectionID: row.ConnectionID, ConnectionName: conn.Name, ProviderCode: conn.ProviderCode,
		ModelID: row.ModelID, DisplayName: row.DisplayName, Capabilities: caps,
		IsDefault: row.IsDefault, Enabled: row.Enabled, Remark: row.Remark,
	}
}

func normalizeCaps(in []string) []string {
	allow := map[string]struct{}{"chat": {}, "vision": {}, "image": {}, "embedding": {}}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, c := range in {
		c = strings.ToLower(strings.TrimSpace(c))
		if _, ok := allow[c]; !ok {
			continue
		}
		if _, dup := seen[c]; dup {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func deref(c *model.Connection) model.Connection {
	if c == nil {
		return model.Connection{}
	}
	return *c
}

func decodeImage(in dto.ImageRequest) ([]byte, string, error) {
	if strings.TrimSpace(in.ImageBase64) != "" {
		raw := in.ImageBase64
		if i := strings.Index(raw, ","); i >= 0 && strings.Contains(raw[:i], "base64") {
			raw = raw[i+1:]
		}
		b, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, "", errors.New("图片内容无法解析")
		}
		return b, "image.png", nil
	}
	if strings.TrimSpace(in.ImageURL) == "" {
		return nil, "", nil
	}
	url := strings.TrimSpace(in.ImageURL)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, "", errors.New("图片地址仅支持 http 或 https")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", errors.New("下载原图失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, "", errors.New("下载原图失败")
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, "", err
	}
	if len(b) == 0 {
		return nil, "", errors.New("原图为空")
	}
	return b, "image.png", nil
}
