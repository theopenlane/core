// Package examples provides working examples of hush field-level encryption.
package examples

import (
	"context"
	"fmt"
	"log"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/hush"
)

// OAuthIntegrationExample demonstrates field-level encryption for OAuth integrations
type OAuthIntegrationExample struct {
	ent.Schema
}

func (OAuthIntegrationExample) Fields() []ent.Field {
	return []ent.Field{
		// Public fields - not encrypted
		field.String("provider").
			Comment("OAuth provider (github, google, etc.)"),
		field.String("client_id").
			Comment("OAuth client ID (public)"),
		field.String("scopes").
			Optional().
			Comment("Requested OAuth scopes"),

		// Encrypted fields - automatically handled
		field.String("client_secret").
			Sensitive().
			Annotations(
				hush.EncryptField(),
			).
			Comment("OAuth client secret (encrypted)"),

		field.String("access_token").
			Optional().
			Sensitive().
			Annotations(
				hush.EncryptField(),
			).
			Comment("OAuth access token (encrypted)"),

		field.String("refresh_token").
			Optional().
			Sensitive().
			Annotations(
				hush.EncryptField(),
			).
			Comment("OAuth refresh token (encrypted)"),

		field.String("webhook_secret").
			Optional().
			Sensitive().
			Annotations(
				hush.EncryptField(),
			).
			Comment("Webhook signing secret (encrypted)"),
	}
}

func (o OAuthIntegrationExample) Mixin() []ent.Mixin {
	// No manual encryption mixin needed!
	// Fields with hush.EncryptField() annotations are automatically encrypted
	return []ent.Mixin{
		// Add other mixins as needed
	}
}

// Hooks are automatically applied for encrypted fields
func (o OAuthIntegrationExample) Hooks() []ent.Hook {
	return hush.AutoEncryptionHook(o)
}

// Interceptors are automatically applied for encrypted fields
func (o OAuthIntegrationExample) Interceptors() []ent.Interceptor {
	return hush.AutoDecryptionInterceptor(o)
}

// ExampleOAuthIntegrationUsage demonstrates how to use the encrypted OAuth integration
func ExampleOAuthIntegrationUsage() {
	// This is a conceptual example - in practice you'd use your actual Ent client

	ctx := context.Background()

	// Assuming you have an Ent client configured with the schema
	// client := ent.NewClient(ent.Driver(driver))

	fmt.Println("=== OAuth Integration Encryption Example ===")

	// Create integration with sensitive data
	// Data is automatically encrypted before storage
	fmt.Println("\n1. Creating OAuth integration (automatic encryption)...")

	integration := &struct {
		Provider      string
		ClientID      string
		ClientSecret  string
		AccessToken   string
		RefreshToken  string
		WebhookSecret string
	}{
		Provider:      "github",
		ClientID:      "github-client-id-12345",
		ClientSecret:  "ghcs_super_secret_client_secret_xyz789",    // Encrypted
		AccessToken:   "gho_access_token_abcdef123456789",          // Encrypted
		RefreshToken:  "ghr_refresh_token_uvwxyz987654321",         // Encrypted
		WebhookSecret: "webhook_signing_secret_qrstuvwxyz123456",   // Encrypted
	}

	fmt.Printf("Provider: %s\n", integration.Provider)
	fmt.Printf("Client ID: %s\n", integration.ClientID)
	fmt.Printf("Client Secret: %s (will be encrypted)\n", integration.ClientSecret)
	fmt.Printf("Access Token: %s (will be encrypted)\n", integration.AccessToken)

	// When saved to database, encrypted fields are base64-encoded ciphertext:
	// client_secret: "AZO/odmO3exK1Cp+exH8XRMJAtjGuFuoMIDFfz0q4sqI..."
	// access_token:  "BXP/pdnP4fyL2Dq7fyI9YSNKBukHvGvpNJEGg05r5tr..."

	fmt.Println("\n2. Reading integration (automatic decryption)...")

	// When queried, fields are automatically decrypted
	// integration, err := client.OAuthIntegration.Get(ctx, id)
	// Values returned are plaintext for application use

	fmt.Printf("Retrieved Client Secret: %s (decrypted)\n", integration.ClientSecret)
	fmt.Printf("Retrieved Access Token: %s (decrypted)\n", integration.AccessToken)

	fmt.Println("\n3. Security features:")
	fmt.Println("- Fields marked Sensitive() are hidden from logs")
	fmt.Println("- Encryption uses AES-256-GCM via Google Tink")
	fmt.Println("- Each encryption uses a unique nonce")
	fmt.Println("- Data is base64-encoded for safe database storage")
	fmt.Println("- Supports key rotation without re-encrypting data")

	_ = ctx // Satisfy linter
}

