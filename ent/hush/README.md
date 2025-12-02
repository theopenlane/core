# Hush field-Level encryption helper for ent

Hush provides automatic field-level encryption for Ent schemas. It was written originally with the intent of just managing the `secret_value` in the hush schema but ended up as a utility allowing you to simply annotate your ent schema sensitive fields with `hush.EncryptField()` and the framework handles encryption transparently.

## Features

- **Automatic Encryption** - Annotate fields and encryption is handled automatically
- **Optional Encryption** - Works without configuration for development, requires keyset for production
- **Transparent Operations** - Seamless encryption on write, decryption on read
- **Secure Implementation** - AES-256-GCM encryption via Google Tink
- **Smart Migration** - Handles existing unencrypted data gracefully
- **Graceful Degradation** - System operates normally even without encryption configured

## Quick Start

### Option 1: Development Mode (No Encryption)

For local development and testing, you can run without encryption:

```bash
# Simply don't set OPENLANE_TINK_KEYSET
# All fields marked with hush.EncryptField() will store plaintext
```

### Option 2: Production Mode (With Encryption)

For production environments where encryption is required:

#### Generate Encryption Key

As with most encryption systems, there needs to be a secret key used to perform the encryption and decryption.

A Taskfile entry has been created for your convenience; you can run `task hush:export`

## Annotate Your Fields

```go
import "github.com/theopenlane/ent/hush"

func (MySchema) Fields() []ent.Field {
    return []ent.Field{
        field.String("public_field"),
        field.String("secret_field").
            Sensitive().
            Annotations(hush.EncryptField()), // Mark field for encryption
    }
}

```

## API Reference

### Annotations

- `hush.EncryptField()` - Mark a field for automatic encryption

### Direct Encryption

The underlying crypto package provides direct encryption functions:

```go
import "github.com/theopenlane/core/internal/crypto"

// Encrypt data
encrypted, err := crypto.Encrypt(ctx, "plaintext")

// Decrypt data
decrypted, err := crypto.Decrypt(ctx, encrypted)
```

## Migration

Migration is automatic when you add `hush.EncryptField()` to an existing field:

1. System detects unencrypted values (not base64 encoded)
1. Encrypts them transparently during operations (if encryption is enabled)
1. Handles mixed encrypted/unencrypted data gracefully
1. No manual migration steps required

### Important Migration Scenarios

- **Enabling Encryption**: Set keyset, existing data migrates gradually
- **Disabling Encryption**: Remove keyset, new data is plaintext but old encrypted data becomes unreadable
- **Development → Production**: Start without keyset locally, add keyset in production

## Configuration

### Environment Variables

```bash
# Optional: Tink keyset for encryption/decryption
# If not set, encryption is disabled and data is stored as plaintext
OPENLANE_TINK_KEYSET=<base64-keyset>
```

### Encryption Behavior

- **With Keyset**: Fields marked with `hush.EncryptField()` are encrypted
- **Without Keyset**: Fields marked with `hush.EncryptField()` store plaintext
- **No Errors**: System operates normally in both modes

### Key Storage

The implementation uses the key from the environment variable at runtime. How you populate that environment variable is up to your deployment infrastructure.

**WARNING**: Be consistent with your encryption configuration. If you encrypt data with a keyset, you'll need that same keyset to decrypt it later.

## Technical Details

### Encryption Specification

- **Algorithm**: AES-256-GCM with AEAD (Authenticated Encryption with Associated Data)
- **Library**: Google Tink - production-ready cryptography library
- **Key Size**: 256-bit AES keys
- **Nonce**: Unique per encryption operation (managed by Tink)
- **Encoding**: Base64 for database-safe storage

### Storage Format

Encrypted data is stored as base64-encoded strings to ensure compatibility with all database systems. Raw encrypted bytes contain:

- Null bytes that can terminate strings in some databases
- Non-UTF8 sequences that cause encoding errors
- Control characters that break text protocols

Base64 encoding ensures encrypted data can be safely stored as text in any database.

## Best Practices

### Recommended Usage

- Encrypt only truly sensitive fields (passwords, tokens, API keys, secrets)
- Always mark encrypted fields with `Sensitive()` annotation
- Test encryption thoroughly in development before production deployment
- Implement proper key management for production environments
- Monitor performance impact for large-scale deployments

### Common Pitfalls

- Avoid encrypting fields used in WHERE clauses (IDs, usernames, emails)
- Do not encrypt fields that need to be indexed or searched
- Never commit encryption keys to version control
- Be cautious with very large text fields (consider performance impact)
- Remember that encrypted fields cannot be used in database-level constraints

## Benchmarks

### Direct Encryption/Decryption:

- Time: ~800 nanoseconds per encrypt+decrypt cycle
- Memory: ~912 bytes allocated per operation
- Allocations: 7 memory allocations per operation

This means each field encryption/decryption adds roughly 0.8 microseconds of overhead.

Field Detection (Reflection):

- Time: ~1.7 microseconds per field detection
- Memory: ~2.9KB allocated per detection
- Allocations: 35 memory allocations per detection

This is a one-time cost during application startup when the encryption hooks are registered.

### 🏁 Real-World Impact

For typical database operations:

- Single entity with 1 encrypted field: +0.8μs overhead
- Single entity with 5 encrypted fields: +4μs overhead
- Bulk operations (100 entities, 1 encrypted field each): +80μs overhead

### Context:

- Network round-trip to database: ~1-10ms (1,000-10,000μs)
- Database query execution: ~100μs-1ms (100-1,000μs)
- Encryption overhead: ~0.8μs per field

### Bottom Line

The encryption overhead is negligible compared to typical database operations:
- <0.1% of typical database query time
- <0.01% of network round-trip time

The field-level encryption provides strong security with minimal performance overhead.
