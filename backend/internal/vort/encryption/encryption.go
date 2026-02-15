// =============================================================================
// Package: encryption
// File: encryption.go
// Purpose: Encryption protocols for secure agent-backend communication
// Created: 2026-02-15
// =============================================================================
// This package provides encryption capabilities for securing data transmission
// between agents and the backend. It implements AES-256-GCM encryption with
// key rotation support.
// =============================================================================

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// =============================================================================
// Errors
// Purpose: Error definitions for encryption operations
// =============================================================================
var (
	ErrInvalidKeyLength = errors.New("invalid key length")
	ErrInvalidCipher   = errors.New("invalid ciphertext")
	ErrDecryptionFailed = errors.New("decryption failed")
)

// =============================================================================
// Type: Encryptor
// Purpose: Thread-safe AES-256-GCM encryption implementation
// =============================================================================
// Encryptor provides symmetric encryption using AES-256 in GCM mode.
// GCM (Galois/Counter Mode) provides both confidentiality and authenticity.
//
// Thread Safety:
//   - Uses sync.RWMutex for concurrent read access
//   - Safe for concurrent use by multiple goroutines
type Encryptor struct {
	mu        sync.RWMutex
	key       []byte       // 32-byte AES-256 key
	gcm       cipher.AEAD  // GCM cipher instance
	algorithm string       // Algorithm identifier
}

// =============================================================================
// Type: KeyManager
// Purpose: Centralized encryption key management
// =============================================================================
// KeyManager handles encryption key storage, rotation, and retrieval.
// It maintains a thread-safe in-memory cache of encryption keys.
//
// Usage:
//   km := encryption.NewKeyManager(30) // 30-day rotation
//   key, _ := km.GenerateKey("agent-123")
//   encryptor, _ := km.GetEncryptor("agent-123")
type KeyManager struct {
	mu           sync.RWMutex
	keys         map[string]*Encryptor // Agent ID -> Encryptor mapping
	rotationDays int                   // Automatic rotation interval
	keySize      int                  // Key size in bytes (32 for AES-256)
}

// NewKeyManager creates a new key manager with specified rotation interval.
//
// Parameters:
//   - rotationDays: Days between automatic key rotations
//
// Returns:
//   - *KeyManager: New key manager instance
func NewKeyManager(rotationDays int) *KeyManager {
	return &KeyManager{
		keys:         make(map[string]*Encryptor),
		rotationDays: rotationDays,
		keySize:      32,
	}
}

// GenerateKey generates a new AES-256 encryption key for an agent.
//
// Parameters:
//   - keyID: Unique identifier for the key (typically agent ID)
//
// Returns:
//   - []byte: Raw encryption key (should be stored securely)
//   - error: Any generation error
//
// Note: The returned key should be transmitted to the agent securely
func (km *KeyManager) GenerateKey(keyID string) ([]byte, error) {
	// Generate cryptographically secure random key
	key := make([]byte, km.keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Create encryptor instance
	encryptor, err := NewEncryptor(key, "AES-256-GCM")
	if err != nil {
		return nil, err
	}

	// Store in key cache
	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys[keyID] = encryptor

	return key, nil
}

// SetKey stores an externally-generated encryption key.
//
// Parameters:
//   - keyID: Unique identifier for the key
//   - key: Raw encryption key bytes
//
// Returns:
//   - error: Any key validation error
func (km *KeyManager) SetKey(keyID string, key []byte) error {
	encryptor, err := NewEncryptor(key, "AES-256-GCM")
	if err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys[keyID] = encryptor

	return nil
}

// GetEncryptor retrieves the encryption instance for a key.
//
// Parameters:
//   - keyID: Key identifier to retrieve
//
// Returns:
//   - *Encryptor: Encryption instance
//   - error: If key not found
func (km *KeyManager) GetEncryptor(keyID string) (*Encryptor, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	encryptor, exists := km.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	return encryptor, nil
}

// RotateKey generates a new key for an existing key ID.
// This implements key rotation for security compliance.
//
// Parameters:
//   - keyID: Key identifier to rotate
//
// Returns:
//   - []byte: New encryption key
//   - error: Any rotation error
func (km *KeyManager) RotateKey(keyID string) ([]byte, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Generate new key
	newKey := make([]byte, km.keySize)
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return nil, fmt.Errorf("failed to generate new key: %w", err)
	}

	encryptor, err := NewEncryptor(newKey, "AES-256-GCM")
	if err != nil {
		return nil, err
	}

	km.keys[keyID] = encryptor
	return newKey, nil
}

// DeleteKey removes a key from the manager.
//
// Parameters:
//   - keyID: Key identifier to remove
func (km *KeyManager) DeleteKey(keyID string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	delete(km.keys, keyID)
}

// =============================================================================
// Type: Encryptor Constructor
// Purpose: Create new AES-256-GCM encryptor
// =============================================================================
// NewEncryptor creates an AES-256-GCM encryptor instance.
//
// Parameters:
//   - key: 32-byte encryption key
//   - algorithm: Algorithm identifier (currently only "AES-256-GCM")
//
// Returns:
//   - *Encryptor: Configured encryptor
//   - error: Key or algorithm error
func NewEncryptor(key []byte, algorithm string) (*Encryptor, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Encryptor{
		key:       key,
		gcm:       gcm,
		algorithm: algorithm,
	}, nil
}

