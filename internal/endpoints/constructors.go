package endpoints

// NewEndpoint creates a new Endpoint with auto-generated schemas and sensible defaults.
// Default method: GET for NoRequest, DELETE for NoResponse, POST otherwise.
func NewEndpoint[Req any, Res any]() *Endpoint[Req, Res] {
	return &Endpoint[Req, Res]{
		Title:       "Not set",
		Description: "Not set",
		Path:        "/not-set",
		Handler:     nil,
	}
}

// NewEndpointRes creates an endpoint with NoRequest type for response-only endpoints.
// Default method is GET.
func NewEndpointRes[Res any]() *Endpoint[NoRequest, Res] {
	return NewEndpoint[NoRequest, Res]()
}

// NewEndpointReq creates an endpoint with NoResponse type for request-only endpoints.
// Default method is DELETE.
func NewEndpointReq[Req any]() *Endpoint[Req, NoResponse] {
	return NewEndpoint[Req, NoResponse]()
}
