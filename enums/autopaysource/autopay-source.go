// Package autopaysource enumerates how an autopay entry was created.
package autopaysource

// Source records the origin of an autopay record.
type Source string

const (
	// Manual entries are added directly by the user.
	Manual Source = "manual"
	// EmailAuto entries are detected from the user's connected mailbox and await
	// confirmation before counting as committed.
	EmailAuto Source = "email_auto"
)

func (s Source) String() string { return string(s) }
