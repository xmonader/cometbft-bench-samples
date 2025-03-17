package storage

import "errors"

// Error definitions for the storage package
var (
	ErrStorageNotFound      = errors.New("storage backend not found")
	ErrAccountNotFound      = errors.New("account not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrTransactionFailed    = errors.New("transaction failed")
	ErrNotInitialized       = errors.New("storage not initialized")
	ErrAlreadyInitialized   = errors.New("storage already initialized")
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrConnectionFailed     = errors.New("connection failed")
)