// ExampleManualEncryption shows direct use of encryption functions
func ExampleManualEncryption() {
	fmt.Println("=== Manual Encryption Example ===")

	// This would typically be imported:
	// "github.com/theopenlane/core/internal/ent/hooks"

	// Example of direct encryption/decryption
	secretData := "super-secret-api-key-xyz123"

	fmt.Printf("Original: %s\n", secretData)

	// Note: In practice, you'd use hooks.Encrypt()
	// encrypted, err := hooks.Encrypt([]byte(secretData))
	// For demo purposes, simulate the result
	encrypted := "AZO/odmO3exK1Cp+exH8XRMJAtjGuFuoMIDFfz0q4sqIuKfCnQ=="

	fmt.Printf("Encrypted: %s\n", encrypted)

	// And hooks.Decrypt()
	// decrypted, err := hooks.Decrypt(encrypted)
	decrypted := secretData // Simulated result

	fmt.Printf("Decrypted: %s\n", decrypted)

	fmt.Println("\nNote: This demonstrates the process - actual usage would be:")
	fmt.Println("  encrypted, err := hooks.Encrypt([]byte(secretData))")
	fmt.Println("  decrypted, err := hooks.Decrypt(encrypted)")
}

// ExampleKeyGeneration shows how to generate encryption keys
func ExampleKeyGeneration() {
	fmt.Println("=== Key Generation Example ===")

	fmt.Println("1. Generate a new Tink keyset:")
	fmt.Println("   go run ./cmd/generate-tink-keyset")
	fmt.Println("")
	fmt.Println("2. Example output:")
	fmt.Println("   OPENLANE_TINK_KEYSET=CNnD/p0JEmQKWAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EiIaID+JaHu+6zMW3YgphNkpL5lVJMVeZdAjJclgAHyShxUOGAEQARjZw/6dCSAB")
	fmt.Println("")
	fmt.Println("3. Set environment variable:")
	fmt.Println("   export OPENLANE_TINK_KEYSET=<generated-keyset>")
	fmt.Println("")
	fmt.Println("4. Test encryption:")
	fmt.Println("   go run ./cmd/hush-demo -secret=\"test\" -quiet")

	// In practice:
	// keyset, err := hooks.GenerateTinkKeyset()
	// if err != nil {
	//     log.Fatal(err)
	// }
	// fmt.Printf("Generated keyset: %s\n", keyset)
}

// ExampleMigration demonstrates migrating existing unencrypted data
func ExampleMigration() {
	fmt.Println("=== Migration Example ===")

	fmt.Println("Scenario: You have existing OAuth integrations with unencrypted secrets")
	fmt.Println("Goal: Encrypt the client_secret field without breaking existing data")
	fmt.Println("")

	fmt.Println("1. Simply add the annotation to your field:")
	fmt.Println("   field.String(\"client_secret\").")
	fmt.Println("       Sensitive().")
	fmt.Println("       Annotations(hush.EncryptField())")
	fmt.Println("")
	fmt.Println("   The system automatically handles migration - no mixins needed!")
	fmt.Println("")

	fmt.Println("2. Or use migration tool:")
	fmt.Println("   import \"github.com/theopenlane/core/internal/ent/hush\"")
	fmt.Println("")
	fmt.Println("   err := hush.MigrateFieldToEncryption(")
	fmt.Println("       ctx, db, \"oauth_integrations\", \"client_secret\")")
	fmt.Println("")

	fmt.Println("3. Validate migration:")
	fmt.Println("   err := hush.ValidateFieldEncryption(")
	fmt.Println("       ctx, db, \"oauth_integrations\", \"client_secret\")")
	fmt.Println("")

	fmt.Println("Migration process:")
	fmt.Println("- Scans all rows in the table")
	fmt.Println("- Identifies unencrypted values (not base64)")
	fmt.Println("- Encrypts and updates each value")
	fmt.Println("- Skips already encrypted values")
	fmt.Println("- Logs progress every 100 rows")
}

