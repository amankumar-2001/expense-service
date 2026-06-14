// Package entity holds internal value objects passed between domain services.
package entity

import "time"

// TokenClaims is the verified payload of an access JWT, attached to the request
// context by the JWT guard. It mirrors the claims auth-service mints.
type TokenClaims struct {
	UserID    int64
	SessionID int64
	Roles     []string
	Verified  bool
	IssuedAt  time.Time
	ExpiresAt time.Time
}
