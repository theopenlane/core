# Key Rotation with Envelope Encryption

## Overview

Your hush encryption system uses **envelope encryption** via Google Tink, which enables key rotation without re-encrypting existing data. This document explains how it works and how to rotate keys safely.

## How Envelope Encryption Works

### Traditional Encryption Problem
```
Data → [Encrypt with Master Key] → Encrypted Data
```
**Problem**: To rotate the master key, you must decrypt and re-encrypt ALL data.

### Envelope Encryption Solution
```
Data → [Encrypt with DEK] → Encrypted Data
DEK  → [Encrypt with Master Key] → Encrypted DEK

Final Storage: [Encrypted DEK + Encrypted Data]
```
**Benefit**: Only the small DEK needs to be decrypted with the master key, not all your data.

### Your Implementation

1. **Master Key**: Stored in `OPENLANE_TINK_KEYSET` environment variable
2. **Data Encryption Keys (DEKs)**: Generated automatically by Tink for each encryption operation
3. **Keyset**: Contains multiple master keys, with one designated as "primary"

When you call `Encrypt()`:
- Tink generates a random DEK for this specific data
- Encrypts your data with the DEK using AES-256-GCM
- Encrypts the DEK with the current primary master key
- Returns: `[Encrypted DEK][Encrypted Data]` as a single base64 string

When you call `Decrypt()`:
- Extracts the encrypted DEK from the start of the ciphertext
- Tries each master key in the keyset until one successfully decrypts the DEK
- Uses the decrypted DEK to decrypt the actual data

## Key Rotation Process

### 1. Generate Initial Keyset
```bash
# Generate your first keyset
go run internal/ent/hush/cmd/hush/main.go generate

# Set it in your environment
export OPENLANE_TINK_KEYSET='<generated-keyset>'
```

### 2. Rotate Keys (Add New Primary)
```bash
# Current keyset info
go run internal/ent/hush/cmd/hush/main.go info "$OPENLANE_TINK_KEYSET"

# Rotate: adds new primary key, keeps old keys for decryption
NEW_KEYSET=$(go run internal/ent/hush/cmd/hush/main.go rotate "$OPENLANE_TINK_KEYSET" | grep -A1 "New keyset:" | tail -1)

# Update environment variable
export OPENLANE_TINK_KEYSET="$NEW_KEYSET"

# Restart your application to use the new keyset
```

### 3. Verification
```bash
# Check that old encrypted data still works
# (your application tests should verify this)

# Check keyset info
go run internal/ent/hush/cmd/hush/main.go info "$OPENLANE_TINK_KEYSET"
```

## Advanced Key Management

### Add Keys Without Changing Primary
```bash
# Useful for preparing for rotation
go run internal/ent/hush/cmd/hush/main.go add "$OPENLANE_TINK_KEYSET"
```

### Disable Old Keys
```bash
# Keep only the 2 most recent keys active
# (Older keys remain for decryption but won't be used for new encryption)
go run internal/ent/hush/cmd/hush/main.go disable "$OPENLANE_TINK_KEYSET" 2
```

### Key Information

```bash
# View detailed keyset information
go run internal/ent/hush/cmd/hush/main.go info "$OPENLANE_TINK_KEYSET"
```

Example output:
```json
{
  "primary_key_id": 3367142278,
  "total_keys": 3,
  "keys": [
    {
      "key_id": 3199333218,
      "status": "ENABLED",
      "key_type": "type.googleapis.com/google.crypto.tink.AesGcmKey",
      "is_primary": false
    },
    {
      "key_id": 3367142278,
      "status": "ENABLED",
      "key_type": "type.googleapis.com/google.crypto.tink.AesGcmKey",
      "is_primary": true
    },
    {
      "key_id": 2847593021,
      "status": "DISABLED",
      "key_type": "type.googleapis.com/google.crypto.tink.AesGcmKey",
      "is_primary": false
    }
  ]
}
```

## Programmatic Key Rotation

You can also rotate keys programmatically in your application:

```go
package main

import (
    "log"
    "os"

    "github.com/theopenlane/ent/hush/crypto"
)

func rotateKeys() error {
    // Get current keyset
    currentKeyset := os.Getenv("OPENLANE_TINK_KEYSET")
    if currentKeyset == "" {
        return fmt.Errorf("no keyset found in environment")
    }

    // Rotate the keyset
    newKeyset, err := crypto.RotateKeyset(currentKeyset)
    if err != nil {
        return fmt.Errorf("failed to rotate keyset: %w", err)
    }

    // Update environment (in production, update your config system)
    os.Setenv("OPENLANE_TINK_KEYSET", newKeyset)

    // Force reload of the encryption system
    err = crypto.ReloadTinkAEAD()
    if err != nil {
        return fmt.Errorf("failed to reload encryption system: %w", err)
    }

    log.Println("Key rotation completed successfully")
    return nil
}
```

### Example Automation Script

```bash
#!/bin/bash
# rotate-keys.sh - Automated key rotation script

set -e

echo "Starting automated key rotation..."

# Get current keyset
CURRENT_KEYSET="$OPENLANE_TINK_KEYSET"

# Rotate keys
NEW_KEYSET=$(./rotate-keys rotate "$CURRENT_KEYSET" | grep "export OPENLANE_TINK_KEYSET" | cut -d"'" -f2)

# Update your configuration system (example: Kubernetes secret)
kubectl create secret generic tink-keyset \
    --from-literal=keyset="$NEW_KEYSET" \
    --dry-run=client -o yaml | kubectl apply -f -

# Restart deployments to pick up new keyset
kubectl rollout restart deployment/your-app

echo "Key rotation completed. Waiting for deployment..."
kubectl rollout status deployment/your-app

echo "Verifying old data is still decryptable..."
# Add your verification tests here

echo "✅ Key rotation successful!"
```
