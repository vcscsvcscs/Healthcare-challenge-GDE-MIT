package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor handles AES-256 encryption for sensitive health data
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryptor with a 32-byte key for AES-256
// Validates: Requirements 10.1
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes for AES-256, got %d bytes", len(key))
	}

	return &Encryptor{
		key: key,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and append nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptSensitiveFields encrypts sensitive health data fields
func (e *Encryptor) EncryptSensitiveFields(data map[string]string) (map[string]string, error) {
	encrypted := make(map[string]string)

	for key, value := range data {
		encryptedValue, err := e.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt field %s: %w", key, err)
		}
		encrypted[key] = encryptedValue
	}

	return encrypted, nil
}

// DecryptSensitiveFields decrypts sensitive health data fields
func (e *Encryptor) DecryptSensitiveFields(data map[string]string) (map[string]string, error) {
	decrypted := make(map[string]string)

	for key, value := range data {
		decryptedValue, err := e.Decrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt field %s: %w", key, err)
		}
		decrypted[key] = decryptedValue
	}

	return decrypted, nil
}
