package request

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// UpdateAutoPayRequest is the PATCH /v1/autopays/:id body. All fields optional;
// only the supplied (non-nil) fields are updated.
type UpdateAutoPayRequest struct {
	Name      *string  `json:"name"`
	Type      *string  `json:"type"`
	Amount    *float64 `json:"amount"`
	DeductDay *int     `json:"deductDay"`
	Notes     *string  `json:"notes"`
	Status    *string  `json:"status"`
}

// Validate checks the PATCH /v1/autopays/:id body.
func (r UpdateAutoPayRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Length(1, 120)),
		validation.Field(&r.Type, validation.By(validTypePtr)),
		validation.Field(&r.Amount, validation.Min(0.0).Exclusive()),
		validation.Field(&r.DeductDay, validation.Min(1), validation.Max(31)),
	)
}

func validTypePtr(value any) error {
	p, ok := value.(*string)
	if !ok || p == nil {
		return nil
	}
	return validType(*p)
}
