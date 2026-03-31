package restchi

import (
	"errors"

	"github.com/go-chi/chi/v5"
	api "github.com/RestStore/RestKit/internal"
	"github.com/RestStore/RestKit/internal/schema"
)

// Mount mounts a Chi router to a RestKit API with automatic route extraction.
// It can optionally accept metadata to enhance route documentation.
// Pass nil for metas to extract all routes without additional metadata.
func Mount(
	a *api.Api,
	prefix string,
	router chi.Router,
	metas []schema.RouteMeta,
) error {
	var routes []schema.MountedRoute
	var err error

	if len(metas) > 0 {
		routes, err = Extract(router, metas)
	} else {
		routes, err = ExtractAll(router)
	}

	if err != nil {
		return errors.New(
			"extracting routes from chi router: " + err.Error(),
		)
	}

	a.MountRouter(prefix, router, routes)

	return nil
}
