package constants

// Header keys extracted by middleware into the request context.
const (
	HeaderAuthorization = "Authorization"
	HeaderRequestID     = "X-Request-Id"
	HeaderDeviceID      = "X-Device-Id"
	HeaderUserAgent     = "User-Agent"
	HeaderForwardedFor  = "X-Forwarded-For"
	HeaderRealIP        = "X-Real-Ip"
	// HeaderInternalKey carries the shared secret for calls to auth-service's
	// /v1/internal routes (phone → user lookup).
	HeaderInternalKey = "X-Internal-Key"
	// HeaderHubSignature is Meta's HMAC-SHA256 signature over the raw webhook body.
	HeaderHubSignature = "X-Hub-Signature-256"
)

// Context keys for values stored by middleware. Using a dedicated type avoids
// collisions with other packages writing to the same context.
type ContextKey string

const (
	CtxRequestID ContextKey = "requestID"
	CtxDeviceID  ContextKey = "deviceID"
	CtxIPAddress ContextKey = "ipAddress"
	CtxUserAgent ContextKey = "userAgent"
	CtxUserID    ContextKey = "userID"
	CtxRoles     ContextKey = "roles"
	CtxSessionID ContextKey = "sessionID"
	CtxVerified  ContextKey = "verified"
)

// BearerPrefix is the scheme prefix on the Authorization header.
const BearerPrefix = "Bearer "
