package utils

import (
	"context"

	"github.com/kharchibook/expense-service/constants"
)

// GetFromContext returns a string value previously stored under key, or "".
func GetFromContext(ctx context.Context, key constants.ContextKey) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}

// GetUserIDFromContext returns the authenticated user's ID stored by the JWT
// guard, or 0 if unauthenticated.
func GetUserIDFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(constants.CtxUserID).(int64); ok {
		return v
	}
	return 0
}

// GetRolesFromContext returns the roles slice stored by the JWT guard, or nil.
func GetRolesFromContext(ctx context.Context) []string {
	if v, ok := ctx.Value(constants.CtxRoles).([]string); ok {
		return v
	}
	return nil
}

// GetVerifiedFromContext reports whether the authenticated user is verified.
func GetVerifiedFromContext(ctx context.Context) bool {
	if v, ok := ctx.Value(constants.CtxVerified).(bool); ok {
		return v
	}
	return false
}
