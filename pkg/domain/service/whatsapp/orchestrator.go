// Package whatsapp orchestrates the WhatsApp worker pipeline: it resolves the
// sender to a user, parses the command, calls the existing domain services, and
// sends a formatted reply. It is invoked once per Kafka message by the worker.
package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kharchibook/expense-service/pkg/domain/dto/message"
	"github.com/kharchibook/expense-service/pkg/domain/dto/request"
	"github.com/kharchibook/expense-service/pkg/domain/parser"
	"github.com/kharchibook/expense-service/pkg/domain/service"
	"github.com/kharchibook/expense-service/pkg/infrastructure/authclient"
	"github.com/kharchibook/expense-service/pkg/infrastructure/msgqueuerepo"
	"github.com/kharchibook/expense-service/pkg/infrastructure/whatsappclient"
	"github.com/kharchibook/expense-service/third_party/platlogger"
)

const (
	maxAttempts    = 3
	retryBackoff   = 500 * time.Millisecond
	idempotencyTTL = 24 * time.Hour
)

// Deps are the orchestrator's collaborators.
type Deps struct {
	Auth      authclient.IClient
	Sender    whatsappclient.ISender
	Publisher msgqueuerepo.IInboundPublisher
	Redis     *redis.Client
	Expenses  service.IExpenseService
	Autopays  service.IAutoPayService
	Analytics service.IAnalyticsService
	Finance   service.IFinanceService
	SignupURL string
}

// Service is the WhatsApp message orchestrator.
type Service struct {
	deps Deps
	now  func() time.Time
}

// New constructs the orchestrator.
func New(deps Deps) *Service {
	return &Service{deps: deps, now: time.Now}
}

// Handle is the Kafka InboundHandler. It is idempotent (skips already-processed
// message ids), retries transient failures up to maxAttempts, and dead-letters on
// exhaustion. It returns nil on permanent failures (they produce a user reply) so
// the consumer commits the offset.
func (s *Service) Handle(ctx context.Context, m message.WhatsAppInbound) error {
	key := "expense:wa:msg:" + m.MsgID
	if seen, _ := s.deps.Redis.Exists(ctx, key).Result(); seen > 0 {
		platlogger.WithContext(ctx).Info("skip duplicate whatsapp message", "msgId", m.MsgID)
		return nil
	}

	reply, transientErr := s.buildReplyWithRetry(ctx, m)
	if transientErr != nil {
		// Exhausted transient retries → dead-letter, mark seen, commit.
		if err := s.deps.Publisher.PublishDLQ(ctx, m, transientErr.Error()); err != nil {
			platlogger.WithContext(ctx).Error("DLQ publish failed", "msgId", m.MsgID, "error", err)
		}
		s.markSeen(ctx, key)
		return transientErr
	}

	if reply != "" {
		if err := s.sendWithRetry(ctx, m.WaID, reply); err != nil {
			if dlqErr := s.deps.Publisher.PublishDLQ(ctx, m, "send failed: "+err.Error()); dlqErr != nil {
				platlogger.WithContext(ctx).Error("DLQ publish failed", "msgId", m.MsgID, "error", dlqErr)
			}
		}
	}
	s.markSeen(ctx, key)
	return nil
}

func (s *Service) markSeen(ctx context.Context, key string) {
	if err := s.deps.Redis.Set(ctx, key, "1", idempotencyTTL).Err(); err != nil {
		platlogger.WithContext(ctx).Warn("failed to mark message seen", "error", err)
	}
}

func (s *Service) buildReplyWithRetry(ctx context.Context, m message.WhatsAppInbound) (string, error) {
	var reply string
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		r, transient, err := s.buildReply(ctx, m)
		if err == nil {
			return r, nil
		}
		if !transient {
			// Permanent failure: surface the friendly reply, stop retrying.
			return r, nil
		}
		lastErr = err
		platlogger.WithContext(ctx).Warn("transient failure, retrying",
			"msgId", m.MsgID, "attempt", attempt, "error", err)
		if attempt < maxAttempts {
			time.Sleep(retryBackoff)
		}
		reply = r
	}
	_ = reply
	return "", lastErr
}

func (s *Service) sendWithRetry(ctx context.Context, to, text string) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := s.deps.Sender.Send(ctx, to, text); err != nil {
			lastErr = err
			if attempt < maxAttempts {
				time.Sleep(retryBackoff)
			}
			continue
		}
		return nil
	}
	return lastErr
}

