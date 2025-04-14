package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

// Password hashing functions
// func hashPassword(password string) (string, error) {
// 	salt := make([]byte, 16)
// 	if _, err := rand.Read(salt); err != nil {
// 		return "", fmt.Errorf("failed to generate salt: %w", err)
// 	}

// 	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
// 	return fmt.Sprintf("%s|%s",
// 		base64.RawStdEncoding.EncodeToString(hash),
// 		base64.RawStdEncoding.EncodeToString(salt)), nil
// }

// func checkPassword(password, encoded string) bool {
// 	parts := strings.Split(encoded, "|")
// 	if len(parts) != 2 {
// 		return false
// 	}

// 	hash, _ := base64.RawStdEncoding.DecodeString(parts[0])
// 	salt, _ := base64.RawStdEncoding.DecodeString(parts[1])

//		newHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
//		return subtle.ConstantTimeCompare(hash, newHash) == 1
//	}
//
// Hash a password using bcrypt with cost factor 14
// Allowed alphanumeric characters: a-zA-Z0-9
var alphanum = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// Generate a random alphanumeric string of given length
func generateSalt(length int) (string, error) {
	b := make([]rune, length)
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphanum[int(buf[i])%len(alphanum)]
	}
	return string(b), nil
}

// // Hash a password using SHA256 and alphanumeric salt
// func hashPassword(password string) (string, error) {
// 	salt, err := generateSalt(2)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to generate salt: %w", err)
// 	}
// 	hash := sha256.Sum256(append([]byte(password), salt...))
// 	return fmt.Sprintf("%x$%s", salt, hash), nil
// }

// // Check if a password matches the hashed version
// func checkPassword(password, encoded string) bool {
// 	parts := strings.Split(encoded, "$")
// 	if len(parts) != 2 {
// 		return false
// 	}
// 	salt, storedHashHex := parts[0], parts[1]

// 	storedHash, err := hex.DecodeString(storedHashHex)
// 	if err != nil {
// 		return false
// 	}

//		computedHash := sha256.Sum256(append([]byte(password), salt...))
//		return subtle.ConstantTimeCompare(storedHash, computedHash[:]) == 1
//	}
func hashPassword(password string) (string, error) {
	// Generate a random 2-character salt
	salt, nil := generateSalt(2) // make([]byte, 2)
	// if _, err := rand.Read(salt); err != nil {
	// 	return "", fmt.Errorf("failed to generate salt: %w", err)
	// }

	// Hash the password with the salt
	hash := sha256.Sum256(append([]byte(password), salt...))

	// Return the salt and hash as a combined string
	return fmt.Sprintf("%s%x", salt, hash), nil
}

// Check if a password matches the hashed version
func checkPassword(password, encoded string) bool {
	// Extract the salt (first 2 characters) and the hash
	salt := encoded[:2]
	hashHex := encoded[2:]

	// Decode the hash
	storedHash, err := hex.DecodeString(hashHex)
	if err != nil {
		return false
	}

	// Hash the input password with the extracted salt
	newHash := sha256.Sum256(append([]byte(password), []byte(salt)...))

	// Compare the hashes
	return subtle.ConstantTimeCompare(storedHash, newHash[:]) == 1
}

// UUID generation
// Generate UUID version 7 using the official package
func generateUUID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return id.String()
	}
	// fallback v4
	uuid := make([]byte, 16)

	rand.Read(uuid)
	// Set version (4) and variant (10)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
