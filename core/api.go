package core

import (
	"github.com/reststore/restkit/internal/api"
	ep "github.com/reststore/restkit/internal/endpoints"
)

// Api is an alias for internal/api.Api. See restkit.Api for details.
type Api = api.Api

// NewApi creates a new Api instance.
var NewApi = api.New

// Group is an alias for internal/endpoints.Group. See restkit.Group for details.
type Group = ep.Group

// NewGroup creates a new group of endpoints with a common URL prefix.
var NewGroup = ep.NewGroup