// buildReply resolves identity, parses, and executes the command. It returns the
// reply text, whether a returned error is permanent (true) or transient (false),
// and the error. Reads (summary/committed/upcoming) treat failures as transient
// (safe to retry); writes (expense/autopay/salary) treat failures as permanent
// with an apology reply, so a retry never double-writes.
func (s *Service) buildReply(ctx context.Context, m message.WhatsAppInbound) (reply string, permanent bool, err error) {
	user, err := s.deps.Auth.UserByPhone(ctx, m.WaID)
	if err != nil {
		if errors.Is(err, authclient.ErrNotRegistered) {
			return s.notRegisteredReply(), true, nil
		}
		return "", false, fmt.Errorf("resolve identity: %w", err) // transient
	}

	cmd := parser.Parse(m.Text)
	uid := user.UserID

	switch cmd.Kind {
	case parser.KindGreeting:
		greeting := "👋 Namaste"
		if user.Name != "" {
			greeting += " " + user.Name
		}
		return greeting + "!\n\n" + formatHelp(), true, nil
	case parser.KindHelp:
		return formatHelp(), true, nil

	case parser.KindExpense:
		out, e := s.deps.Expenses.Create(ctx, uid, request.CreateExpenseRequest{
			Amount:   cmd.Amount,
			Category: cmd.Category.String(),
			Note:     cmd.Note,
			RawText:  m.Text,
		})
		if e != nil {
			return "⚠️ Couldn't log that expense. Please try again.", true, nil
		}
		return formatExpenseConfirmation(out), true, nil

	case parser.KindSummary:
		out, e := s.deps.Expenses.Summary(ctx, uid, s.monthQuery(cmd.Month))
		if e != nil {
			return "", false, fmt.Errorf("summary: %w", e) // read → transient
		}
		return formatSummary(out), true, nil

	case parser.KindAddAutoPay:
		out, e := s.deps.Autopays.Create(ctx, uid, request.CreateAutoPayRequest{
			Name:      cmd.Name,
			Type:      cmd.Type.String(),
			Amount:    cmd.Amount,
			DeductDay: cmd.DeductDay,
		})
		if e != nil {
			return "⚠️ Couldn't add that. Try: `add emi home loan 22000 on 5th`", true, nil
		}
		return formatAutoPayAdded(out), true, nil

	case parser.KindAutoPayList, parser.KindCommitted:
		out, e := s.deps.Analytics.Committed(ctx, uid)
		if e != nil {
			return "", false, fmt.Errorf("committed: %w", e) // read → transient
		}
		return formatCommitted(out), true, nil

	case parser.KindUpcoming:
		out, e := s.deps.Analytics.Upcoming(ctx, uid, 7)
		if e != nil {
			return "", false, fmt.Errorf("upcoming: %w", e) // read → transient
		}
		return formatUpcoming(out), true, nil

	case parser.KindRemove:
		return s.removeAutoPay(ctx, uid, cmd.Name), true, nil

	case parser.KindSalary:
		if cmd.Amount <= 0 || cmd.SalaryDay < 1 || cmd.SalaryDay > 31 {
			return "⚠️ Try: `salary 65000 on 1st`", true, nil
		}
		if _, e := s.deps.Finance.Set(ctx, uid, cmd.Amount, cmd.SalaryDay); e != nil {
			return "⚠️ Couldn't save your salary. Please try again.", true, nil
		}
		return formatSalarySet(cmd.Amount, cmd.SalaryDay), true, nil

	default:
		return "🤔 I didn't understand that. Type *help* to see what I can do.", true, nil
	}
}

// removeAutoPay finds an autopay by (case-insensitive) name and deletes it. There
// is no delete-by-name service method, so we resolve the id here.
func (s *Service) removeAutoPay(ctx context.Context, uid int64, name string) string {
	if strings.TrimSpace(name) == "" {
		return "⚠️ Tell me which one to remove, e.g. `remove netflix`."
	}
	list, err := s.deps.Autopays.List(ctx, uid, "", "")
	if err != nil {
		return "⚠️ Couldn't reach your commitments. Please try again."
	}
	target := strings.ToLower(strings.TrimSpace(name))
	for _, a := range list {
		if strings.Contains(strings.ToLower(a.Name), target) {
			id, perr := strconv.ParseInt(a.ID, 10, 64)
			if perr != nil {
				continue
			}
			if err := s.deps.Autopays.Delete(ctx, uid, id); err != nil {
				return "⚠️ Couldn't remove " + a.Name + ". Please try again."
			}
			return "🗑️ Removed " + a.Name + "."
		}
	}
	return fmt.Sprintf("🤷 I couldn't find an autopay matching %q. Type *autopay list* to see them.", name)
}

func (s *Service) notRegisteredReply() string {
	return fmt.Sprintf("This number isn't registered with KharchiBook yet. Sign up at %s, then message me from your registered phone.", s.deps.SignupURL)
}

// monthQuery converts a free-text month token ("june", "Jun") to the "2006-01"
// form ExpenseService.Summary expects. Empty input → "" (current month). A month
// later than the current one is assumed to mean last year.
func (s *Service) monthQuery(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	now := s.now()
	for _, layout := range []string{"January", "Jan"} {
		if t, err := time.Parse(layout, strings.Title(strings.ToLower(token))); err == nil {
			year := now.Year()
			if int(t.Month()) > int(now.Month()) {
				year--
			}
			return fmt.Sprintf("%04d-%02d", year, int(t.Month()))
		}
	}
	// Already "2006-01"? Pass through; otherwise fall back to current month.
	if _, err := time.Parse("2006-01", token); err == nil {
		return token
	}
	return ""
}
