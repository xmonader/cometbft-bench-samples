package storage

import (
	"sync"

	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// MemoryStorage implements the Storage interface using in-memory data structures
type MemoryStorage struct {
	accounts    map[int]*types.Account
	mutex       sync.RWMutex
	initialized bool
	inTx        bool
	txAccounts  map[int]*types.Account // Accounts in the current transaction
}

// NewMemoryStorage creates a new memory storage instance
func NewMemoryStorage(config map[string]interface{}) (Storage, error) {
	return &MemoryStorage{
		accounts:   make(map[int]*types.Account),
		txAccounts: make(map[int]*types.Account),
	}, nil
}

// Initialize initializes the storage
func (s *MemoryStorage) Initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.initialized {
		return ErrAlreadyInitialized
	}

	s.accounts = make(map[int]*types.Account)
	s.txAccounts = make(map[int]*types.Account)
	s.initialized = true
	return nil
}

// Close closes the storage
func (s *MemoryStorage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	s.accounts = nil
	s.txAccounts = nil
	s.initialized = false
	return nil
}

// GetAccount retrieves an account by ID
func (s *MemoryStorage) GetAccount(id int) (*types.Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, ErrNotInitialized
	}

	// If in a transaction, check the transaction accounts first
	if s.inTx {
		if acc, exists := s.txAccounts[id]; exists {
			return acc, nil
		}
	}

	// Check the main accounts
	if acc, exists := s.accounts[id]; exists {
		// If in a transaction, make a copy for the transaction
		if s.inTx {
			accCopy := &types.Account{
				ID:      acc.ID,
				Balance: acc.Balance,
			}
			s.txAccounts[id] = accCopy
			return accCopy, nil
		}
		return acc, nil
	}

	return nil, ErrAccountNotFound
}

// AccountExists checks if an account exists
func (s *MemoryStorage) AccountExists(id int) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return false, ErrNotInitialized
	}

	// If in a transaction, check the transaction accounts first
	if s.inTx {
		if _, exists := s.txAccounts[id]; exists {
			return true, nil
		}
	}

	_, exists := s.accounts[id]
	return exists, nil
}

// UpdateBalance updates an account's balance
func (s *MemoryStorage) UpdateBalance(id int, delta int) error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// Get the account (this will handle transaction state)
	acc, err := s.GetAccount(id)
	if err != nil {
		// If account doesn't exist, create it with the delta as initial balance
		// (only if delta is positive)
		if err == ErrAccountNotFound && delta > 0 {
			return s.CreateAccount(id, delta)
		}
		return err
	}

	// Check if the balance would go negative
	newBalance := acc.Balance + delta
	if newBalance < 0 {
		return ErrInsufficientBalance
	}

	// Update the balance
	acc.Balance = newBalance
	return nil
}

// CreateAccount creates a new account
func (s *MemoryStorage) CreateAccount(id int, initialBalance int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// Check if account already exists
	if s.inTx {
		if _, exists := s.txAccounts[id]; exists {
			return ErrAccountAlreadyExists
		}
	}

	if _, exists := s.accounts[id]; exists && !s.inTx {
		return ErrAccountAlreadyExists
	}

	// Create the account
	acc := &types.Account{
		ID:      id,
		Balance: initialBalance,
	}

	// Store the account
	if s.inTx {
		s.txAccounts[id] = acc
	} else {
		s.accounts[id] = acc
	}

	return nil
}

// Commit commits any pending changes
func (s *MemoryStorage) Commit() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// If not in a transaction, nothing to commit
	if !s.inTx {
		return nil
	}

	// Apply transaction changes to the main accounts
	for id, acc := range s.txAccounts {
		s.accounts[id] = acc
	}

	// Clear transaction state
	s.txAccounts = make(map[int]*types.Account)
	s.inTx = false

	return nil
}

// Rollback rolls back any pending changes
func (s *MemoryStorage) Rollback() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// If not in a transaction, nothing to rollback
	if !s.inTx {
		return nil
	}

	// Clear transaction state
	s.txAccounts = make(map[int]*types.Account)
	s.inTx = false

	return nil
}

// BeginTransaction begins a new transaction
func (s *MemoryStorage) BeginTransaction() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// If already in a transaction, commit it first
	if s.inTx {
		// Apply transaction changes to the main accounts
		for id, acc := range s.txAccounts {
			s.accounts[id] = acc
		}
	}

	// Start a new transaction
	s.txAccounts = make(map[int]*types.Account)
	s.inTx = true

	return nil
}

// GetBalance gets an account's current balance
func (s *MemoryStorage) GetBalance(id int) (int, error) {
	acc, err := s.GetAccount(id)
	if err != nil {
		return 0, err
	}
	return acc.Balance, nil
}

// GetAllAccounts gets all accounts
func (s *MemoryStorage) GetAllAccounts() ([]*types.Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, ErrNotInitialized
	}

	accounts := make([]*types.Account, 0, len(s.accounts))

	// If in a transaction, use the transaction state
	if s.inTx {
		// Start with all accounts from the main state
		accountMap := make(map[int]*types.Account)
		for id, acc := range s.accounts {
			// Make a copy to avoid modifying the original
			accountMap[id] = &types.Account{
				ID:      acc.ID,
				Balance: acc.Balance,
			}
		}

		// Apply transaction changes
		for id, acc := range s.txAccounts {
			accountMap[id] = acc
		}

		// Convert map to slice
		for _, acc := range accountMap {
			accounts = append(accounts, acc)
		}
	} else {
		// Just use the main state
		for _, acc := range s.accounts {
			accounts = append(accounts, acc)
		}
	}

	return accounts, nil
}

func init() {
	// Register the memory storage backend
	RegisterStorage("memory", NewMemoryStorage)
}
