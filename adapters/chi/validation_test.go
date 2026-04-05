package restchi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/reststore/restkit/internal/api"
	ep "github.com/reststore/restkit/internal/endpoints"
	errs "github.com/reststore/restkit/internal/errors"
)

// TestEndpointValidationWithAPIValidator tests that API-level validator is used by endpoints
func TestEndpointValidationWithAPIValidator(t *testing.T) {
	type TestUser struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	userCount := 0

	// Create validator that checks for short names and invalid emails
	validator := func(ctx context.Context, s any) errs.ValidationResult {
		result := errs.ValidationResult{}
		if req, ok := s.(CreateUserRequest); ok {
			hasErrors := false
			if len(req.Name) < 2 {
				hasErrors = true
				result.Errors = append(result.Errors, errs.ValidationError{
					Field:   "name",
					Message: "name must be at least 2 characters",
				})
			}
			if req.Email == "" || len(req.Email) < 3 {
				hasErrors = true
				result.Errors = append(result.Errors, errs.ValidationError{
					Field:   "email",
					Message: "email must be a valid email address",
				})
			}
			if hasErrors {
				result.Status = 422
				result.Code = errs.ErrCodeValidation
				result.Message = "validation failed"
			}
		}
		return result
	}

	r := chi.NewRouter()
	apiInstance := api.New().
		WithValidator(validator).
		WithTitle("Test API").
		WithVersion("1.0.0")

	// Add endpoint using NewEndpoint (not CRUD helper) with validation
	apiInstance.AddEndpoint(ep.NewEndpoint[CreateUserRequest, TestUser]().
		WithMethod(http.MethodPost).
		WithPath("/users").
		WithHandler(func(ctx context.Context, req CreateUserRequest) (TestUser, error) {
			t.Logf("Handler called - this should NOT happen for invalid requests!")
			userCount++
			return TestUser{
				ID:    userCount,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		}))

	RegisterRoutes(r, apiInstance)

	t.Run("valid request should succeed", func(t *testing.T) {
		userCount = 0

		body := []byte(`{"name":"John Doe","email":"john@example.com"}`)
		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		if userCount != 1 {
			t.Errorf("expected 1 user created, got %d", userCount)
		}
	})

	t.Run("invalid request should be blocked by validation", func(t *testing.T) {
		userCount = 0

		body := []byte(`{"name":"J","email":"x"}`)
		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status 422, got %d", rec.Code)
		}

		if userCount > 0 {
			t.Errorf("BUG: Handler was called %d time(s) despite validation failure", userCount)
		}
	})
}

// TestCRUDEndpointValidationExplicit tests validation with explicit WithValidation override
func TestCRUDEndpointValidationExplicit(t *testing.T) {
	type TestUser struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	userCount := 0

	// This is the endpoint-specific validation function
	endpointValidation := func(ctx context.Context, req CreateUserRequest) ep.ValidationResult {
		result := ep.ValidationResult{}
		hasErrors := false
		if len(req.Name) < 2 {
			hasErrors = true
			result.Errors = append(result.Errors, ep.ValidationError{
				Field:   "name",
				Message: "name must be at least 2 characters",
			})
		}
		if req.Email == "" {
			hasErrors = true
			result.Errors = append(result.Errors, ep.ValidationError{
				Field:   "email",
				Message: "email is required",
			})
		}
		if hasErrors {
			result.Status = 422
			result.Code = "validation"
			result.Message = "validation failed"
		}
		return result
	}

	r := chi.NewRouter()
	apiInstance := api.New().
		WithTitle("Test API").
		WithVersion("1.0.0")

	// Create endpoint with explicit validation (option B - endpoint override)
	endpoint := ep.NewEndpoint[CreateUserRequest, TestUser]().
		WithMethod(http.MethodPost).
		WithPath("/users").
		WithValidation(endpointValidation).
		WithHandler(func(ctx context.Context, req CreateUserRequest) (TestUser, error) {
			t.Logf("Handler called - this should NOT happen for invalid requests!")
			userCount++
			return TestUser{
				ID:    userCount,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		})

	apiInstance.AddEndpoint(endpoint)
	RegisterRoutes(r, apiInstance)

	t.Run("valid request should succeed", func(t *testing.T) {
		userCount = 0

		body := []byte(`{"name":"John Doe","email":"john@example.com"}`)
		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		if userCount != 1 {
			t.Errorf("expected 1 user created, got %d", userCount)
		}
	})

	t.Run("invalid request should be blocked by validation", func(t *testing.T) {
		userCount = 0

		body := []byte(`{"name":"J","email":""}`)
		req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status 422, got %d", rec.Code)
		}

		if userCount > 0 {
			t.Errorf("BUG: Handler was called %d time(s) despite validation failure", userCount)
		}
	})
}
