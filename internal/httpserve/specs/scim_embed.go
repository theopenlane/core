package specs

import _ "embed"

// SCIMSpec contains the embedded SCIM 2.0 OpenAPI specification

//go:embed scim.yaml
var SCIMSpec []byte

// OpenlaneSpec contains the fully-composed OpenAPI 3.1 specification served at /api-docs

//go:embed openlane.openapi.json
var OpenlaneSpec []byte
