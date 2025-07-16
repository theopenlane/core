# Hush - Field-Level Encryption for Ent

Hush provides **completely automatic** field-level encryption for Ent schemas. Just annotate your sensitive fields with `hush.EncryptField()` and everything else is handled automatically.

## Features

- 🎯 **100% Automatic** - Just annotate fields, everything else is automatic
- 🔒 **Transparent Encryption** - Seamless encryption on write, decryption on read
- 🔑 **Key Rotation** - Envelope encryption via Google Tink
- 📝 **Single Annotation** - Only `hush.EncryptField()` required
- 🚀 **Zero Configuration** - No manual hooks, interceptors, or mixins
- 🛡️ **Secure by Default** - AES-256-GCM encryption via Google Tink
- 🔄 **Smart Migration** - Automatically handles existing unencrypted data

## Quick Start

### 1. Generate Encryption Key

```bash
go run ./cmd/generate-tink-keyset
export OPENLANE_TINK_KEYSET=<generated-keyset>
```

### 2. Annotate Your Fields

```go
import "github.com/theopenlane/core/internal/ent/hush"

func (MySchema) Fields() []ent.Field {
    return []ent.Field{
        field.String("public_field"),
        field.String("secret_field").
            Sensitive().
            Annotations(hush.EncryptField()), // That's it!
    }
}

func (m MySchema) Hooks() []ent.Hook {
    return hush.AutoEncryptionHook(m)
}

func (m MySchema) Interceptors() []ent.Interceptor {
    return hush.AutoDecryptionInterceptor(m)
}
```

### 3. Use Normally

```go
// Create - automatically encrypted
entity, _ := client.MySchema.Create().
    SetSecretField("encrypted-automatically").
    Save(ctx)

// Query - automatically decrypted
entity, _ = client.MySchema.Get(ctx, id)
fmt.Println(entity.SecretField) // Decrypted value
```

## API Reference

### Annotations

- `hush.EncryptField()` - Mark a field for automatic encryption

### Schema Methods

Required methods in your schema:

```go
func (MySchema) Hooks() []ent.Hook {
    return hush.AutoEncryptionHook(MySchema{})
}

func (MySchema) Interceptors() []ent.Interceptor {
    return hush.AutoDecryptionInterceptor(MySchema{})
}
```

### Direct Encryption

```go
import "github.com/theopenlane/core/internal/ent/hooks"

// Encrypt data
encrypted, err := hooks.Encrypt([]byte("plaintext"))

// Decrypt data
decrypted, err := hooks.Decrypt(encrypted)

// Generate keyset
keyset, err := hooks.GenerateTinkKeyset()
```

## Examples

### OAuth Integration

```go
type OAuthApp struct {
    ent.Schema
}

func (OAuthApp) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
        field.String("client_id"),
        field.String("client_secret").
            Sensitive().
            Annotations(hush.EncryptField()),
        field.String("access_token").
            Optional().
            Sensitive().
            Annotations(hush.EncryptField()),
    }
}

func (o OAuthApp) Hooks() []ent.Hook {
    return hush.AutoEncryptionHook(o)
}

func (o OAuthApp) Interceptors() []ent.Interceptor {
    return hush.AutoDecryptionInterceptor(o)
}
```

### Database Credentials

```go
type Database struct {
    ent.Schema
}

func (Database) Fields() []ent.Field {
    return []ent.Field{
        field.String("host"),
        field.String("username"),
        field.String("password").
            Sensitive().
            Annotations(hush.EncryptField()),
    }
}

func (d Database) Hooks() []ent.Hook {
    return hush.AutoEncryptionHook(d)
}

func (d Database) Interceptors() []ent.Interceptor {
    return hush.AutoDecryptionInterceptor(d)
}
```

## Migration

Migration is **completely automatic**! When you add `hush.EncryptField()` to an existing field:

1. System detects unencrypted values (not base64)
2. Encrypts them transparently during operations
3. Handles mixed encrypted/unencrypted data gracefully
4. No manual steps required

### Adding Encryption to Existing Field

```go
// Before: Plain field
field.String("api_key").Sensitive()

// After: Just add annotation - migration is automatic!
field.String("api_key").
    Sensitive().
    Annotations(hush.EncryptField())
```

## Tools

### Demo Tool

```bash
# Encrypt
go run ./cmd/hush-demo -secret="test"

# Decrypt
go run ./cmd/hush-demo -encrypted="<encrypted>"

# Quiet mode
go run ./cmd/hush-demo -secret="test" -quiet
```

### Keyset Generator

```bash
go run ./cmd/generate-tink-keyset
```

## Configuration

### Environment Variables

```bash
# Required: Tink keyset
OPENLANE_TINK_KEYSET=<base64-keyset>
```

### Production Key Storage

```bash
# AWS Secrets Manager
aws secretsmanager create-secret \
  --name openlane-tink-keyset \
  --secret-string "<keyset>"

# Google Secret Manager
gcloud secrets create openlane-tink-keyset \
  --data-file=keyset.txt

# HashCorp Vault
vault kv put secret/openlane tink_keyset="<keyset>"
```

## Key Rotation

Tink's envelope encryption supports rotation without re-encrypting data:

1. Generate new keyset with rotation tool
2. Update environment with new keyset containing both keys
3. New encryptions use new key, old data remains decryptable
4. Optionally re-encrypt old data over time

## Security

- **Algorithm**: AES-256-GCM with AEAD
- **Library**: Google Tink (production-ready crypto)
- **Key Size**: 256-bit AES keys
- **Nonce**: Unique per encryption (managed by Tink)
- **Storage**: Base64-encoded for database safety

## Why Base64?

Encrypted data contains:
- Null bytes that terminate strings in databases
- Invalid UTF-8 that causes encoding errors
- Control characters that break protocols

Base64 ensures safe text storage in any database.

## Best Practices

### ✅ DO
- Only encrypt sensitive fields (passwords, tokens, keys, secrets)
- Mark encrypted fields as `Sensitive()`
- Use proper key storage in production
- Test migrations on staging first

### ❌ DON'T
- Encrypt searchable fields (IDs, usernames, emails)
- Store keys in environment variables in production
- Commit keys to version control
- Encrypt very large fields without testing

## Troubleshooting

### Field Not Encrypted
- Check `hush.EncryptField()` annotation exists
- Verify `AutoEncryptionHook()` in schema
- Ensure `OPENLANE_TINK_KEYSET` is set

### Decryption Failed
- Verify same keyset used for encryption
- Check base64 encoding is intact
- Ensure data wasn't corrupted

### Performance Issues
- Only encrypt necessary fields
- Consider field size impact
- Monitor encryption overhead

## How It Works

1. **Write Operations**: Hooks intercept mutations and encrypt marked fields
2. **Storage**: Encrypted data is base64-encoded for safe database storage
3. **Read Operations**: Interceptors decrypt fields transparently when queried
4. **Key Management**: Tink handles encryption with envelope encryption for rotation