// ExampleSimplifiedFieldDefinition shows the new simplified approach
func ExampleSimplifiedFieldDefinition() {
	fmt.Println("=== Simplified Field Definition ===")

	fmt.Println("The new approach: just annotate your fields!")
	fmt.Println("")

	fmt.Println("OAuth Client Secret:")
	fmt.Println("   field.String(\"client_secret\").")
	fmt.Println("       Sensitive().")
	fmt.Println("       Annotations(hush.EncryptField())")
	fmt.Println("")

	fmt.Println("API Token:")
	fmt.Println("   field.String(\"api_token\").")
	fmt.Println("       Sensitive().")
	fmt.Println("       Annotations(hush.EncryptField())")
	fmt.Println("")

	fmt.Println("Database Password:")
	fmt.Println("   field.String(\"db_password\").")
	fmt.Println("       Sensitive().")
	fmt.Println("       Annotations(hush.EncryptField())")
	fmt.Println("")

	fmt.Println("That's it! No mixins, no manual hook setup, no field specification.")
	fmt.Println("Encryption and decryption happen automatically.")
}

// ExampleSecurityConsiderations shows security best practices
func ExampleSecurityConsiderations() {
	fmt.Println("=== Security Considerations ===")

	fmt.Println("✅ DO:")
	fmt.Println("- Only encrypt truly sensitive data (passwords, tokens, keys)")
	fmt.Println("- Mark encrypted fields as Sensitive() to hide from logs")
	fmt.Println("- Use proper key storage (AWS Secrets Manager, etc.)")
	fmt.Println("- Set up key rotation policies")
	fmt.Println("- Monitor for decryption failures")
	fmt.Println("- Test migrations thoroughly on staging")
	fmt.Println("- Use separate keys for different environments")
	fmt.Println("")

	fmt.Println("❌ DON'T:")
	fmt.Println("- Encrypt fields used in WHERE clauses (breaks queries)")
	fmt.Println("- Encrypt IDs, usernames, or other searchable data")
	fmt.Println("- Store keys in environment variables in production")
	fmt.Println("- Commit keys to version control")
	fmt.Println("- Encrypt very large fields (performance impact)")
	fmt.Println("- Change keys without migration plan")
	fmt.Println("")

	fmt.Println("🔍 MONITORING:")
	fmt.Println("- Alert on decryption failures")
	fmt.Println("- Track encryption/decryption performance")
	fmt.Println("- Monitor for plaintext leaks in logs")
	fmt.Println("- Audit field access patterns")
	fmt.Println("")

	fmt.Println("📋 COMPLIANCE:")
	fmt.Println("- Document which fields are encrypted")
	fmt.Println("- Maintain encryption inventory")
	fmt.Println("- Track key rotation history")
	fmt.Println("- Implement access controls")
}

// ExampleMain demonstrates running the examples
func ExampleMain() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Hush Field-Level Encryption Examples")
	fmt.Println("====================================")
	fmt.Println("")

	ExampleOAuthIntegrationUsage()
	fmt.Println("")

	ExampleManualEncryption()
	fmt.Println("")

	ExampleKeyGeneration()
	fmt.Println("")

	ExampleMigration()
	fmt.Println("")

	ExampleSimplifiedFieldDefinition()
	fmt.Println("")

	ExampleSecurityConsiderations()
}