package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// BadgerStorage implements the Storage interface using BadgerDB
type BadgerStorage struct {
	db          *badger.DB
	initialized bool
	dbPath      string
	txn         *badger.Txn
}

// NewBadgerStorage creates a new BadgerDB storage instance
func NewBadgerStorage(config map[string]interface{}) (Storage, error) {
	// Get the database path from config
	dbPathInterface, ok := config["db_path"]
	if !ok {
		return nil, fmt.Errorf("%w: db_path is required", ErrInvalidConfiguration)
	}

	dbPath, ok := dbPathInterface.(string)
	if !ok {
		return nil, fmt.Errorf("%w: db_path must be a string", ErrInvalidConfiguration)
	}

	// Create the directory if it doesn't exist
	if err := createDirIfNotExists(filepath.Dir(dbPath)); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &BadgerStorage{
		dbPath: dbPath,
	}, nil
}

// Initialize initializes the storage
func (s *BadgerStorage) Initialize() error {
	if s.initialized {
		return ErrAlreadyInitialized
	}

	// Open the BadgerDB database
	opts := badger.DefaultOptions(s.dbPath)
	opts.Logger = nil // Disable logging
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	s.db = db
	s.initialized = true
	return nil
}

// Close closes the storage
func (s *BadgerStorage) Close() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	if s.txn != nil {
		s.txn.Discard()
		s.txn = nil
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	s.initialized = false
	return nil
}

// accountKey generates a key for an account
func accountKey(id int) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, uint64(id))
	return key
}

// GetAccount retrieves an account by ID
func (s *BadgerStorage) GetAccount(id int) (*types.Account, error) {
	if !s.initialized {
		return nil, ErrNotInitialized
	}

	key := accountKey(id)
	var account *types.Account

	// Use the transaction if one is active
	if s.txn != nil {
		item, err := s.txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil, ErrAccountNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get account: %w", err)
		}

		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &account)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal account: %w", err)
		}

		return account, nil
	}

	// No active transaction, use a read-only transaction
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		return item.Value(func(val []byte) error {
			account = &types.Account{}
			return json.Unmarshal(val, account)
		})
	})

	if err == ErrAccountNotFound {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

// AccountExists checks if an account exists
func (s *BadgerStorage) AccountExists(id int) (bool, error) {
	if !s.initialized {
		return false, ErrNotInitialized
	}

	key := accountKey(id)

	// Use the transaction if one is active
	if s.txn != nil {
		_, err := s.txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("failed to check account existence: %w", err)
		}
		return true, nil
	}

	// No active transaction, use a read-only transaction
	var exists bool
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			exists = false
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to check account existence: %w", err)
		}
		exists = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

// UpdateBalance updates an account's balance
func (s *BadgerStorage) UpdateBalance(id int, delta int) error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// Get the account
	account, err := s.GetAccount(id)
	if err != nil {
		// If account doesn't exist, create it with the delta as initial balance
		// (only if delta is positive)
		if err == ErrAccountNotFound && delta > 0 {
			return s.CreateAccount(id, delta)
		}
		return err
	}

	// Check if the balance would go negative
	newBalance := account.Balance + delta
	if newBalance < 0 {
		return ErrInsufficientBalance
	}

	// Update the balance
	account.Balance = newBalance

	// Serialize the account
	accountData, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	// Use the transaction if one is active
	if s.txn != nil {
		if err := s.txn.Set(accountKey(id), accountData); err != nil {
			return fmt.Errorf("failed to update account: %w", err)
		}
		return nil
	}

	// No active transaction, use a new transaction
	return s.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(accountKey(id), accountData); err != nil {
			return fmt.Errorf("failed to update account: %w", err)
		}
		return nil
	})
}

// CreateAccount creates a new account
func (s *BadgerStorage) CreateAccount(id int, initialBalance int) error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// Check if account already exists
	exists, err := s.AccountExists(id)
	if err != nil {
		return err
	}
	if exists {
		return ErrAccountAlreadyExists
	}

	// Create the account
	account := &types.Account{
		ID:      id,
		Balance: initialBalance,
	}

	// Serialize the account
	accountData, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	// Use the transaction if one is active
	if s.txn != nil {
		if err := s.txn.Set(accountKey(id), accountData); err != nil {
			return fmt.Errorf("failed to create account: %w", err)
		}
		return nil
	}

	// No active transaction, use a new transaction
	return s.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(accountKey(id), accountData); err != nil {
			return fmt.Errorf("failed to create account: %w", err)
		}
		return nil
	})
}

// Commit commits any pending changes
func (s *BadgerStorage) Commit() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// If no transaction is active, nothing to commit
	if s.txn == nil {
		return nil
	}

	// Commit the transaction
	if err := s.txn.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrTransactionFailed, err)
	}

	// Clear the transaction
	s.txn = nil

	return nil
}

// Rollback rolls back any pending changes
func (s *BadgerStorage) Rollback() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// If no transaction is active, nothing to rollback
	if s.txn == nil {
		return nil
	}

	// Discard the transaction
	s.txn.Discard()
	s.txn = nil

	return nil
}

// BeginTransaction begins a new transaction
func (s *BadgerStorage) BeginTransaction() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// If a transaction is already active, commit it first
	if s.txn != nil {
		if err := s.Commit(); err != nil {
			return err
		}
	}

	// Start a new transaction
	s.txn = s.db.NewTransaction(true)

	return nil
}

// GetBalance gets an account's current balance
func (s *BadgerStorage) GetBalance(id int) (int, error) {
	account, err := s.GetAccount(id)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

// GetAllAccounts gets all accounts
func (s *BadgerStorage) GetAllAccounts() ([]*types.Account, error) {
	if !s.initialized {
		return nil, ErrNotInitialized
	}

	var accounts []*types.Account

	// Use the transaction if one is active
	if s.txn != nil {
		it := s.txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var account types.Account
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &account)
			})
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal account: %w", err)
			}
			accounts = append(accounts, &account)
		}

		return accounts, nil
	}

	// No active transaction, use a read-only transaction
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var account types.Account
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &account)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal account: %w", err)
			}
			accounts = append(accounts, &account)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func init() {
	// Register the BadgerDB storage backend
	RegisterStorage("badger", NewBadgerStorage)
}
