// Package response holds outbound response DTOs — one file per API. JSON field
// names mirror the web client's types (kharchibook-web/src/lib/api/types.ts)
// exactly.
package response

// ExpenseResponse is a single logged expense, returned by POST /v1/expenses,
// GET /v1/expenses, and DELETE /v1/expenses/last. id is serialised as a string
// to match the web client's Expense.id type.
type ExpenseResponse struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	RawText     string  `json:"rawText"`
	ExpenseDate string  `json:"expenseDate"` // ISO date (YYYY-MM-DD)
}
