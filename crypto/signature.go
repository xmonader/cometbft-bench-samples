package crypto

import (
	"encoding/base64"
	"fmt"

	"github.com/cometbft/cometbft/crypto/ed25519"
)

// KeyPair represents an Ed25519 key pair
type KeyPair struct {
	PrivateKey ed25519.PrivKey
	PublicKey  ed25519.PubKey
}

// GenerateKeyPair generates a new Ed25519 key pair
func GenerateKeyPair() *KeyPair {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey().(ed25519.PubKey)

	return &KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
}

// Sign signs a message with the private key
func (kp *KeyPair) Sign(message []byte) (string, error) {
	signature, err := kp.PrivateKey.Sign(message)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies a signature against a message and public key
func VerifySignature(pubKey ed25519.PubKey, message []byte, signatureBase64 string) (bool, error) {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	return pubKey.VerifySignature(message, signature), nil
}

// PublicKeyFromBytes creates a public key from bytes
func PublicKeyFromBytes(data []byte) (ed25519.PubKey, error) {
	// Ed25519 public key size is 32 bytes
	if len(data) != 32 {
		return nil, fmt.Errorf("invalid public key size: got %d, want %d", len(data), 32)
	}
	return ed25519.PubKey(data), nil
}

// PrivateKeyFromBytes creates a private key from bytes
func PrivateKeyFromBytes(data []byte) (ed25519.PrivKey, error) {
	// Ed25519 private key size is 64 bytes
	if len(data) != 64 {
		return nil, fmt.Errorf("invalid private key size: got %d, want %d", len(data), 64)
	}
	return ed25519.PrivKey(data), nil
}

// PublicKeyToBase64 encodes a public key to base64
func PublicKeyToBase64(pubKey ed25519.PubKey) string {
	return base64.StdEncoding.EncodeToString(pubKey)
}

// PrivateKeyToBase64 encodes a private key to base64
func PrivateKeyToBase64(privKey ed25519.PrivKey) string {
	return base64.StdEncoding.EncodeToString(privKey)
}

// PublicKeyFromBase64 decodes a public key from base64
func PublicKeyFromBase64(encoded string) (ed25519.PubKey, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	return PublicKeyFromBytes(data)
}

// PrivateKeyFromBase64 decodes a private key from base64
func PrivateKeyFromBase64(encoded string) (ed25519.PrivKey, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	return PrivateKeyFromBytes(data)
}
