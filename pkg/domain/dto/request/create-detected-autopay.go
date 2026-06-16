package request

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// CreateDetectedAutoPayRequest is the body of POST /v1/internal/autopays — a
// recurring commitment detected from the user's mailbox by the mcp-gateway.
// Detected entries get source="email_auto" and status="inactive": they await the
// user's confirmation before counting as committed money. ConfidenceScore (0..1)
// records how strong the detection signal was.
type CreateDetectedAutoPayRequest struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Amount          float64  `json:"amount"`
	DeductDay       int      `json:"deductDay"`
	Notes           string   `json:"notes"`
	ConfidenceScore *float64 `json:"confidenceScore"`
}

// Validate checks the POST /v1/internal/autopays body. It mirrors the manual
// create rules but also bounds the confidence score to [0, 1].
func (r CreateDetectedAutoPayRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 120)),
		validation.Field(&r.Type, validation.Required, validation.By(validType)),
		validation.Field(&r.Amount, validation.Required, validation.Min(0.0).Exclusive()),
		validation.Field(&r.DeductDay, validation.Required, validation.Min(1), validation.Max(31)),
		validation.Field(&r.ConfidenceScore, validation.Min(0.0), validation.Max(1.0)),
	)
}
