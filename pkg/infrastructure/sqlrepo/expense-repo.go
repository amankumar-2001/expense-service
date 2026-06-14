// Package sqlrepo holds GORM-backed PostgreSQL repositories. Each repository is
// defined as an interface (for DI and testability) plus a concrete impl.
package sqlrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/dto/entity"
	"github.com/kharchibook/expense-service/pkg/domain/models/dao"
	"gorm.io/gorm"
)

// IExpenseRepository is the persistence contract for logged expenses.
type IExpenseRepository interface {
	Create(ctx context.Context, e *dao.Expense) error
	// List returns the user's expenses in [from, to), newest first. An empty
	// category means "all categories".
	List(ctx context.Context, userID int64, from, to time.Time, category string) ([]dao.Expense, error)
	// DeleteLast removes and returns the user's most recently logged expense.
	DeleteLast(ctx context.Context, userID int64) (*dao.Expense, error)
	// Total returns the sum of the user's expenses in [from, to).
	Total(ctx context.Context, userID int64, from, to time.Time) (float64, error)
	// SummaryByCategory returns per-category totals in [from, to), highest first.
	SummaryByCategory(ctx context.Context, userID int64, from, to time.Time) ([]entity.CategoryTotal, error)
}

type expenseRepository struct {
	db *gorm.DB
}

// NewExpenseRepository constructs the GORM expense repository.
func NewExpenseRepository(db *gorm.DB) IExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(ctx context.Context, e *dao.Expense) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("create expense: %w", err)
	}
	return nil
}

func (r *expenseRepository) List(ctx context.Context, userID int64, from, to time.Time, category string) ([]dao.Expense, error) {
	q := r.db.WithContext(ctx).
		Where("user_id = ? AND expense_date >= ? AND expense_date < ?", userID, from, to)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var out []dao.Expense
	if err := q.Order("expense_date DESC, id DESC").Find(&out).Error; err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	return out, nil
}

func (r *expenseRepository) DeleteLast(ctx context.Context, userID int64) (*dao.Expense, error) {
	var e dao.Expense
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("find last expense: %w", err)
	}
	if err := r.db.WithContext(ctx).Delete(&dao.Expense{}, e.ID).Error; err != nil {
		return nil, fmt.Errorf("delete last expense: %w", err)
	}
	return &e, nil
}

func (r *expenseRepository) Total(ctx context.Context, userID int64, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&dao.Expense{}).
		Where("user_id = ? AND expense_date >= ? AND expense_date < ?", userID, from, to).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("sum expenses: %w", err)
	}
	return total, nil
}

func (r *expenseRepository) SummaryByCategory(ctx context.Context, userID int64, from, to time.Time) ([]entity.CategoryTotal, error) {
	var rows []entity.CategoryTotal
	err := r.db.WithContext(ctx).Model(&dao.Expense{}).
		Select("category, COALESCE(SUM(amount), 0) AS total").
		Where("user_id = ? AND expense_date >= ? AND expense_date < ?", userID, from, to).
		Group("category").
		Order("total DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("summary by category: %w", err)
	}
	return rows, nil
}
