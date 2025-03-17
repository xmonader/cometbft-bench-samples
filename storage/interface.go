package storage

import (
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// Storage defines the interface for all storage backends
type Storage interface {
	// Initialize initializes the storage backend
	Initialize() error

	// Close closes the storage backend
	Close() error

	// GetAccount retrieves an account by ID
	GetAccount(id int) (*types.Account, error)

	// AccountExists checks if an account exists
	AccountExists(id int) (bool, error)

	// UpdateBalance updates an account's balance
	UpdateBalance(id int, delta int) error

	// CreateAccount creates a new account
	CreateAccount(id int, initialBalance int) error

	// Commit commits any pending changes
	Commit() error

	// Rollback rolls back any pending changes
	Rollback() error

	// BeginTransaction begins a new transaction
	BeginTransaction() error

	// GetBalance gets an account's current balance
	GetBalance(id int) (int, error)

	// GetAllAccounts gets all accounts
	GetAllAccounts() ([]*types.Account, error)
}

// StorageFactory is a function that creates a new storage instance
type StorageFactory func(config map[string]interface{}) (Storage, error)

// StorageRegistry keeps track of available storage backends
var StorageRegistry = make(map[string]StorageFactory)

// RegisterStorage registers a storage backend
func RegisterStorage(name string, factory StorageFactory) {
	StorageRegistry[name] = factory
}

// GetStorage returns a storage backend by name
func GetStorage(name string, config map[string]interface{}) (Storage, error) {
	factory, exists := StorageRegistry[name]
	if !exists {
		return nil, ErrStorageNotFound
	}
	return factory(config)
}
