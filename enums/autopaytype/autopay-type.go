// Package autopaytype enumerates the kinds of recurring commitment an autopay
// can represent.
package autopaytype

// Type is the kind of recurring deduction.
type Type string

const (
	EMI          Type = "emi"
	Subscription Type = "subscription"
	Insurance    Type = "insurance"
	SIP          Type = "sip"
	Other        Type = "other"
)

func (t Type) String() string { return string(t) }

// Valid reports whether t is a recognised autopay type.
func (t Type) Valid() bool {
	switch t {
	case EMI, Subscription, Insurance, SIP, Other:
		return true
	default:
		return false
	}
}
