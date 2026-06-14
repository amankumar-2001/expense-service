package sqlrepo

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/models/dao"
	"gorm.io/gorm"
)

// IAutoPayRepository is the persistence contract for recurring commitments.
type IAutoPayRepository interface {
	Create(ctx context.Context, a *dao.AutoPay) error
	// List returns the user's autopays. An empty status/typ means "any".
	List(ctx context.Context, userID int64, status, typ string) ([]dao.AutoPay, error)
	// GetByID returns a single autopay scoped to the owning user.
	GetByID(ctx context.Context, userID, id int64) (*dao.AutoPay, error)
	// UpdateFields applies a partial update scoped to the owning user.
	UpdateFields(ctx context.Context, userID, id int64, fields map[string]any) error
}

type autopayRepository struct {
	db *gorm.DB
}

// NewAutoPayRepository constructs the GORM autopay repository.
func NewAutoPayRepository(db *gorm.DB) IAutoPayRepository {
	return &autopayRepository{db: db}
}

func (r *autopayRepository) Create(ctx context.Context, a *dao.AutoPay) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("create autopay: %w", err)
	}
	return nil
}

func (r *autopayRepository) List(ctx context.Context, userID int64, status, typ string) ([]dao.AutoPay, error) {
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	var out []dao.AutoPay
	if err := q.Order("deduct_day ASC, id ASC").Find(&out).Error; err != nil {
		return nil, fmt.Errorf("list autopays: %w", err)
	}
	return out, nil
}

func (r *autopayRepository) GetByID(ctx context.Context, userID, id int64) (*dao.AutoPay, error) {
	var a dao.AutoPay
	err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get autopay: %w", err)
	}
	return &a, nil
}

func (r *autopayRepository) UpdateFields(ctx context.Context, userID, id int64, fields map[string]any) error {
	res := r.db.WithContext(ctx).Model(&dao.AutoPay{}).
		Where("user_id = ? AND id = ?", userID, id).Updates(fields)
	if res.Error != nil {
		return fmt.Errorf("update autopay: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
