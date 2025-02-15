package types

// ContextKey type for request context keys
type ContextKey string

const (
	// Context keys
	CtxKeyLogger ContextKey = "logger"
	CtxKeyDB     ContextKey = "db"
)
