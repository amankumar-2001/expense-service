package request

import validation "github.com/go-ozzo/ozzo-validation/v4"

// SetSalaryRequest is the PUT /v1/salary body — the inputs to the free-money
// calculation.
type SetSalaryRequest struct {
	Amount    float64 `json:"amount"`
	SalaryDay int     `json:"salaryDay"`
}

// Validate checks the PUT /v1/salary body.
func (r SetSalaryRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Amount, validation.Required, validation.Min(0.0).Exclusive()),
		validation.Field(&r.SalaryDay, validation.Required, validation.Min(1), validation.Max(31)),
	)
}
