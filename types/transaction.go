package types

import (
	"encoding/json"
	"fmt"
)

// Operation represents a single token transfer operation with its own signature
type Operation struct {
	From      int    `json:"from"`
	To        int    `json:"to"`
	Amount    int    `json:"amount"`
	Signature string `json:"signature"`
}

// Transaction represents a batch of operations, each with its own signature
type Transaction struct {
	Operations []Operation `json:"operations"`
}

// ParseTransaction parses a JSON transaction
func ParseTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	err := json.Unmarshal(data, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}
	return &tx, nil
}

// Serialize serializes a transaction to JSON
func (tx *Transaction) Serialize() ([]byte, error) {
	return json.Marshal(tx)
}

// GetOperationForSigning returns the operation data that should be signed
func (op *Operation) GetDataForSigning() ([]byte, error) {
	// For signing, we only include the operation details, not the signature itself
	opForSigning := struct {
		From   int `json:"from"`
		To     int `json:"to"`
		Amount int `json:"amount"`
	}{
		From:   op.From,
		To:     op.To,
		Amount: op.Amount,
	}
	return json.Marshal(opForSigning)
}

// String returns a string representation of the transaction
func (tx *Transaction) String() string {
	data, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling transaction: %v", err)
	}
	return string(data)
}

// Validate performs basic validation on the transaction
func (tx *Transaction) Validate() error {
	if len(tx.Operations) == 0 {
		return fmt.Errorf("transaction must contain at least one operation")
	}

	for i, op := range tx.Operations {
		if op.From <= 0 {
			return fmt.Errorf("operation %d: invalid sender ID %d", i, op.From)
		}
		if op.To <= 0 {
			return fmt.Errorf("operation %d: invalid recipient ID %d", i, op.To)
		}
		if op.Amount <= 0 {
			return fmt.Errorf("operation %d: invalid amount %d", i, op.Amount)
		}
		if op.Signature == "" {
			return fmt.Errorf("operation %d: missing signature", i)
		}
	}

	return nil
}

// ValidateOperation performs basic validation on a single operation
func ValidateOperation(op *Operation) error {
	if op.From <= 0 {
		return fmt.Errorf("invalid sender ID %d", op.From)
	}
	if op.To <= 0 {
		return fmt.Errorf("invalid recipient ID %d", op.To)
	}
	if op.Amount <= 0 {
		return fmt.Errorf("invalid amount %d", op.Amount)
	}
	if op.Signature == "" {
		return fmt.Errorf("missing signature")
	}
	return nil
}
