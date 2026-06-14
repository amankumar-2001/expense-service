package response

// UpcomingDeductionResponse is one deduction due in the requested window
// (GET /v1/analytics/upcoming).
type UpcomingDeductionResponse struct {
	AutopayID string  `json:"autopayId"`
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	DeductDay int     `json:"deductDay"`
	InDays    int     `json:"inDays"`
}