// =============================================================================
// Method: Encrypt
// Purpose: Encrypt plaintext using AES-256-GCM
// =============================================================================
// Encrypt encrypts the given plaintext using AES-256-GCM.
// The output includes a random nonce prepended to the ciphertext.
//
// Parameters:
//   - plaintext: Data to encrypt
//
// Returns:
//   - []byte: Encrypted data (nonce + ciphertext + auth tag)
//   - error: Encryption error
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	// Generate random nonce
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt with GCM (includes authentication tag)
	ciphertext := e.gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// =============================================================================
// Method: Decrypt
// Purpose: Decrypt ciphertext using AES-256-GCM
// =============================================================================
// Decrypt decrypts data encrypted with Encrypt().
// Returns error if authentication fails (tampering detection).
//
// Parameters:
//   - ciphertext: Encrypted data (nonce + ciphertext + auth tag)
//
// Returns:
//   - []byte: Decrypted plaintext
//   - error: Decryption or authentication error
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := e.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCipher
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and authenticate
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// =============================================================================
// Method: EncryptString
// Purpose: Convenience method for string encryption
// =============================================================================
// EncryptString encrypts a string and returns base64-encoded result.
//
// Parameters:
//   - plaintext: String to encrypt
//
// Returns:
//   - string: Base64-encoded encrypted data
//   - error: Encryption error
func (e *Encryptor) EncryptString(plaintext string) (string, error) {
	ciphertext, err := e.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// =============================================================================
// Method: DecryptString
// Purpose: Convenience method for string decryption
// =============================================================================
// DecryptString decrypts a base64-encoded encrypted string.
//
// Parameters:
//   - ciphertextBase64: Base64-encoded encrypted data
//
// Returns:
//   - string: Decrypted string
//   - error: Decryption error
func (e *Encryptor) DecryptString(ciphertextBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// =============================================================================
// Function: HashKey
// Purpose: Create a deterministic hash of a key for display/logging
// =============================================================================
// HashKey creates a SHA-256 hash of a key for safe logging and display
// without exposing the actual key material.
//
// Parameters:
//   - key: Raw key bytes
//
// Returns:
//   - string: Base64-encoded key hash
func HashKey(key []byte) string {
	hash := sha256.Sum256(key)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// =============================================================================
// Type: EncryptedMessage
// Purpose: Structured encrypted message format
// =============================================================================
// EncryptedMessage wraps encrypted data with metadata for secure transmission.
type EncryptedMessage struct {
	KeyID      string    `json:"key_id"`      // Key identifier for decryption
	Ciphertext []byte   `json:"ciphertext"` // Encrypted data
	Timestamp  time.Time `json:"timestamp"`  // Encryption timestamp
	Nonce      []byte    `json:"nonce,omitempty"` // Nonce (if separate)
}

// EncryptMessage encrypts a struct/interface to an EncryptedMessage.
//
// Parameters:
//   - data: Serializable data to encrypt
//
// Returns:
//   - *EncryptedMessage: Encrypted message with metadata
//   - error: Serialization or encryption error
func (e *Encryptor) EncryptMessage(data interface{}) (*EncryptedMessage, error) {
	// Serialize data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Encrypt
	ciphertext, err := e.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}

	return &EncryptedMessage{
		Ciphertext: ciphertext,
		Timestamp:  time.Now(),
	}, nil
}

// DecryptMessage decrypts an EncryptedMessage.
//
// Parameters:
//   - msg: Encrypted message
//
// Returns:
//   - []byte: Decrypted data
//   - error: Decryption error
func (e *Encryptor) DecryptMessage(msg *EncryptedMessage) ([]byte, error) {
	return e.Decrypt(msg.Ciphertext)
}

// =============================================================================
// Type: KeyRotation
// Purpose: Automatic key rotation manager
// =============================================================================
// KeyRotation handles periodic encryption key rotation for security compliance.
// It automatically rotates keys at configured intervals.
//
// TODO: Persist key history for decryption of old data
// TODO: Add webhook notification for rotation events
type KeyRotation struct {
	keyManager    *KeyManager
	rotationTimer *time.Ticker
	stopCh        chan struct{}
}

// NewKeyRotation creates a new key rotation manager.
//
// Parameters:
//   - km: KeyManager to rotate
//   - interval: Rotation interval
//
// Returns:
//   - *KeyRotation: Rotation manager
func NewKeyRotation(km *KeyManager, interval time.Duration) *KeyRotation {
	return &KeyRotation{
		keyManager:    km,
		rotationTimer: time.NewTicker(interval),
		stopCh:       make(chan struct{}),
	}
}

// Start begins automatic key rotation in background goroutine.
func (kr *KeyRotation) Start() {
	go func() {
		for {
			select {
			case <-kr.rotationTimer.C:
				kr.rotateAllKeys()
			case <-kr.stopCh:
				return
			}
		}
	}()
}

// Stop halts automatic key rotation.
func (kr *KeyRotation) Stop() {
	kr.rotationTimer.Stop()
	close(kr.stopCh)
}

// rotateAllKeys rotates keys for all known agents.
func (kr *KeyRotation) rotateAllKeys() {
	kr.keyManager.mu.RLock()
	keyIDs := make([]string, 0, len(kr.keyManager.keys))
	for keyID := range kr.keyManager.keys {
		keyIDs = append(keyIDs, keyID)
	}
	kr.keyManager.mu.RUnlock()

	for _, keyID := range keyIDs {
		if newKey, err := kr.keyManager.RotateKey(keyID); err == nil {
			fmt.Printf("Rotated key %s, new key: %s\n", keyID, HashKey(newKey))
		}
	}
}
