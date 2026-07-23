package server

import (
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/route"
)

// GenerateOpenAPISpecDocument builds a fully-registered OpenAPI document
func GenerateOpenAPISpecDocument() (*openapi3.T, error) {
	router, err := registerSpecRoutes()
	if err != nil {
		return nil, err
	}

	if missing := missingSpecTypes(router); len(missing) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrMissingSpecInstances, missing)
	}

	log.Debug().Msg("generating OpenAPI specification document")

	// ensure tags are populated from the registered operations
	spec := generateTagsFromOperations(router.OAS)

	loader := openapi3.NewLoader()

	if err := loader.ResolveRefsIn(spec, nil); err != nil {
		log.Error().Err(err).Msg("failed to resolve OpenAPI references")

		return nil, err
	}

	if err := spec.Validate(loader.Context); err != nil {
		log.Error().Err(err).Msg("openapi spec validation failed")

		return nil, err
	}

	return spec, nil
}

// CollectSpecTypes registers all routes in spec-build mode and returns the sorted qualified names
// of every model type the analyzed handlers need; used by the instance emitter
func CollectSpecTypes() ([]string, error) {
	router, err := registerSpecRoutes()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(router.SpecTypes))
	for name := range router.SpecTypes {
		names = append(names, name)
	}

	sort.Strings(names)

	return names, nil
}

// registerSpecRoutes builds a spec-mode router with all routes registered
func registerSpecRoutes() (*route.Router, error) {
	router, err := NewSpecRouter(LogConfig{})
	if err != nil {
		log.Error().Err(err).Msg("failed to create router")

		return nil, err
	}

	// minimal handler so route registration can succeed without hitting dependencies
	router.Handler = &handlers.Handler{}

	log.Debug().Msg("registering routes for OpenAPI spec generation")

	if err := route.RegisterRoutes(router); err != nil {
		log.Error().Err(err).Msg("failed to register routes for OpenAPI spec generation")

		return nil, err
	}

	return router, nil
}

// missingSpecTypes returns the analyzed model types that have no generated instance
func missingSpecTypes(router *route.Router) []string {
	missing := make([]string, 0)

	for name := range router.SpecTypes {
		if _, ok := router.SpecInstances[name]; !ok {
			missing = append(missing, name)
		}
	}

	sort.Strings(missing)

	return missing
}
