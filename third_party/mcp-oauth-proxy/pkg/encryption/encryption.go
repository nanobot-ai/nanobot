package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// EncryptedProps represents encrypted properties with metadata
type EncryptedProps struct {
	Data      string `json:"data"`      // Base64 encoded encrypted data
	IV        string `json:"iv"`        // Base64 encoded initialization vector
	Algorithm string `json:"algorithm"` // Encryption algorithm used
}

// EncryptData encrypts sensitive data using AES-256-GCM
func EncryptData(data map[string]any, encryptionKey []byte) (*EncryptedProps, error) {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random IV
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, iv, jsonData, nil)

	return &EncryptedProps{
		Data:      base64.StdEncoding.EncodeToString(ciphertext),
		IV:        base64.StdEncoding.EncodeToString(iv),
		Algorithm: "AES-256-GCM",
	}, nil
}

// DecryptData decrypts encrypted data using AES-256-GCM
func DecryptData(encryptedProps *EncryptedProps, encryptionKey []byte) (map[string]any, error) {
	// Decode base64 data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedProps.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(encryptedProps.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt the data
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Unmarshal JSON data
	var data map[string]any
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted data: %w", err)
	}

	return data, nil
}

// DecryptPropsIfNeeded decrypts props data if it's encrypted, otherwise returns the original data
func DecryptPropsIfNeeded(encryptionKey []byte, props map[string]any) (map[string]any, error) {
	// Check if data is encrypted
	encrypted, ok := props["encrypted"].(bool)
	if !ok || !encrypted {
		// Data is not encrypted, return as is
		return props, nil
	}

	// Extract encrypted data
	encryptedData, ok := props["encrypted_data"].(string)
	if !ok {
		return nil, fmt.Errorf("encrypted_data field not found")
	}

	iv, ok := props["iv"].(string)
	if !ok {
		return nil, fmt.Errorf("iv field not found")
	}

	algorithm, ok := props["algorithm"].(string)
	if !ok {
		return nil, fmt.Errorf("algorithm field not found")
	}

	// Create EncryptedProps struct
	encryptedProps := &EncryptedProps{
		Data:      encryptedData,
		IV:        iv,
		Algorithm: algorithm,
	}

	// Decrypt the data
	decryptedData, err := DecryptData(encryptedProps, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Merge decrypted data with non-sensitive props
	result := make(map[string]any)
	for key, value := range props {
		if key != "encrypted_data" && key != "iv" && key != "algorithm" && key != "encrypted" {
			result[key] = value
		}
	}
	for key, value := range decryptedData {
		result[key] = value
	}

	return result, nil
}

// EncryptString encrypts a string using AES-256-GCM
func EncryptString(encryptionKey []byte, plaintext string) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random IV
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)

	// Combine IV and ciphertext, then base64 encode
	combined := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptString decrypts a string using AES-256-GCM
func DecryptString(encryptionKey []byte, encryptedData string) (string, error) {
	// Decode base64 data
	combined, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract IV and ciphertext
	ivSize := gcm.NonceSize()
	if len(combined) < ivSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	iv := combined[:ivSize]
	ciphertext := combined[ivSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintext), nil
}

// EncryptCookie encrypts a cookie value using AES-256-GCM
// Returns a base64-encoded encrypted string suitable for cookie storage
func EncryptCookie(encryptionKey []byte, value string) (string, error) {
	return EncryptString(encryptionKey, value)
}

// DecryptCookie decrypts an encrypted cookie value using AES-256-GCM
// Takes a base64-encoded encrypted string and returns the plaintext
func DecryptCookie(encryptionKey []byte, encryptedValue string) (string, error) {
	return DecryptString(encryptionKey, encryptedValue)
}
