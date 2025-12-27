package server

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/route"
)

// GenerateOpenAPISpecDocument builds a fully-registered OpenAPI document
func GenerateOpenAPISpecDocument() (*openapi3.T, error) {
	router, err := NewRouter(LogConfig{})
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

	log.Debug().Msg("generating OpenAPI specification document")

	// ensure tags are populated the same way server startup does
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
