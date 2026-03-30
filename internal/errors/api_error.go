package errors

// APIError represents a standardized API error response
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new API error
func NewAPIError(status int, code, message string) APIError {
	return APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(
	status int, code, message, details string,
) APIError {
	return APIError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) (APIError, bool) {
	if apiErr, ok := err.(APIError); ok {
		return apiErr, true
	}
	return APIError{}, false
}