# Entitlements Catalog

## Overview

## JSONSchema generation

`go run pkg/catalog/genjsonschema/catalog_schema.go` -> generates a jsonschema from the structs inside of catalog.go.

This jsonschema is how we validate that the yaml input is indeed valid and coforms to the specification. The `LoadCatalog` function will fail if the yaml input does not conform to the schema specification, offering some guardrails in the event of misconfiguration / bad yaml / missing fields, etc.

` go run ./cmd/catalog --stripe-key="[insertstrpekey]"` will take the catalog file, pull products and prices from stripe, and prompt you as to whether or not:
- the definitions in your catalog file have corresponding products and prices in the stripe instance matching the key you provided
-



## Catalog Versioning

The `SaveCatalog` method is responsible for saving a `Catalog` struct to disk in YAML format, while also managing versioning and integrity via a SHA hash. The function first checks if the receiver (`c`) is `nil`, in which case it returns immediately with no error. It then attempts to read the existing catalog file from the provided path. If the file does not exist, that's acceptable, but any other read error is returned.

If the file exists and contains data, it is unmarshaled from YAML into a temporary `Catalog` struct called `orig`. This allows the function to compare the current catalog with the previous version. If the current catalog's version is unset but the original has one, the version is carried over. Similarly, if the SHA is missing, it is computed based on the version string.

The catalog is then marshaled into YAML. If the new YAML differs from the original file's contents, the function attempts to bump the patch version (using semantic versioning) and recomputes the SHA. The catalog is re-marshaled to reflect these changes. A unified diff is generated between the old and new YAML representations, providing a summary of what changed.

Finally, the new YAML data is written to disk with standard permissions. The function returns the diff string, which can be used for logging or review. This approach ensures that the catalog file is always up-to-date, versioned, and its integrity can be verified via the SHA. A subtle point is that the version is only bumped if the contents have changed, helping to avoid unnecessary version increments.
