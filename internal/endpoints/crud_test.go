package endpoints_test

import (
	"context"
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

func TestGetEndpoint(t *testing.T) {
	type GetRequest struct {
		ID int64 `path:"id"`
	}
	type GetResponse struct {
		Name string `json:"name"`
	}

	fn := func(ctx context.Context, queries any, req GetRequest) (GetResponse, error) {
		return GetResponse{Name: "test"}, nil
	}

	endpoint := endpoints.GetWithQueries[any, GetRequest, GetResponse]("/users/{id}", fn)

	if endpoint.Method != "GET" {
		t.Errorf("expected method=GET, got %s", endpoint.Method)
	}
	if endpoint.Path != "/users/{id}" {
		t.Errorf("expected path=/users/{id}, got %s", endpoint.Path)
	}
	if endpoint.Bind == nil {
		t.Error("expected Bind to be set")
	}
	if len(endpoint.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(endpoint.Parameters))
	}
}

func TestListEndpoint(t *testing.T) {
	type ListRequest struct {
		Limit  int32 `query:"limit"  default:"20"`
		Offset int32 `query:"offset" default:"0"`
	}
	type Item struct {
		ID int64 `json:"id"`
	}

	fn := func(ctx context.Context, queries any, req ListRequest) ([]Item, error) {
		return []Item{{ID: 1}}, nil
	}

	endpoint := endpoints.ListWithQueries[any, ListRequest, Item]("/users", fn)

	if endpoint.Method != "GET" {
		t.Errorf("expected method=GET, got %s", endpoint.Method)
	}
	if endpoint.Bind == nil {
		t.Error("expected Bind to be set")
	}
	if len(endpoint.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(endpoint.Parameters))
	}
}

func TestSearchEndpoint(t *testing.T) {
	type SearchRequest struct {
		Query *string `query:"q"`
	}
	type Item struct {
		Name string `json:"name"`
	}

	fn := func(ctx context.Context, queries any, req SearchRequest) ([]Item, error) {
		return []Item{{Name: "found"}}, nil
	}

	endpoint := endpoints.SearchWithQueries[any, SearchRequest, Item]("/search", fn)

	if endpoint.Method != "GET" {
		t.Errorf("expected method=GET, got %s", endpoint.Method)
	}
	if endpoint.Bind == nil {
		t.Error("expected Bind to be set")
	}
}

func TestCreateEndpoint(t *testing.T) {
	type CreateRequest struct {
		Name string `json:"name"`
	}
	type CreateResponse struct {
		ID int64 `json:"id"`
	}

	fn := func(ctx context.Context, queries any, req CreateRequest) (CreateResponse, error) {
		return CreateResponse{ID: 1}, nil
	}

	endpoint := endpoints.CreateWithQueries[any, CreateRequest, CreateResponse]("/users", fn)

	if endpoint.Method != "POST" {
		t.Errorf("expected method=POST, got %s", endpoint.Method)
	}
	if endpoint.Bind != nil {
		t.Error("expected Bind to be nil for Create (uses default)")
	}
}

func TestUpdateEndpoint(t *testing.T) {
	type UpdateRequest struct {
		ID   int64  `path:"id"`
		Name string `          json:"name"`
	}

	fn := func(ctx context.Context, queries any, req UpdateRequest) (struct{}, error) {
		return struct{}{}, nil
	}

	endpoint := endpoints.UpdateWithQueries("/users/{id}", fn)

	if endpoint.Method != "PATCH" {
		t.Errorf("expected method=PATCH, got %s", endpoint.Method)
	}
	if endpoint.Bind == nil {
		t.Error("expected Bind to be set")
	}
	if len(endpoint.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(endpoint.Parameters))
	}
}

func TestDeleteEndpoint(t *testing.T) {
	type DeleteRequest struct {
		ID int64 `path:"id"`
	}

	fn := func(ctx context.Context, queries any, req DeleteRequest) error {
		return nil
	}

	endpoint := endpoints.DeleteWithQueries[any, DeleteRequest]("/users/{id}", fn)

	if endpoint.Method != "DELETE" {
		t.Errorf("expected method=DELETE, got %s", endpoint.Method)
	}
	// DeleteWithQueries doesn't set Bind or Parameters - they are optional
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
