package endpoints_test

import (
	"testing"

	"github.com/reststore/restkit/internal/endpoints"
)

func TestExtractParams(t *testing.T) {
	t.Run("extracts query params", func(t *testing.T) {
		type ListRequest struct {
			Limit  int32 `query:"limit"  default:"20"`
			Offset int32 `query:"offset"`
		}

		params := endpoints.ExtractParams[ListRequest]()

		if len(params) != 2 {
			t.Fatalf("expected 2 params, got %d", len(params))
		}

		found := make(map[string]endpoints.Parameter)
		for _, p := range params {
			found[p.Name] = p
		}

		if p, ok := found["limit"]; !ok {
			t.Error("missing 'limit' param")
		} else {
			if p.Location != endpoints.ParamLocationQuery {
				t.Errorf("expected limit location=query, got %s", p.Location)
			}
			if p.Type != "integer" {
				t.Errorf("expected limit type=integer, got %s", p.Type)
			}
		}

		if _, ok := found["offset"]; !ok {
			t.Error("missing 'offset' param")
		}
	})

	t.Run("extracts path params", func(t *testing.T) {
		type GetRequest struct {
			ID int64 `path:"id"`
		}

		params := endpoints.ExtractParams[GetRequest]()

		if len(params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(params))
		}

		if params[0].Name != "id" {
			t.Errorf("expected name=id, got %s", params[0].Name)
		}
		if params[0].Location != endpoints.ParamLocationPath {
			t.Errorf("expected location=path, got %s", params[0].Location)
		}
		if !params[0].Required {
			t.Error("expected path param to be required")
		}
	})

	t.Run("extracts mixed params", func(t *testing.T) {
		type UpdateRequest struct {
			ID   int64  `path:"id"`
			Name string `          query:"name"`
		}

		params := endpoints.ExtractParams[UpdateRequest]()

		if len(params) != 2 {
			t.Fatalf("expected 2 params, got %d", len(params))
		}

		found := make(map[string]endpoints.Parameter)
		for _, p := range params {
			found[p.Name] = p
		}

		if p, ok := found["id"]; !ok {
			t.Error("missing 'id' param")
		} else if p.Location != endpoints.ParamLocationPath {
			t.Errorf("expected id location=path, got %s", p.Location)
		}

		if p, ok := found["name"]; !ok {
			t.Error("missing 'name' param")
		} else if p.Location != endpoints.ParamLocationQuery {
			t.Errorf("expected name location=query, got %s", p.Location)
		}
	})
}

func TestParseID(t *testing.T) {
	t.Run("valid int64 id", func(t *testing.T) {
		id, err := endpoints.ParseID("12345")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != 12345 {
			t.Errorf("expected id=12345, got %d", id)
		}
	})

	t.Run("valid large int64 id", func(t *testing.T) {
		id, err := endpoints.ParseID("9223372036854775807") // max int64
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != 9223372036854775807 {
			t.Errorf("expected max int64, got %d", id)
		}
	})

	t.Run("invalid id - non-numeric", func(t *testing.T) {
		_, err := endpoints.ParseID("abc")
		if err == nil {
			t.Error("expected error for non-numeric id")
		}
	})

	t.Run("invalid id - empty string", func(t *testing.T) {
		_, err := endpoints.ParseID("")
		if err == nil {
			t.Error("expected error for empty id")
		}
	})

	t.Run("invalid id - float", func(t *testing.T) {
		_, err := endpoints.ParseID("123.45")
		if err == nil {
			t.Error("expected error for float id")
		}
	})

	t.Run("negative id", func(t *testing.T) {
		id, err := endpoints.ParseID("-123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != -123 {
			t.Errorf("expected id=-123, got %d", id)
		}
	})

	t.Run("zero id", func(t *testing.T) {
		id, err := endpoints.ParseID("0")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != 0 {
			t.Errorf("expected id=0, got %d", id)
		}
	})
}

func TestParseIntID(t *testing.T) {
	t.Run("valid int id", func(t *testing.T) {
		id, err := endpoints.ParseIntID("12345")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != 12345 {
			t.Errorf("expected id=12345, got %d", id)
		}
	})

	t.Run("invalid id - non-numeric", func(t *testing.T) {
		_, err := endpoints.ParseIntID("abc")
		if err == nil {
			t.Error("expected error for non-numeric id")
		}
	})

	t.Run("invalid id - empty string", func(t *testing.T) {
		_, err := endpoints.ParseIntID("")
		if err == nil {
			t.Error("expected error for empty id")
		}
	})

	t.Run("valid zero id", func(t *testing.T) {
		id, err := endpoints.ParseIntID("0")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != 0 {
			t.Errorf("expected id=0, got %d", id)
		}
	})
}
