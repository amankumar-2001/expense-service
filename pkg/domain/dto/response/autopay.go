package response

// AutoPayResponse is a single recurring commitment, returned by the autopay list,
// create, update, and confirm endpoints. id is serialised as a string to match
// the web client's AutoPay.id type. ConfidenceScore is omitted for
// confirmed/manual entries.
type AutoPayResponse struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Amount          float64  `json:"amount"`
	DeductDay       int      `json:"deductDay"`
	Source          string   `json:"source"`
	Status          string   `json:"status"`
	ConfidenceScore *float64 `json:"confidenceScore,omitempty"`
}
