package repo

import (
	"time"

	"aimodelcore/internal/model"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) ListConnections(tenantID uint64) ([]model.Connection, error) {
	var list []model.Connection
	err := r.db.Where("tenant_id = ?", tenantID).Order("id desc").Find(&list).Error
	return list, err
}

func (r *Repo) GetConnection(tenantID, id uint64) (*model.Connection, error) {
	var row model.Connection
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) SaveConnection(row *model.Connection) error {
	return r.db.Save(row).Error
}

func (r *Repo) DeleteConnection(tenantID, id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND connection_id = ?", tenantID, id).Delete(&model.ModelEndpoint{}).Error; err != nil {
			return err
		}
		res := tx.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.Connection{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *Repo) ListModels(tenantID uint64, connectionID uint64) ([]model.ModelEndpoint, error) {
	q := r.db.Where("tenant_id = ?", tenantID)
	if connectionID > 0 {
		q = q.Where("connection_id = ?", connectionID)
	}
	var list []model.ModelEndpoint
	err := q.Order("is_default desc, id desc").Find(&list).Error
	return list, err
}

func (r *Repo) GetModel(tenantID, id uint64) (*model.ModelEndpoint, error) {
	var row model.ModelEndpoint
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) FindModel(tenantID, id uint64) (*model.ModelEndpoint, error) {
	return r.GetModel(tenantID, id)
}

func (r *Repo) SaveModel(row *model.ModelEndpoint) error {
	return r.db.Save(row).Error
}

func (r *Repo) DeleteModel(tenantID, id uint64) error {
	res := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.ModelEndpoint{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repo) ModelExists(tenantID, connectionID uint64, modelID string) (bool, error) {
	var n int64
	err := r.db.Model(&model.ModelEndpoint{}).
		Where("tenant_id = ? AND connection_id = ? AND model_id = ?", tenantID, connectionID, modelID).
		Count(&n).Error
	return n > 0, err
}

func (r *Repo) ResolveModelByName(tenantID, connectionID uint64, modelName, capability string) (*model.ModelEndpoint, error) {
	q := r.db.Where("tenant_id = ? AND enabled = 1 AND model_id = ? AND capabilities LIKE ?", tenantID, modelName, "%"+capability+"%")
	if connectionID > 0 {
		q = q.Where("connection_id = ?", connectionID)
	}
	var row model.ModelEndpoint
	if err := q.Order("is_default desc, id desc").First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) ResolveModel(tenantID, connectionID, modelRowID uint64, capability string) (*model.ModelEndpoint, error) {
	q := r.db.Where("tenant_id = ? AND enabled = 1", tenantID)
	if modelRowID > 0 {
		q = q.Where("id = ?", modelRowID)
	} else {
		q = q.Where("capabilities LIKE ?", "%"+capability+"%")
		if connectionID > 0 {
			q = q.Where("connection_id = ?", connectionID)
		}
	}
	var row model.ModelEndpoint
	err := q.Order("is_default desc, id desc").First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) CountConnections(tenantID uint64) (total int64, enabled int64, err error) {
	err = r.db.Model(&model.Connection{}).Where("tenant_id = ?", tenantID).Count(&total).Error
	if err != nil {
		return
	}
	err = r.db.Model(&model.Connection{}).Where("tenant_id = ? AND enabled = 1", tenantID).Count(&enabled).Error
	return
}

func (r *Repo) CountModels(tenantID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.ModelEndpoint{}).Where("tenant_id = ? AND enabled = 1", tenantID).Count(&n).Error
	return n, err
}

func (r *Repo) AddUsage(row *model.UsageRecord) error {
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now()
	}
	return r.db.Create(row).Error
}

func (r *Repo) ListUsage(tenantID uint64, limit int) ([]model.UsageRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var list []model.UsageRecord
	err := r.db.Where("tenant_id = ?", tenantID).Order("id desc").Limit(limit).Find(&list).Error
	return list, err
}

type TodayUsage struct {
	Calls  int64
	Tokens int64
	Errors int64
}

func (r *Repo) TodayUsage(tenantID uint64, start time.Time) (TodayUsage, error) {
	var out TodayUsage
	base := r.db.Model(&model.UsageRecord{}).Where("tenant_id = ? AND created_at >= ?", tenantID, start)
	if err := base.Count(&out.Calls).Error; err != nil {
		return out, err
	}
	if err := r.db.Model(&model.UsageRecord{}).
		Where("tenant_id = ? AND created_at >= ? AND status <> ?", tenantID, start, "ok").
		Count(&out.Errors).Error; err != nil {
		return out, err
	}
	var tokens *int64
	if err := r.db.Model(&model.UsageRecord{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, start).
		Select("COALESCE(SUM(total_tokens),0)").
		Scan(&tokens).Error; err != nil {
		return out, err
	}
	if tokens != nil {
		out.Tokens = *tokens
	}
	return out, nil
}
