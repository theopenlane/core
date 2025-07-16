package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/theopenlane/core/internal/ent/hooks"
)

const (
	// percentageMultiplier is used for calculating encryption overhead percentage
	percentageMultiplier = 100
	// truncateLength is the number of characters to show in truncated output
	truncateLength = 50
)

func main() {
	var (
		secret    = flag.String("secret", "super-secret-api-key-xyz123", "Secret value to encrypt/decrypt")
		showKey   = flag.Bool("show-key", false, "Show partial encryption key (for demo purposes)")
		quiet     = flag.Bool("quiet", false, "Quiet mode - minimal output")
		encrypted = flag.String("encrypted", "", "Base64 encrypted value to decrypt")
	)
	flag.Parse()

	if *encrypted != "" {
		// Decrypt mode
		decryptValue(*encrypted, *quiet)
		return
	}

	// Demo mode
	if !*quiet {
		fmt.Println("🔐 Hush Encryption System Demo")
		fmt.Println("==============================")
	}

	// 1. Show encryption system info
	if *showKey {
		fmt.Printf("\n📋 Encryption System Info:\n")
		fmt.Printf("   System: Google Tink with envelope encryption\n")
		fmt.Printf("   Algorithm: AES-256-GCM\n")
		fmt.Printf("   Key management: Automatic key rotation support\n")
		fmt.Printf("   Configuration: OPENLANE_TINK_KEYSET environment variable\n")
	}

	// 2. Encrypt the secret
	if !*quiet {
		fmt.Printf("\n🔒 Encrypting Secret:\n")
		fmt.Printf("   Original: %s\n", *secret)
	}

	encryptedValue, err := hooks.Encrypt([]byte(*secret))
	if err != nil {
		log.Fatalf("❌ Encryption failed: %v", err)
	}

	// encryptedValue is already base64 encoded

	if !*quiet {
		fmt.Printf("   Encrypted: %s\n", encryptedValue)
		fmt.Printf("   Size change: %d → %d bytes (%.1f%% increase)\n",
			len(*secret), len(encryptedValue),
			float64(len(encryptedValue)-len(*secret))/float64(len(*secret))*percentageMultiplier)
	} else {
		fmt.Println(encryptedValue)
	}

	// 3. Decrypt to verify
	if !*quiet {
		fmt.Printf("\n🔓 Decrypting to Verify:\n")

		decrypted, err := hooks.Decrypt(encryptedValue)
		if err != nil {
			log.Fatalf("❌ Decryption failed: %v", err)
		}

		fmt.Printf("   Decrypted: %s\n", string(decrypted))
		fmt.Printf("   Match: %t\n", string(decrypted) == *secret)

		// 4. Show security properties
		fmt.Printf("\n🔐 Security Properties:\n")

		// Multiple encryptions produce different ciphertexts
		encrypted2, _ := hooks.Encrypt([]byte(*secret))

		fmt.Printf("   Nonce randomization: %t (same input → different outputs)\n", encryptedValue != encrypted2)
		fmt.Printf("   First encryption:  %s...\n", encryptedValue[:min(truncateLength, len(encryptedValue))])
		fmt.Printf("   Second encryption: %s...\n", encrypted2[:min(truncateLength, len(encrypted2))])

		// Both decrypt to the same value
		decrypted2, _ := hooks.Decrypt(encrypted2)
		fmt.Printf("   Both decrypt correctly: %t\n", string(decrypted) == string(decrypted2))

		// 5. Usage info
		fmt.Printf("\n💡 Usage in Hush Schema:\n")
		fmt.Printf("   • Add hush.EncryptField() annotation to any string field\n")
		fmt.Printf("   • Use NewAutoHushEncryptionMixin() in your schema\n")
		fmt.Printf("   • Fields are automatically encrypted on write, decrypted on read\n")
		fmt.Printf("   • Supports migration from unencrypted to encrypted data\n")

		fmt.Printf("\n🚀 Try it:\n")
		fmt.Printf("   %s -secret=\"your-secret-here\"\n", os.Args[0])
		fmt.Printf("   %s -encrypted=\"%s\"\n", os.Args[0], encryptedValue[:30]+"...")
		fmt.Printf("   %s -quiet -secret=\"just-the-encrypted-output\"\n", os.Args[0])
	}
}

func decryptValue(encryptedValue string, quiet bool) {
	if !quiet {
		fmt.Printf("🔓 Decrypting Value:\n")
		fmt.Printf("   Encrypted: %s\n", encryptedValue)
	}

	// Decrypt using Tink
	decrypted, err := hooks.Decrypt(encryptedValue)
	if err != nil {
		log.Fatalf("❌ Decryption failed: %v", err)
	}

	if !quiet {
		fmt.Printf("   Decrypted: %s\n", string(decrypted))
	} else {
		fmt.Println(string(decrypted))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}