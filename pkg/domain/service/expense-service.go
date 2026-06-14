package service

import (
	"context"
	"strconv"
	"time"

	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/dto/request"
	"github.com/kharchibook/expense-service/pkg/domain/dto/response"
	"github.com/kharchibook/expense-service/pkg/domain/models/dao"
	"github.com/kharchibook/expense-service/pkg/infrastructure/sqlrepo"
	"github.com/kharchibook/expense-service/utils"
)

// IExpenseService owns expense logging and the monthly spend summary.
type IExpenseService interface {
	Create(ctx context.Context, userID int64, req request.CreateExpenseRequest) (*response.ExpenseResponse, error)
	List(ctx context.Context, userID int64, month, category string) ([]response.ExpenseResponse, error)
	DeleteLast(ctx context.Context, userID int64) (*response.ExpenseResponse, error)
	Summary(ctx context.Context, userID int64, month string) (*response.ExpenseSummaryResponse, error)
}

type expenseService struct {
	repo sqlrepo.IExpenseRepository
}

// NewExpenseService constructs the expense service.
func NewExpenseService(repo sqlrepo.IExpenseRepository) IExpenseService {
	return &expenseService{repo: repo}
}

func (s *expenseService) Create(ctx context.Context, userID int64, req request.CreateExpenseRequest) (*response.ExpenseResponse, error) {
	date := time.Now()
	if req.ExpenseDate != "" {
		parsed, err := time.ParseInLocation("2006-01-02", req.ExpenseDate, time.Local)
		if err != nil {
			return nil, apperrors.BadRequestError("expenseDate must be YYYY-MM-DD")
		}
		date = parsed
	}

	e := &dao.Expense{
		UserID:      userID,
		Amount:      req.Amount,
		Category:    req.Category,
		Note:        req.Note,
		RawText:     req.RawText,
		ExpenseDate: date,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	out := toExpenseResponse(e)
	return &out, nil
}

func (s *expenseService) List(ctx context.Context, userID int64, month, category string) ([]response.ExpenseResponse, error) {
	from, err := utils.ParseMonth(month, time.Now(), time.Local)
	if err != nil {
		return nil, apperrors.BadRequestError("month must be YYYY-MM")
	}
	to := from.AddDate(0, 1, 0)

	rows, err := s.repo.List(ctx, userID, from, to, category)
	if err != nil {
		return nil, err
	}
	out := make([]response.ExpenseResponse, 0, len(rows))
	for i := range rows {
		out = append(out, toExpenseResponse(&rows[i]))
	}
	return out, nil
}

func (s *expenseService) DeleteLast(ctx context.Context, userID int64) (*response.ExpenseResponse, error) {
	e, err := s.repo.DeleteLast(ctx, userID)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.NotFoundError("no expense to delete")
		}
		return nil, err
	}
	out := toExpenseResponse(e)
	return &out, nil
}

func (s *expenseService) Summary(ctx context.Context, userID int64, month string) (*response.ExpenseSummaryResponse, error) {
	from, err := utils.ParseMonth(month, time.Now(), time.Local)
	if err != nil {
		return nil, apperrors.BadRequestError("month must be YYYY-MM")
	}
	to := from.AddDate(0, 1, 0)

	total, err := s.repo.Total(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	cats, err := s.repo.SummaryByCategory(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	byCategory := make([]response.CategoryTotalResponse, 0, len(cats))
	for _, c := range cats {
		byCategory = append(byCategory, response.CategoryTotalResponse{Category: c.Category, Total: c.Total})
	}

	return &response.ExpenseSummaryResponse{
		Month:      utils.MonthLabel(from),
		Total:      total,
		ByCategory: byCategory,
	}, nil
}

func toExpenseResponse(e *dao.Expense) response.ExpenseResponse {
	return response.ExpenseResponse{
		ID:          strconv.FormatInt(e.ID, 10),
		Amount:      e.Amount,
		Category:    e.Category,
		Note:        e.Note,
		RawText:     e.RawText,
		ExpenseDate: e.ExpenseDate.Format("2006-01-02"),
	}
}
