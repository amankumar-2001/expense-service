// Package message holds the event payloads exchanged over the message queue
// (Kafka). One file per event type.
package message

// WhatsAppInbound is one inbound WhatsApp text message, published by the webhook
// ingress to the `whatsapp.inbound` topic and consumed by the WhatsApp worker.
// It carries only what the worker needs; identity resolution happens downstream.
type WhatsAppInbound struct {
	// WaID is the sender's WhatsApp id (phone digits, no '+'), e.g. "919876543210".
	WaID string `json:"waId"`
	// Text is the message body the user typed.
	Text string `json:"text"`
	// MsgID is Meta's message id — the idempotency key (Meta + Kafka both retry).
	MsgID string `json:"msgId"`
	// Timestamp is Meta's unix-seconds send time, as a string (verbatim from Meta).
	Timestamp string `json:"timestamp"`
}
