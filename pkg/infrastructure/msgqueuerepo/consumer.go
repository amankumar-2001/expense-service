package msgqueuerepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kharchibook/expense-service/pkg/domain/dto/message"
	"github.com/kharchibook/expense-service/third_party/platlogger"
	"github.com/twmb/franz-go/pkg/kgo"
)

// InboundHandler processes one inbound message. The handler owns retry + DLQ
// policy; the consumer commits the record once the handler returns, so a slow or
// failing message never blocks the partition (at-least-once + idempotency key).
type InboundHandler func(ctx context.Context, m message.WhatsAppInbound) error

// Consumer reads the inbound topic via a franz-go consumer group.
type Consumer struct {
	client *kgo.Client
}

// NewKafkaConsumer builds a consumer-group client for the inbound topic.
func NewKafkaConsumer(brokers []string, group, topic string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		// Commit explicitly after each batch is handled.
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("new kafka consumer: %w", err)
	}
	return &Consumer{client: client}, nil
}

// Run polls until ctx is cancelled, dispatching each record to handle.
func (c *Consumer) Run(ctx context.Context, handle InboundHandler) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		fetches := c.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			// Context cancellation surfaces here on shutdown — return cleanly.
			if ctx.Err() != nil {
				return ctx.Err()
			}
			for _, e := range errs {
				platlogger.WithContext(ctx).Error("kafka fetch error", "topic", e.Topic, "error", e.Err)
			}
			continue
		}
		fetches.EachRecord(func(r *kgo.Record) {
			var m message.WhatsAppInbound
			if err := json.Unmarshal(r.Value, &m); err != nil {
				platlogger.WithContext(ctx).Error("drop unparseable record", "error", err)
				return
			}
			if err := handle(ctx, m); err != nil {
				// Handler already exhausted retries / DLQ'd — log and move on so the
				// offset still commits.
				platlogger.WithContext(ctx).Error("inbound handler failed", "msgId", m.MsgID, "error", err)
			}
		})
		if err := c.client.CommitUncommittedOffsets(ctx); err != nil && !errors.Is(err, context.Canceled) {
			platlogger.WithContext(ctx).Error("commit offsets failed", "error", err)
		}
	}
}

// Close releases the consumer.
func (c *Consumer) Close() error {
	c.client.Close()
	return nil
}
