package request

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kharchibook/expense-service/enums/autopaytype"
)

// CreateAutoPayRequest is the POST /v1/autopays body. Manually added entries get
// source="manual" and status="active".
type CreateAutoPayRequest struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	DeductDay int     `json:"deductDay"`
	Notes     string  `json:"notes"`
}

// Validate checks the POST /v1/autopays body.
func (r CreateAutoPayRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 120)),
		validation.Field(&r.Type, validation.Required, validation.By(validType)),
		validation.Field(&r.Amount, validation.Required, validation.Min(0.0).Exclusive()),
		validation.Field(&r.DeductDay, validation.Required, validation.Min(1), validation.Max(31)),
	)
}

func validType(value any) error {
	s, _ := value.(string)
	if !autopaytype.Type(s).Valid() {
		return validation.NewError("validation_type", "unknown autopay type")
	}
	return nil
}
