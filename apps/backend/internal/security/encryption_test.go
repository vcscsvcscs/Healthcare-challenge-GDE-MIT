package security

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	// Generate a 32-byte key for AES-256
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "Hello, World!",
		},
		{
			name:      "sensitive health data",
			plaintext: "Patient has diabetes and high blood pressure",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "unicode text",
			plaintext: "Fáj a fejem és rossz a közérzetem",
		},
		{
			name:      "long text",
			plaintext: "This is a very long text that contains sensitive health information about a patient's medical history, symptoms, medications, and treatment plans. It should be encrypted properly to ensure GDPR compliance.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := encryptor.Encrypt(tc.plaintext)
			require.NoError(t, err)

			// Empty plaintext should return empty ciphertext
			if tc.plaintext == "" {
				assert.Equal(t, "", ciphertext)
				return
			}

			// Ciphertext should be different from plaintext
			assert.NotEqual(t, tc.plaintext, ciphertext)

			// Ciphertext should not be empty
			assert.NotEmpty(t, ciphertext)

			// Decrypt
			decrypted, err := encryptor.Decrypt(ciphertext)
			require.NoError(t, err)

			// Decrypted text should match original plaintext
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}

func TestEncryptor_InvalidKey(t *testing.T) {
	testCases := []struct {
		name    string
		keySize int
	}{
		{name: "too short", keySize: 16},
		{name: "too long", keySize: 64},
		{name: "empty", keySize: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			_, err := NewEncryptor(key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "encryption key must be 32 bytes")
		})
	}
}

func TestEncryptor_EncryptSensitiveFields(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	data := map[string]string{
		"symptoms":         "headache, fever, cough",
		"general_feeling":  "feeling tired and weak",
		"additional_notes": "patient mentioned family history of diabetes",
	}

	// Encrypt all fields
	encrypted, err := encryptor.EncryptSensitiveFields(data)
	require.NoError(t, err)

	// All fields should be encrypted
	assert.Len(t, encrypted, len(data))
	for key, value := range encrypted {
		assert.NotEqual(t, data[key], value, "field %s should be encrypted", key)
		assert.NotEmpty(t, value, "encrypted field %s should not be empty", key)
	}

	// Decrypt all fields
	decrypted, err := encryptor.DecryptSensitiveFields(encrypted)
	require.NoError(t, err)

	// All fields should match original
	assert.Equal(t, data, decrypted)
}

func TestEncryptor_DifferentCiphertexts(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	plaintext := "sensitive health data"

	// Encrypt the same plaintext multiple times
	ciphertext1, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)

	ciphertext2, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)

	// Ciphertexts should be different due to random nonce
	assert.NotEqual(t, ciphertext1, ciphertext2, "encrypting the same plaintext should produce different ciphertexts")

	// Both should decrypt to the same plaintext
	decrypted1, err := encryptor.Decrypt(ciphertext1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := encryptor.Decrypt(ciphertext2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestEncryptor_InvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	testCases := []struct {
		name       string
		ciphertext string
	}{
		{name: "invalid base64", ciphertext: "not-valid-base64!!!"},
		{name: "too short", ciphertext: "YWJj"},
		{name: "corrupted data", ciphertext: "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo="},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tc.ciphertext)
			assert.Error(t, err)
		})
	}
}
