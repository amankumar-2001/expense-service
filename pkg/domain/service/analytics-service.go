package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/kharchibook/expense-service/enums/autopaystatus"
	"github.com/kharchibook/expense-service/pkg/domain/dto/response"
	"github.com/kharchibook/expense-service/pkg/infrastructure/cacherepo"
	"github.com/kharchibook/expense-service/pkg/infrastructure/sqlrepo"
	"github.com/kharchibook/expense-service/third_party/platlogger"
	"github.com/kharchibook/expense-service/utils"
)

// IAnalyticsService computes the dashboard's derived figures: committed/free
// money and upcoming deductions.
type IAnalyticsService interface {
	Committed(ctx context.Context, userID int64) (*response.CommittedSummaryResponse, error)
	Upcoming(ctx context.Context, userID int64, days int) ([]response.UpcomingDeductionResponse, error)
}

type analyticsService struct {
	autopays sqlrepo.IAutoPayRepository
	finance  IFinanceService
	cache    cacherepo.ISummaryCache
	cacheTTL time.Duration
}

// NewAnalyticsService constructs the analytics service.
func NewAnalyticsService(
	autopays sqlrepo.IAutoPayRepository,
	finance IFinanceService,
	cache cacherepo.ISummaryCache,
	cacheTTL time.Duration,
) IAnalyticsService {
	return &analyticsService{autopays: autopays, finance: finance, cache: cache, cacheTTL: cacheTTL}
}

func (s *analyticsService) Committed(ctx context.Context, userID int64) (*response.CommittedSummaryResponse, error) {
	// Read-through cache: the dashboard hits this on every load.
	if data, hit, err := s.cache.GetCommitted(ctx, userID); err != nil {
		platlogger.WithContext(ctx).Warn("read committed cache", "error", err)
	} else if hit {
		var cached response.CommittedSummaryResponse
		if json.Unmarshal(data, &cached) == nil {
			return &cached, nil
		}
	}

	salary, err := s.finance.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.autopays.List(ctx, userID, autopaystatus.Active.String(), "")
	if err != nil {
		return nil, err
	}

	var totalCommitted float64
	autopays := make([]response.AutoPayResponse, 0, len(rows))
	for i := range rows {
		totalCommitted += rows[i].Amount
		autopays = append(autopays, toAutoPayResponse(&rows[i]))
	}

	out := &response.CommittedSummaryResponse{
		MonthlySalary:  salary.MonthlySalary,
		SalaryDay:      salary.SalaryDay,
		TotalCommitted: totalCommitted,
		FreeMoney:      salary.MonthlySalary - totalCommitted,
		Autopays:       autopays,
	}

	if data, err := json.Marshal(out); err == nil {
		if err := s.cache.SetCommitted(ctx, userID, data, s.cacheTTL); err != nil {
			platlogger.WithContext(ctx).Warn("write committed cache", "error", err)
		}
	}
	return out, nil
}

func (s *analyticsService) Upcoming(ctx context.Context, userID int64, days int) ([]response.UpcomingDeductionResponse, error) {
	rows, err := s.autopays.List(ctx, userID, autopaystatus.Active.String(), "")
	if err != nil {
		return nil, err
	}

	now := time.Now()
	out := make([]response.UpcomingDeductionResponse, 0, len(rows))
	for i := range rows {
		inDays := utils.NextDeductionInDays(now, rows[i].DeductDay)
		if inDays > days {
			continue
		}
		out = append(out, response.UpcomingDeductionResponse{
			AutopayID: strconv.FormatInt(rows[i].ID, 10),
			Name:      rows[i].Name,
			Amount:    rows[i].Amount,
			DeductDay: rows[i].DeductDay,
			InDays:    inDays,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].InDays < out[j].InDays })
	return out, nil
}
