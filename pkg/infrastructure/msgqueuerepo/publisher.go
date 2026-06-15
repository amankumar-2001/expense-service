// Package msgqueuerepo publishes and consumes the WhatsApp inbound stream over
// Kafka (franz-go). It ships a logging stub used when the queue is disabled, so
// local dev runs with no broker; the real producer/consumer drop in behind the
// same interfaces without touching callers.
package msgqueuerepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kharchibook/expense-service/pkg/domain/dto/message"
	"github.com/kharchibook/expense-service/third_party/platlogger"
	"github.com/twmb/franz-go/pkg/kgo"
)

// IInboundPublisher publishes inbound WhatsApp messages for the worker to consume.
type IInboundPublisher interface {
	// PublishInbound enqueues one message. Keyed by WaID for per-user ordering.
	PublishInbound(ctx context.Context, m message.WhatsAppInbound) error
	// PublishDLQ routes a message that exhausted its retries to the dead-letter
	// topic, annotated with the failure reason.
	PublishDLQ(ctx context.Context, m message.WhatsAppInbound, reason string) error
	// Close flushes and releases the producer.
	Close() error
}

// kafkaPublisher is the franz-go backed producer.
type kafkaPublisher struct {
	client       *kgo.Client
	inboundTopic string
	dlqTopic     string
}

// NewKafkaPublisher builds a franz-go producer against the given brokers.
func NewKafkaPublisher(brokers []string, inboundTopic, dlqTopic string) (IInboundPublisher, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("new kafka client: %w", err)
	}
	return &kafkaPublisher{client: client, inboundTopic: inboundTopic, dlqTopic: dlqTopic}, nil
}

func (p *kafkaPublisher) PublishInbound(ctx context.Context, m message.WhatsAppInbound) error {
	return p.produce(ctx, p.inboundTopic, []byte(m.WaID), m)
}

func (p *kafkaPublisher) PublishDLQ(ctx context.Context, m message.WhatsAppInbound, reason string) error {
	envelope := struct {
		message.WhatsAppInbound
		Reason string `json:"reason"`
	}{WhatsAppInbound: m, Reason: reason}
	return p.produce(ctx, p.dlqTopic, []byte(m.WaID), envelope)
}

func (p *kafkaPublisher) produce(ctx context.Context, topic string, key []byte, v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	// Synchronous produce so the HTTP ingress only 200s once the message is durably
	// accepted by the broker.
	res := p.client.ProduceSync(ctx, &kgo.Record{Topic: topic, Key: key, Value: payload})
	return res.FirstErr()
}

func (p *kafkaPublisher) Close() error {
	p.client.Close()
	return nil
}

// stubPublisher logs instead of producing — used when MsgQueue.Enabled is false.
type stubPublisher struct{}

// NewStubPublisher returns the logging stub publisher.
func NewStubPublisher() IInboundPublisher { return &stubPublisher{} }

func (s *stubPublisher) PublishInbound(ctx context.Context, m message.WhatsAppInbound) error {
	platlogger.WithContext(ctx).Info("whatsapp inbound published (stub)", "msgId", m.MsgID, "waId", mask(m.WaID))
	return nil
}

func (s *stubPublisher) PublishDLQ(ctx context.Context, m message.WhatsAppInbound, reason string) error {
	platlogger.WithContext(ctx).Warn("whatsapp inbound → DLQ (stub)", "msgId", m.MsgID, "reason", reason)
	return nil
}

func (s *stubPublisher) Close() error { return nil }

// mask trims a phone/wa_id to its last 4 digits for logs.
func mask(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}
