package types

import (
	"encoding/json"
	"testing"
)

func TestOperationValidation(t *testing.T) {
	// Valid operation
	op := Operation{
		From:      1,
		To:        2,
		Amount:    50,
		Signature: "test-signature",
	}

	// Create transaction with the operation
	tx := Transaction{
		Operations: []Operation{op},
	}

	// Validate transaction
	if err := tx.Validate(); err != nil {
		t.Errorf("Valid transaction failed validation: %v", err)
	}

	// Test invalid sender
	invalidSenderTx := Transaction{
		Operations: []Operation{{From: 0, To: 2, Amount: 50, Signature: "test-signature"}},
	}
	if err := invalidSenderTx.Validate(); err == nil {
		t.Error("Transaction with invalid sender passed validation")
	}

	// Test invalid recipient
	invalidRecipientTx := Transaction{
		Operations: []Operation{{From: 1, To: 0, Amount: 50, Signature: "test-signature"}},
	}
	if err := invalidRecipientTx.Validate(); err == nil {
		t.Error("Transaction with invalid recipient passed validation")
	}

	// Test invalid amount
	invalidAmountTx := Transaction{
		Operations: []Operation{{From: 1, To: 2, Amount: 0, Signature: "test-signature"}},
	}
	if err := invalidAmountTx.Validate(); err == nil {
		t.Error("Transaction with invalid amount passed validation")
	}

	// Test missing signature
	missingSignatureTx := Transaction{
		Operations: []Operation{{From: 1, To: 2, Amount: 50, Signature: ""}},
	}
	if err := missingSignatureTx.Validate(); err == nil {
		t.Error("Transaction with missing signature passed validation")
	}

	// Test empty operations
	emptyOpsTx := Transaction{
		Operations: []Operation{},
	}
	if err := emptyOpsTx.Validate(); err == nil {
		t.Error("Transaction with empty operations passed validation")
	}
}

func TestTransactionSerialization(t *testing.T) {
	// Create a transaction
	tx := Transaction{
		Operations: []Operation{
			{From: 1, To: 2, Amount: 50, Signature: "test-signature-1"},
			{From: 1, To: 3, Amount: 30, Signature: "test-signature-2"},
		},
	}

	// Serialize
	data, err := tx.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize transaction: %v", err)
	}

	// Deserialize
	parsedTx, err := ParseTransaction(data)
	if err != nil {
		t.Fatalf("Failed to parse transaction: %v", err)
	}

	// Verify fields
	if len(parsedTx.Operations) != len(tx.Operations) {
		t.Errorf("Operation count mismatch: got %d, want %d", len(parsedTx.Operations), len(tx.Operations))
	}

	if parsedTx.Operations[0].From != tx.Operations[0].From {
		t.Errorf("From mismatch: got %d, want %d", parsedTx.Operations[0].From, tx.Operations[0].From)
	}

	if parsedTx.Operations[0].To != tx.Operations[0].To {
		t.Errorf("To mismatch: got %d, want %d", parsedTx.Operations[0].To, tx.Operations[0].To)
	}

	if parsedTx.Operations[0].Amount != tx.Operations[0].Amount {
		t.Errorf("Amount mismatch: got %d, want %d", parsedTx.Operations[0].Amount, tx.Operations[0].Amount)
	}

	if parsedTx.Operations[0].Signature != tx.Operations[0].Signature {
		t.Errorf("Signature mismatch: got %s, want %s", parsedTx.Operations[0].Signature, tx.Operations[0].Signature)
	}
}

func TestGetOperationDataForSigning(t *testing.T) {
	// Create an operation
	op := Operation{
		From:      1,
		To:        2,
		Amount:    50,
		Signature: "test-signature",
	}

	// Get data for signing
	data, err := op.GetDataForSigning()
	if err != nil {
		t.Fatalf("Failed to get operation data for signing: %v", err)
	}

	// Parse the data
	var opForSigning struct {
		From   int `json:"from"`
		To     int `json:"to"`
		Amount int `json:"amount"`
	}
	if err := json.Unmarshal(data, &opForSigning); err != nil {
		t.Fatalf("Failed to parse operation data for signing: %v", err)
	}

	// Verify that the signature is not included
	if opForSigning.From != op.From {
		t.Errorf("From mismatch: got %d, want %d", opForSigning.From, op.From)
	}

	if opForSigning.To != op.To {
		t.Errorf("To mismatch: got %d, want %d", opForSigning.To, op.To)
	}

	if opForSigning.Amount != op.Amount {
		t.Errorf("Amount mismatch: got %d, want %d", opForSigning.Amount, op.Amount)
	}
}

func TestBatchedTransactions(t *testing.T) {
	// Create a transaction with multiple operations
	tx := Transaction{
		Operations: []Operation{
			{From: 1, To: 2, Amount: 50, Signature: "test-signature-1"},
			{From: 1, To: 3, Amount: 30, Signature: "test-signature-2"},
			{From: 1, To: 4, Amount: 20, Signature: "test-signature-3"},
		},
	}

	// Validate transaction
	if err := tx.Validate(); err != nil {
		t.Errorf("Valid batched transaction failed validation: %v", err)
	}

	// Verify operation count
	if len(tx.Operations) != 3 {
		t.Errorf("Operation count mismatch: got %d, want %d", len(tx.Operations), 3)
	}

	// Test with a large batch
	largeOps := make([]Operation, 100)
	for i := 0; i < 100; i++ {
		largeOps[i] = Operation{From: 1, To: 2, Amount: 1, Signature: "test-signature"}
	}

	largeTx := Transaction{
		Operations: largeOps,
	}

	// Validate large transaction
	if err := largeTx.Validate(); err != nil {
		t.Errorf("Valid large batched transaction failed validation: %v", err)
	}

	// Serialize large transaction
	data, err := largeTx.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize large transaction: %v", err)
	}

	// Deserialize large transaction
	parsedTx, err := ParseTransaction(data)
	if err != nil {
		t.Fatalf("Failed to parse large transaction: %v", err)
	}

	// Verify operation count
	if len(parsedTx.Operations) != len(largeOps) {
		t.Errorf("Operation count mismatch: got %d, want %d", len(parsedTx.Operations), len(largeOps))
	}
}
