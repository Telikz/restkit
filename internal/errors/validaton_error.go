package errors

// ValidationResult is returned by validation functions
type ValidationResult struct {
	Status  int               `json:"status"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v ValidationResult) IsValid() bool {
	return len(v.Errors) == 0
}

func (v ValidationResult) ToAPIError() APIError {
	return APIError{
		Status:  v.Status,
		Code:    v.Code,
		Message: v.Message,
	}
}

// HasErrors returns true if there are validation errors
func (v *ValidationResult) HasErrors() bool {
	if v == nil {
		return false
	}
	return len(v.Errors) > 0
}

// NewValidation creates an empty validation result to populate with errors
func NewValidation() ValidationResult {
	return ValidationResult{}
}

// ValidationFailed creates a failed validation result with a single error
func ValidationFailed(
	status int, code, message, field, fieldMessage string,
) ValidationResult {
	return ValidationResult{
		Status:  status,
		Code:    code,
		Message: message,
		Errors: []ValidationError{
			{Field: field, Message: fieldMessage},
		},
	}
}

// ValidationFailedMulti creates a failed validation result with multiple errors
func ValidationFailedMulti(
	status int, code, message string, errors ...ValidationError,
) ValidationResult {
	return ValidationResult{
		Status:  status,
		Code:    code,
		Message: message,
		Errors:  errors,
	}
}

// AddError adds a validation error to the result
func (v *ValidationResult) AddError(field, message string) {
	v.Errors = append(v.Errors,
		ValidationError{Field: field, Message: message})
}

// WithStatus sets the HTTP status code on the result
func (v *ValidationResult) WithStatus(status int) {
	v.Status = status
}

// WithCode sets the error code on the result
func (v *ValidationResult) WithCode(code string) {
	v.Code = code
}

// WithMessage sets the message on the result
func (v *ValidationResult) WithMessage(message string) {
	v.Message = message
}
