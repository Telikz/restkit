package errors

// Common error codes used throughout the API
const (
	// Error categories
	ErrCodeInternal      = "internal"      // Internal server error
	ErrCodeConfiguration = "configuration" // Configuration/setup error
	ErrCodeValidation    = "validation"    // Validation failed
	ErrCodeBind          = "bind"          // Request binding/parsing error
	ErrCodeNotFound      = "not_found"     // Resource not found
	ErrCodeUnauthorized  = "unauthorized"  // Authentication required
	ErrCodeForbidden     = "forbidden"     // Access denied
	ErrCodeBadRequest    = "bad_request"   // Malformed request
	ErrCodeMissingParam  = "missing_param" // Missing path parameter

	// Error messages
	ErrMsgInternal         = "An internal server error occurred"
	ErrMsgConfiguration    = "Endpoint is not properly configured"
	ErrMsgValidation       = "Validation failed"
	ErrMsgBind             = "Failed to parse request"
	ErrMsgNotFound         = "Resource not found"
	ErrMsgUnauthorized     = "Authentication required"
	ErrMsgForbidden        = "Access denied"
	ErrMsgBadRequest       = "Invalid request"
	ErrMsgHandlerNotSet    = "endpoint handler is not configured"
	ErrMsgMissingPathParam = "missing path parameter"
	ErrMsgInvalidInteger   = "invalid integer in path"
)
