package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Password hashing functions
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%s|%s",
		base64.RawStdEncoding.EncodeToString(hash),
		base64.RawStdEncoding.EncodeToString(salt)), nil
}

func checkPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "|")
	if len(parts) != 2 {
		return false
	}

	hash, _ := base64.RawStdEncoding.DecodeString(parts[0])
	salt, _ := base64.RawStdEncoding.DecodeString(parts[1])

	newHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, newHash) == 1
}

// UUID generation
func generateUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	// Set version (4) and variant (10)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
