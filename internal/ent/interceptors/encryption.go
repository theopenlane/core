package interceptors

import (
	"context"
	"reflect"

	"entgo.io/ent"
	"gocloud.dev/secrets"

	"github.com/theopenlane/core/internal/ent/hooks"
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

			// Get the secrets keeper from the query
			secretsKeeper := getSecretsKeeperFromQuery(q)
			if secretsKeeper == nil {
				// Use fallback AES decryption
				return decryptQueryResultAES(result, fieldNames)
			}

			// Decrypt the specified fields
			if err := decryptQueryResult(ctx, secretsKeeper, result, fieldNames); err != nil {
				return nil, err
			}

			return result, nil
		})
	})
}

// decryptQueryResult decrypts specified fields in query results
func decryptQueryResult(ctx context.Context, keeper *secrets.Keeper, result ent.Value, fieldNames []string) error {
	if result == nil {
		return nil
	}

	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle slice of results
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := hooks.DecryptEntityFields(ctx, keeper, item.Interface(), fieldNames); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle single result
	return hooks.DecryptEntityFields(ctx, keeper, result, fieldNames)
}

// decryptQueryResultAES decrypts fields using AES fallback
func decryptQueryResultAES(result ent.Value, fieldNames []string) (ent.Value, error) {
	if result == nil {
		return result, nil
	}

	// Get encryption key
	key := hooks.GetEncryptionKey()

	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle slice of results
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := hooks.DecryptEntityFieldsAES(item.Interface(), key, fieldNames); err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	// Handle single result
	err := hooks.DecryptEntityFieldsAES(result, key, fieldNames)
	return result, err
}

// getSecretsKeeperFromQuery extracts secrets keeper from query
func getSecretsKeeperFromQuery(q ent.Query) *secrets.Keeper {
	v := reflect.ValueOf(q)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Look for Secrets field
	secretsField := v.FieldByName("Secrets")
	if secretsField.IsValid() && !secretsField.IsNil() {
		if keeper, ok := secretsField.Interface().(*secrets.Keeper); ok {
			return keeper
		}
	}

	return nil
}