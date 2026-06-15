// Package whatsappclient sends outbound WhatsApp messages via the Meta Cloud API.
package whatsappclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kharchibook/expense-service/config"
)

// ISender sends a plain-text WhatsApp reply to a recipient phone (E.164 or wa_id).
type ISender interface {
	Send(ctx context.Context, toPhone, text string) error
}

type client struct {
	endpoint    string
	accessToken string
	http        *http.Client
}

// New constructs the Meta Cloud API sender from WhatsApp config.
func New(cfg config.WhatsApp) ISender {
	endpoint := fmt.Sprintf("%s/%s/%s/messages",
		cfg.GraphBaseURL, cfg.GraphVersion, cfg.PhoneNumberID)
	return &client{
		endpoint:    endpoint,
		accessToken: cfg.AccessToken,
		http:        &http.Client{Timeout: 10 * time.Second},
	}
}

type textBody struct {
	Body string `json:"body"`
}

type sendRequest struct {
	MessagingProduct string   `json:"messaging_product"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Text             textBody `json:"text"`
}

func (c *client) Send(ctx context.Context, toPhone, text string) error {
	payload, err := json.Marshal(sendRequest{
		MessagingProduct: "whatsapp",
		To:               toPhone,
		Type:             "text",
		Text:             textBody{Body: text},
	})
	if err != nil {
		return fmt.Errorf("marshal send request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build send request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send to graph api: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("graph api returned %d: %s", res.StatusCode, string(body))
	}
	return nil
}
