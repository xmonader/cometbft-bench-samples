package app

import (
	"fmt"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/xmonader/test_batched_tx_tendermint/crypto"
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// TransactionProcessor processes transactions
type TransactionProcessor struct {
	stateStore *StateStore
	// Map of user ID to public key for signature verification
	// In a real application, this would be more sophisticated
	userKeys map[int]ed25519.PubKey
}

// NewTransactionProcessor creates a new transaction processor
func NewTransactionProcessor(stateStore *StateStore) *TransactionProcessor {
	return &TransactionProcessor{
		stateStore: stateStore,
		userKeys:   make(map[int]ed25519.PubKey),
	}
}

// RegisterUserKey registers a public key for a user
func (tp *TransactionProcessor) RegisterUserKey(userID int, pubKey ed25519.PubKey) {
	tp.userKeys[userID] = pubKey
}

// GetUserKey gets a user's public key
func (tp *TransactionProcessor) GetUserKey(userID int) (ed25519.PubKey, bool) {
	key, exists := tp.userKeys[userID]
	return key, exists
}

// ValidateTransaction validates a transaction
func (tp *TransactionProcessor) ValidateTransaction(tx *types.Transaction) error {
	// Basic validation
	if err := tx.Validate(); err != nil {
		return err
	}

	// Validate each operation individually
	for i, op := range tx.Operations {
		if err := tp.ValidateOperation(&op); err != nil {
			return fmt.Errorf("operation %d: %w", i, err)
		}
	}

	return nil
}

// ValidateOperation validates a single operation
func (tp *TransactionProcessor) ValidateOperation(op *types.Operation) error {
	// Basic validation
	if err := types.ValidateOperation(op); err != nil {
		return err
	}

	// Get the sender's public key
	pubKey, exists := tp.GetUserKey(op.From)
	if !exists {
		return fmt.Errorf("no public key registered for user %d", op.From)
	}

	// Get the data that was signed
	dataToVerify, err := op.GetDataForSigning()
	if err != nil {
		return fmt.Errorf("failed to get operation data for signing: %w", err)
	}

	// Verify the signature
	valid, err := crypto.VerifySignature(pubKey, dataToVerify, op.Signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid signature")
	}

	// Check if sender has sufficient balance
	account := tp.stateStore.GetAccount(op.From)
	if account.Balance < op.Amount {
		return fmt.Errorf("insufficient balance: %d < %d", account.Balance, op.Amount)
	}

	return nil
}

// ProcessTransaction processes a transaction
func (tp *TransactionProcessor) ProcessTransaction(tx *types.Transaction) error {
	// Validate the transaction
	if err := tp.ValidateTransaction(tx); err != nil {
		return err
	}

	// Process all operations
	for i, op := range tx.Operations {
		// Deduct from sender
		if err := tp.stateStore.UpdateBalance(op.From, -op.Amount); err != nil {
			return fmt.Errorf("failed to deduct from sender in operation %d: %w", i, err)
		}

		// Add to recipient
		if err := tp.stateStore.UpdateBalance(op.To, op.Amount); err != nil {
			// This should never happen, but if it does, we need to revert the sender's deduction
			// In a real application, you'd use a proper transaction mechanism
			_ = tp.stateStore.UpdateBalance(op.From, op.Amount)
			return fmt.Errorf("failed to add to recipient in operation %d: %w", i, err)
		}
	}

	return nil
}
