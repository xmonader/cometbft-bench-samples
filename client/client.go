package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/xmonader/test_batched_tx_tendermint/crypto"
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// Client represents a client for creating and signing transactions
type Client struct {
	keyPair *crypto.KeyPair
	userID  int
}

// NewClient creates a new client with a generated key pair
func NewClient(userID int) *Client {
	keyPair := crypto.GenerateKeyPair()
	return &Client{
		keyPair: keyPair,
		userID:  userID,
	}
}

// LoadClient loads a client from a key file
func LoadClient(keyFile string) (*Client, error) {
	// Read key file
	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse key file
	var keyData struct {
		UserID     int    `json:"user_id"`
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
	}
	if err := json.Unmarshal(data, &keyData); err != nil {
		return nil, fmt.Errorf("failed to parse key file: %w", err)
	}

	// Parse keys
	privKey, err := crypto.PrivateKeyFromBase64(keyData.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	pubKey, err := crypto.PublicKeyFromBase64(keyData.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	return &Client{
		keyPair: &crypto.KeyPair{
			PrivateKey: privKey,
			PublicKey:  pubKey,
		},
		userID: keyData.UserID,
	}, nil
}

// SaveClient saves a client to a key file
func (c *Client) SaveClient(keyFile string) error {
	// Create key data
	keyData := struct {
		UserID     int    `json:"user_id"`
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
	}{
		UserID:     c.userID,
		PrivateKey: crypto.PrivateKeyToBase64(c.keyPair.PrivateKey),
		PublicKey:  crypto.PublicKeyToBase64(c.keyPair.PublicKey),
	}

	// Serialize key data
	data, err := json.MarshalIndent(keyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize key data: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(keyFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write key file
	if err := ioutil.WriteFile(keyFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// GetUserID returns the client's user ID
func (c *Client) GetUserID() int {
	return c.userID
}

// GetPublicKey returns the client's public key
func (c *Client) GetPublicKey() ed25519.PubKey {
	return c.keyPair.PublicKey
}

// GetPublicKeyBase64 returns the client's public key as a base64 string
func (c *Client) GetPublicKeyBase64() string {
	return crypto.PublicKeyToBase64(c.keyPair.PublicKey)
}

// CreateTransaction creates a new transaction with the given operations
func (c *Client) CreateTransaction(operations []types.Operation) (*types.Transaction, error) {
	// Sign each operation
	signedOperations := make([]types.Operation, len(operations))
	for i, op := range operations {
		// Make sure the operation is from this client
		if op.From != c.userID {
			return nil, fmt.Errorf("operation %d has sender %d, but client is for user %d", i, op.From, c.userID)
		}

		// Get data to sign
		dataToSign, err := op.GetDataForSigning()
		if err != nil {
			return nil, fmt.Errorf("failed to get operation %d data for signing: %w", i, err)
		}

		// Sign operation
		signature, err := c.keyPair.Sign(dataToSign)
		if err != nil {
			return nil, fmt.Errorf("failed to sign operation %d: %w", i, err)
		}

		// Create signed operation
		signedOp := op
		signedOp.Signature = signature
		signedOperations[i] = signedOp
	}

	// Create transaction with signed operations
	tx := &types.Transaction{
		Operations: signedOperations,
	}

	return tx, nil
}

// CreateTransferOperation creates a new transfer operation (unsigned)
func (c *Client) CreateTransferOperation(to int, amount int) types.Operation {
	return types.Operation{
		From:   c.userID,
		To:     to,
		Amount: amount,
	}
}

// SignOperation signs an operation
func (c *Client) SignOperation(op types.Operation) (types.Operation, error) {
	// Make sure the operation is from this client
	if op.From != c.userID {
		return op, fmt.Errorf("operation has sender %d, but client is for user %d", op.From, c.userID)
	}

	// Get data to sign
	dataToSign, err := op.GetDataForSigning()
	if err != nil {
		return op, fmt.Errorf("failed to get operation data for signing: %w", err)
	}

	// Sign operation
	signature, err := c.keyPair.Sign(dataToSign)
	if err != nil {
		return op, fmt.Errorf("failed to sign operation: %w", err)
	}

	// Create signed operation
	signedOp := op
	signedOp.Signature = signature
	return signedOp, nil
}

// CreateBatchedTransferOperations creates a batch of transfer operations
func (c *Client) CreateBatchedTransferOperations(recipients []int, amounts []int) ([]types.Operation, error) {
	if len(recipients) != len(amounts) {
		return nil, fmt.Errorf("recipients and amounts must have the same length")
	}

	operations := make([]types.Operation, len(recipients))
	for i := 0; i < len(recipients); i++ {
		operations[i] = c.CreateTransferOperation(recipients[i], amounts[i])
	}

	return operations, nil
}
