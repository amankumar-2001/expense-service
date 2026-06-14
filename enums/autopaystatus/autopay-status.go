// Package autopaystatus enumerates the lifecycle states of an autopay record.
package autopaystatus

// Status is the lifecycle state of an autopay. Only Active entries count toward
// committed money.
type Status string

const (
	Active    Status = "active"
	Inactive  Status = "inactive"
	Cancelled Status = "cancelled"
	Duplicate Status = "duplicate"
)

func (s Status) String() string { return string(s) }

// Valid reports whether s is a recognised status.
func (s Status) Valid() bool {
	switch s {
	case Active, Inactive, Cancelled, Duplicate:
		return true
	default:
		return false
	}
}
