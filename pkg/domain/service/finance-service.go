package service

import (
	"context"

	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/models/dao"
	"github.com/kharchibook/expense-service/pkg/infrastructure/cacherepo"
	"github.com/kharchibook/expense-service/pkg/infrastructure/sqlrepo"
	"github.com/kharchibook/expense-service/third_party/platlogger"
)

// Salary is the resolved per-user salary inputs. When a user has not set a
// salary, SalaryDay defaults to 1 and MonthlySalary is 0.
type Salary struct {
	MonthlySalary float64
	SalaryDay     int
}

// IFinanceService owns the per-user salary inputs to the free-money calculation.
type IFinanceService interface {
	Get(ctx context.Context, userID int64) (Salary, error)
	Set(ctx context.Context, userID int64, amount float64, salaryDay int) (Salary, error)
}

type financeService struct {
	repo  sqlrepo.IFinanceRepository
	cache cacherepo.ISummaryCache
}

// NewFinanceService constructs the finance service.
func NewFinanceService(repo sqlrepo.IFinanceRepository, cache cacherepo.ISummaryCache) IFinanceService {
	return &financeService{repo: repo, cache: cache}
}

func (s *financeService) Get(ctx context.Context, userID int64) (Salary, error) {
	f, err := s.repo.Get(ctx, userID)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrNotFound) {
			return Salary{MonthlySalary: 0, SalaryDay: 1}, nil
		}
		return Salary{}, err
	}
	return Salary{MonthlySalary: f.MonthlySalary, SalaryDay: f.SalaryDay}, nil
}

func (s *financeService) Set(ctx context.Context, userID int64, amount float64, salaryDay int) (Salary, error) {
	f := &dao.UserFinance{
		UserID:        userID,
		MonthlySalary: amount,
		SalaryDay:     salaryDay,
	}
	if err := s.repo.Upsert(ctx, f); err != nil {
		return Salary{}, err
	}
	if err := s.cache.Invalidate(ctx, userID); err != nil {
		platlogger.WithContext(ctx).Warn("invalidate committed cache", "error", err)
	}
	return Salary{MonthlySalary: amount, SalaryDay: salaryDay}, nil
}
