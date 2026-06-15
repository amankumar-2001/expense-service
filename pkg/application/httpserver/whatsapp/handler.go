// Package whatsapp holds the public WhatsApp webhook HTTP handlers. The ingress
// is deliberately thin: it verifies Meta's request (GET handshake / POST HMAC
// signature) and publishes inbound messages to Kafka, returning 200 fast. All
// business logic runs in the separate worker.
package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kharchibook/expense-service/constants"
	"github.com/kharchibook/expense-service/pkg/di"
	"github.com/kharchibook/expense-service/pkg/domain/dto/message"
	"github.com/kharchibook/expense-service/third_party/platlogger"
)

// Handler serves /whatsapp/webhook (Meta Cloud API).
type Handler struct {
	app di.AppInterface
}

// NewHandler constructs the webhook handler.
func NewHandler(app di.AppInterface) *Handler {
	return &Handler{app: app}
}

// Routes mounts the webhook routes. These are PUBLIC (no JWT) — authenticated by
// the verify token (GET) and the X-Hub-Signature-256 HMAC (POST).
func (h *Handler) Routes(r gin.IRouter) {
	r.GET("/whatsapp/webhook", h.Verify)
	r.POST("/whatsapp/webhook", h.Receive)
}

// Verify handles Meta's one-time subscription handshake: echo hub.challenge when
// the verify token matches.
func (h *Handler) Verify(c *gin.Context) {
	cfg := h.app.Config().WhatsApp
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token != "" &&
		subtle.ConstantTimeCompare([]byte(token), []byte(cfg.VerifyToken)) == 1 {
		c.String(http.StatusOK, challenge)
		return
	}
	c.Status(http.StatusForbidden)
}

// cloudAPIPayload is the subset of the Meta Cloud API webhook envelope we read.
type cloudAPIPayload struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

// Receive verifies the payload signature and publishes each text message to Kafka.
// It always returns 200 on a well-formed, authentic request so Meta doesn't retry.
func (h *Handler) Receive(c *gin.Context) {
	cfg := h.app.Config().WhatsApp

	body, err := c.GetRawData()
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if !validSignature(cfg.AppSecret, c.GetHeader(constants.HeaderHubSignature), body) {
		platlogger.WithContext(c.Request.Context()).Warn("whatsapp webhook signature rejected")
		c.Status(http.StatusForbidden)
		return
	}

	var payload cloudAPIPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		// Malformed but authentic — ack so Meta stops retrying.
		c.Status(http.StatusOK)
		return
	}

	ctx := c.Request.Context()
	pub := h.app.InboundPublisher()
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if msg.Type != "text" || msg.Text.Body == "" {
					continue
				}
				m := message.WhatsAppInbound{
					WaID:      msg.From,
					Text:      msg.Text.Body,
					MsgID:     msg.ID,
					Timestamp: msg.Timestamp,
				}
				if err := pub.PublishInbound(ctx, m); err != nil {
					platlogger.WithContext(ctx).Error("publish inbound failed", "msgId", m.MsgID, "error", err)
				}
			}
		}
	}
	c.Status(http.StatusOK)
}

// validSignature verifies Meta's X-Hub-Signature-256 ("sha256=<hex>") over the
// raw body. An empty configured secret disables verification (local dev only).
func validSignature(appSecret, header string, body []byte) bool {
	if appSecret == "" {
		return true
	}
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	want, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	return hmac.Equal(want, mac.Sum(nil))
}
