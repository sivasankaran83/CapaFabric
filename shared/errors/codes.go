package errors

// String constants for ErrorCode values — use these when constructing raw
// InvokeError.Code strings in models (which use plain string, not ErrorCode).
const (
	ErrCodeCapabilityNotFound  = string(ErrCapabilityNotFound)
	ErrCodeManifestInvalid     = string(ErrManifestInvalid)
	ErrCodeManifestParseFailed = string(ErrManifestParseFailed)
	ErrCodeTransportError      = string(ErrTransportError)
	ErrCodeInvocationFailed    = string(ErrInvocationFailed)
	ErrCodeInternal            = string(ErrInternal)
)
