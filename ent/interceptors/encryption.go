package interceptors

import (
	"context"
	"reflect"

	"entgo.io/ent"

	"github.com/theopenlane/ent/hooks"
)

// InterceptorEncryption provides transparent decryption for specified fields on query results
func InterceptorEncryption(fieldNames ...string) ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
			// Execute the query
			result, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			// Decrypt the specified fields using Tink
			return decryptQueryResult(result, fieldNames)
		})
	})
}

// InterceptorFieldEncryption provides decryption for a single field (for backward compatibility)
func InterceptorFieldEncryption(fieldName string, _ bool) ent.Interceptor {
	return InterceptorEncryption(fieldName)
}

// decryptQueryResult decrypts specified fields in query results using Tink
func decryptQueryResult(result ent.Value, fieldNames []string) (ent.Value, error) {
	if result == nil {
		return result, nil
	}

	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle slice of results
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := hooks.DecryptEntityFields(item.Interface(), fieldNames); err != nil {
				return nil, err
			}
		}

		return result, nil
	}

	// Handle single result
	err := hooks.DecryptEntityFields(result, fieldNames)

	return result, err
}
