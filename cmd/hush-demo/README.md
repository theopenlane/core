# Hush Encryption Demo

Interactive command-line tool to demonstrate the Hush field-level encryption system.

## Features

- **Interactive Demo**: Shows step-by-step encryption/decryption process
- **Encrypt Mode**: Encrypt any secret and get base64 output
- **Decrypt Mode**: Decrypt base64 encrypted values
- **Quiet Mode**: Script-friendly output for automation
- **Security Analysis**: Shows nonce randomization and other security properties

## Usage

### Basic Demo
```bash
./hush-demo
```

### Encrypt Custom Secret
```bash
./hush-demo -secret="your-secret-here"
```

### Decrypt Encrypted Value
```bash
./hush-demo -encrypted="base64-encrypted-value"
```

### Show Encryption Key Details
```bash
./hush-demo -show-key
```

### Quiet Mode (for scripts)
```bash
# Get just the encrypted output
./hush-demo -quiet -secret="my-secret"

# Get just the decrypted output
./hush-demo -quiet -encrypted="encrypted-value"
```

## Options

- `-secret string`: Secret value to encrypt (default: "super-secret-api-key-xyz123")
- `-encrypted string`: Base64 encrypted value to decrypt
- `-show-key`: Show partial encryption key information
- `-quiet`: Minimal output for scripting

## Examples

### Development/Testing
```bash
# Test encryption with different secrets
./hush-demo -secret="database-password-123"
./hush-demo -secret="github-token-xyz"

# Verify decryption works
ENCRYPTED=$(./hush-demo -quiet -secret="test")
./hush-demo -encrypted="$ENCRYPTED"
```

### Integration with Scripts
```bash
#!/bin/bash
# Encrypt secrets for database seeding
DB_PASSWORD_ENCRYPTED=$(./hush-demo -quiet -secret="$DB_PASSWORD")
API_KEY_ENCRYPTED=$(./hush-demo -quiet -secret="$API_KEY")

# Store in config or environment
echo "DB_PASSWORD_ENCRYPTED=$DB_PASSWORD_ENCRYPTED" >> .env
echo "API_KEY_ENCRYPTED=$API_KEY_ENCRYPTED" >> .env
```

### Understanding the System
```bash
# See how encryption keys work
./hush-demo -show-key -secret="example"

# Verify nonce randomization (same input → different outputs)
./hush-demo -secret="same-secret"
./hush-demo -secret="same-secret"
```

## Security Properties Demonstrated

1. **AES-256-GCM Encryption**: Industry-standard authenticated encryption
2. **Nonce Randomization**: Same plaintext produces different ciphertexts
3. **Base64 Encoding**: Safe for database storage and text transmission
4. **Key Management**: Uses the same key derivation as the main system
5. **Authenticated Encryption**: Tampering detection built-in

## Why Base64 Encoding?

The demo shows base64 encoding/decoding because that's exactly what happens in production:

```
Plaintext → AES Encryption → Raw Binary → Base64 → Database Storage
Database → Base64 → Raw Binary → AES Decryption → Plaintext
```

**Why not store raw binary directly?**
- Raw AES output contains null bytes and invalid UTF-8 sequences
- Database TEXT/VARCHAR fields expect valid UTF-8 strings
- Storing binary data directly would cause corruption or database errors

**Base64 makes it database-safe:**
- Converts binary data to ASCII text (A-Z, a-z, 0-9, +, /, =)
- No null bytes or encoding issues
- Safe for PostgreSQL VARCHAR, MySQL TEXT, etc.
- Only 33% size increase vs 100% for hex encoding

The base64 step is an **implementation detail** for safe storage, not part of the encryption itself.

## Building

```bash
go build -o bin/hush-demo cmd/hush-demo/main.go
```

## Related Documentation

- [Hush Package](../../internal/ent/hush/) - Main encryption implementation
- [Field Encryption Guide](../../internal/ent/hush/examples/) - Usage examples
- [Schema Integration](../../internal/ent/schema/) - How to use in Ent schemas