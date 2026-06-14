package sqlrepo

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/models/dao"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IFinanceRepository persists the per-user salary inputs.
type IFinanceRepository interface {
	// Get returns the user's finance row, or ErrNotFound if salary is unset.
	Get(ctx context.Context, userID int64) (*dao.UserFinance, error)
	// Upsert creates or updates the user's salary and salary day.
	Upsert(ctx context.Context, f *dao.UserFinance) error
}

type financeRepository struct {
	db *gorm.DB
}

// NewFinanceRepository constructs the GORM finance repository.
func NewFinanceRepository(db *gorm.DB) IFinanceRepository {
	return &financeRepository{db: db}
}

func (r *financeRepository) Get(ctx context.Context, userID int64) (*dao.UserFinance, error) {
	var f dao.UserFinance
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&f).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get user finance: %w", err)
	}
	return &f, nil
}

func (r *financeRepository) Upsert(ctx context.Context, f *dao.UserFinance) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"monthly_salary", "salary_day", "updated_at"}),
	}).Create(f).Error
	if err != nil {
		return fmt.Errorf("upsert user finance: %w", err)
	}
	return nil
}
