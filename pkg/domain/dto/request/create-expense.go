// Package request holds inbound request DTOs with their ozzo-validation rules.
// Each API endpoint that accepts a body has its own file here.
package request

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kharchibook/expense-service/enums/expensecategory"
)

// CreateExpenseRequest is the POST /v1/expenses body. This pass supports manual,
// structured entry; natural-language parsing of RawText is a later phase.
type CreateExpenseRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Note     string  `json:"note"`
	RawText  string  `json:"rawText"`
	// ExpenseDate is an optional ISO date (YYYY-MM-DD); defaults to today.
	ExpenseDate string `json:"expenseDate"`
}

// Validate checks the POST /v1/expenses body.
func (r CreateExpenseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Amount, validation.Required, validation.Min(0.0).Exclusive()),
		validation.Field(&r.Category, validation.Required, validation.By(validCategory)),
		validation.Field(&r.ExpenseDate, validation.Date("2006-01-02")),
	)
}

func validCategory(value any) error {
	s, _ := value.(string)
	if !expensecategory.Category(s).Valid() {
		return validation.NewError("validation_category", "unknown expense category")
	}
	return nil
}
