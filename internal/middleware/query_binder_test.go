package middleware

import (
	"net/http"
	"net/url"
	"testing"
)

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func TestQueryBinder(t *testing.T) {
	type SearchRequest struct {
		ID        *string `query:"id"`
		Name      *string `query:"name"`
		Limit     int32   `query:"limit"      default:"20"`
		Offset    int32   `query:"offset"     default:"0"`
		CreatedAt *string `query:"created_at"`
	}

	binder := QueryBinder[SearchRequest]()

	t.Run("binds query parameters", func(t *testing.T) {
		req := &http.Request{
			URL:    mustParseURL("/search?id=123&name=john&created_at=2024-01-01"),
			Method: "GET",
		}

		result, err := binder(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.ID == nil || *result.ID != "123" {
			t.Errorf("expected ID=123, got %v", result.ID)
		}
		if result.Name == nil || *result.Name != "john" {
			t.Errorf("expected Name=john, got %v", result.Name)
		}
		if result.CreatedAt == nil || *result.CreatedAt != "2024-01-01" {
			t.Errorf("expected CreatedAt=2024-01-01, got %v", result.CreatedAt)
		}
	})

	t.Run("applies defaults", func(t *testing.T) {
		req := &http.Request{
			URL:    mustParseURL("/search"),
			Method: "GET",
		}

		result, err := binder(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Limit != 20 {
			t.Errorf("expected default Limit=20, got %d", result.Limit)
		}
		if result.Offset != 0 {
			t.Errorf("expected default Offset=0, got %d", result.Offset)
		}
	})

	t.Run("leaves optional fields nil", func(t *testing.T) {
		req := &http.Request{
			URL:    mustParseURL("/search"),
			Method: "GET",
		}

		result, err := binder(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.ID != nil {
			t.Errorf("expected ID to be nil, got %v", *result.ID)
		}
		if result.Name != nil {
			t.Errorf("expected Name to be nil, got %v", *result.Name)
		}
	})
}

func TestQueryBinderWithPathParams(t *testing.T) {
	type GetRequest struct {
		ID int64 `path:"id"`
	}

	binder := QueryBinder[GetRequest]()

	t.Run("binds path parameter", func(t *testing.T) {
		req := &http.Request{
			URL:    mustParseURL("/users/123"),
			Method: "GET",
		}
		req.SetPathValue("id", "123")

		result, err := binder(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.ID != 123 {
			t.Errorf("expected ID=123, got %d", result.ID)
		}
	})

	t.Run("returns error for missing path param", func(t *testing.T) {
		req := &http.Request{
			URL:    mustParseURL("/users/"),
			Method: "GET",
		}

		_, err := binder(req)
		if err == nil {
			t.Error("expected error for missing path parameter")
		}
	})
}

func TestQueryBinderTypes(t *testing.T) {
	type TypedRequest struct {
		StringField string  `query:"s"`
		IntField    int     `query:"i"`
		Int64Field  int64   `query:"i64"`
		Int32Field  int32   `query:"i32"`
		BoolField   bool    `query:"b"`
		FloatField  float64 `query:"f"`
		OptionalInt *int    `query:"opt"`
	}

	binder := QueryBinder[TypedRequest]()

	t.Run("binds various types", func(t *testing.T) {
		req := &http.Request{
			URL:    mustParseURL("/test?s=hello&i=42&i64=100&i32=50&b=true&f=3.14&opt=99"),
			Method: "GET",
		}

		result, err := binder(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.StringField != "hello" {
			t.Errorf("expected StringField=hello, got %s", result.StringField)
		}
		if result.IntField != 42 {
			t.Errorf("expected IntField=42, got %d", result.IntField)
		}
		if result.Int64Field != 100 {
			t.Errorf("expected Int64Field=100, got %d", result.Int64Field)
		}
		if result.Int32Field != 50 {
			t.Errorf("expected Int32Field=50, got %d", result.Int32Field)
		}
		if result.BoolField != true {
			t.Errorf("expected BoolField=true, got %v", result.BoolField)
		}
		if result.FloatField != 3.14 {
			t.Errorf("expected FloatField=3.14, got %f", result.FloatField)
		}
		if result.OptionalInt == nil || *result.OptionalInt != 99 {
			t.Errorf("expected OptionalInt=99, got %v", result.OptionalInt)
		}
	})
}

func TestExtractQueryParams(t *testing.T) {
	type TestRequest struct {
		ID     *string `query:"id"`
		Name   string  `query:"name"`
		Limit  int32   `query:"limit"  default:"20"`
		Offset int32   `query:"offset" default:"0"  required:"true"`
	}

	params := ExtractQueryParams[TestRequest]()

	if len(params) != 4 {
		t.Fatalf("expected 4 params, got %d", len(params))
	}

	found := make(map[string]QueryParamInfo)
	for _, p := range params {
		found[p.Name] = p
	}

	if p, ok := found["id"]; !ok {
		t.Error("missing 'id' param")
	} else if p.Type != "string" {
		t.Errorf("expected id type=string, got %s", p.Type)
	}

	if p, ok := found["name"]; !ok {
		t.Error("missing 'name' param")
	} else if p.Type != "string" {
		t.Errorf("expected name type=string, got %s", p.Type)
	}

	if p, ok := found["limit"]; !ok {
		t.Error("missing 'limit' param")
	} else if p.Default != "20" {
		t.Errorf("expected limit default=20, got %s", p.Default)
	}

	if p, ok := found["offset"]; !ok {
		t.Error("missing 'offset' param")
	} else if !p.Required {
		t.Error("expected offset to be required")
	}
}

func TestExtractPathParams(t *testing.T) {
	type TestRequest struct {
		UserID int64  `path:"user_id"`
		PostID string `path:"post_id"`
	}

	params := ExtractPathParams[TestRequest]()

	if len(params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(params))
	}

	found := make(map[string]PathParamInfo)
	for _, p := range params {
		found[p.Name] = p
	}

	if p, ok := found["user_id"]; !ok {
		t.Error("missing 'user_id' param")
	} else if p.Type != "integer" {
		t.Errorf("expected user_id type=integer, got %s", p.Type)
	}

	if p, ok := found["post_id"]; !ok {
		t.Error("missing 'post_id' param")
	} else if p.Type != "string" {
		t.Errorf("expected post_id type=string, got %s", p.Type)
	}
}
