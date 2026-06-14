package response

// CategoryTotalResponse is the summed spend for one category.
type CategoryTotalResponse struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

// ExpenseSummaryResponse is the monthly spend breakdown (GET /v1/expenses/summary).
type ExpenseSummaryResponse struct {
	Month      string                  `json:"month"` // e.g. "June 2026"
	Total      float64                 `json:"total"`
	ByCategory []CategoryTotalResponse `json:"byCategory"`
}
