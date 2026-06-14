package response

// CommittedSummaryResponse is the free-money breakdown (GET /v1/analytics/committed).
type CommittedSummaryResponse struct {
	MonthlySalary  float64           `json:"monthlySalary"`
	SalaryDay      int               `json:"salaryDay"`
	TotalCommitted float64           `json:"totalCommitted"`
	FreeMoney      float64           `json:"freeMoney"`
	Autopays       []AutoPayResponse `json:"autopays"`
}
