package response

// SalaryResponse is the PUT /v1/salary result — the stored salary inputs echoed
// back to the caller.
type SalaryResponse struct {
	MonthlySalary float64 `json:"monthlySalary"`
	SalaryDay     int     `json:"salaryDay"`
}
