//go:build codegen

// stub so ent codegen never compiles the possibly stale graphql generated package
package hooks

func isQueryValid(_ string, _ map[string]any) error {
	return nil
}
