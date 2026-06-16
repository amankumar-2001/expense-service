package service

import (
	"context"
	"strconv"

	"github.com/kharchibook/expense-service/enums/autopaysource"
	"github.com/kharchibook/expense-service/enums/autopaystatus"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/dto/request"
	"github.com/kharchibook/expense-service/pkg/domain/dto/response"
	"github.com/kharchibook/expense-service/pkg/domain/models/dao"
	"github.com/kharchibook/expense-service/pkg/infrastructure/cacherepo"
	"github.com/kharchibook/expense-service/pkg/infrastructure/sqlrepo"
	"github.com/kharchibook/expense-service/third_party/platlogger"
)

// IAutoPayService owns CRUD + confirmation of recurring commitments. Any change
// invalidates the cached committed-money summary for that user.
type IAutoPayService interface {
	Create(ctx context.Context, userID int64, req request.CreateAutoPayRequest) (*response.AutoPayResponse, error)
	// CreateDetected stores a mailbox-detected commitment as a pending
	// (source="email_auto", status="inactive") entry awaiting user confirmation.
	CreateDetected(ctx context.Context, userID int64, req request.CreateDetectedAutoPayRequest) (*response.AutoPayResponse, error)
	List(ctx context.Context, userID int64, status, typ string) ([]response.AutoPayResponse, error)
	Update(ctx context.Context, userID, id int64, req request.UpdateAutoPayRequest) (*response.AutoPayResponse, error)
	Delete(ctx context.Context, userID, id int64) error
	Confirm(ctx context.Context, userID, id int64) (*response.AutoPayResponse, error)
}

type autopayService struct {
	repo  sqlrepo.IAutoPayRepository
	cache cacherepo.ISummaryCache
}

// NewAutoPayService constructs the autopay service.
func NewAutoPayService(repo sqlrepo.IAutoPayRepository, cache cacherepo.ISummaryCache) IAutoPayService {
	return &autopayService{repo: repo, cache: cache}
}

func (s *autopayService) Create(ctx context.Context, userID int64, req request.CreateAutoPayRequest) (*response.AutoPayResponse, error) {
	a := &dao.AutoPay{
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		Amount:    req.Amount,
		DeductDay: req.DeductDay,
		Source:    autopaysource.Manual.String(),
		Status:    autopaystatus.Active.String(),
		Notes:     req.Notes,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	s.invalidate(ctx, userID)
	out := toAutoPayResponse(a)
	return &out, nil
}

func (s *autopayService) CreateDetected(ctx context.Context, userID int64, req request.CreateDetectedAutoPayRequest) (*response.AutoPayResponse, error) {
	a := &dao.AutoPay{
		UserID:          userID,
		Name:            req.Name,
		Type:            req.Type,
		Amount:          req.Amount,
		DeductDay:       req.DeductDay,
		Source:          autopaysource.EmailAuto.String(),
		Status:          autopaystatus.Inactive.String(),
		ConfidenceScore: req.ConfidenceScore,
		Notes:           req.Notes,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	// A pending (inactive) detection doesn't change committed money, so no cache
	// invalidation is needed until it's confirmed.
	out := toAutoPayResponse(a)
	return &out, nil
}

func (s *autopayService) List(ctx context.Context, userID int64, status, typ string) ([]response.AutoPayResponse, error) {
	rows, err := s.repo.List(ctx, userID, status, typ)
	if err != nil {
		return nil, err
	}
	out := make([]response.AutoPayResponse, 0, len(rows))
	for i := range rows {
		out = append(out, toAutoPayResponse(&rows[i]))
	}
	return out, nil
}

func (s *autopayService) Update(ctx context.Context, userID, id int64, req request.UpdateAutoPayRequest) (*response.AutoPayResponse, error) {
	fields := map[string]any{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if req.Amount != nil {
		fields["amount"] = *req.Amount
	}
	if req.DeductDay != nil {
		fields["deduct_day"] = *req.DeductDay
	}
	if req.Notes != nil {
		fields["notes"] = *req.Notes
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if len(fields) > 0 {
		if err := s.repo.UpdateFields(ctx, userID, id, fields); err != nil {
			return nil, mapAutoPayErr(err)
		}
		s.invalidate(ctx, userID)
	}

	a, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, mapAutoPayErr(err)
	}
	out := toAutoPayResponse(a)
	return &out, nil
}

func (s *autopayService) Delete(ctx context.Context, userID, id int64) error {
	// Soft-delete: keep the row for history, mark it cancelled.
	if err := s.repo.UpdateFields(ctx, userID, id, map[string]any{
		"status": autopaystatus.Cancelled.String(),
	}); err != nil {
		return mapAutoPayErr(err)
	}
	s.invalidate(ctx, userID)
	return nil
}

func (s *autopayService) Confirm(ctx context.Context, userID, id int64) (*response.AutoPayResponse, error) {
	// Confirming an auto-detected entry activates it and clears the (now moot)
	// confidence score.
	if err := s.repo.UpdateFields(ctx, userID, id, map[string]any{
		"status":           autopaystatus.Active.String(),
		"confidence_score": nil,
	}); err != nil {
		return nil, mapAutoPayErr(err)
	}
	s.invalidate(ctx, userID)

	a, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, mapAutoPayErr(err)
	}
	out := toAutoPayResponse(a)
	return &out, nil
}

// mapAutoPayErr maps the repository not-found sentinel to a 404.
func mapAutoPayErr(err error) error {
	if apperrors.Is(err, apperrors.ErrNotFound) {
		return apperrors.NotFoundError("autopay not found")
	}
	return err
}

func (s *autopayService) invalidate(ctx context.Context, userID int64) {
	if err := s.cache.Invalidate(ctx, userID); err != nil {
		platlogger.WithContext(ctx).Warn("invalidate committed cache", "error", err)
	}
}

func toAutoPayResponse(a *dao.AutoPay) response.AutoPayResponse {
	return response.AutoPayResponse{
		ID:              strconv.FormatInt(a.ID, 10),
		Name:            a.Name,
		Type:            a.Type,
		Amount:          a.Amount,
		DeductDay:       a.DeductDay,
		Source:          a.Source,
		Status:          a.Status,
		ConfidenceScore: a.ConfidenceScore,
	}
}
