package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// RedisStorage implements the Storage interface using Redis
type RedisStorage struct {
	client      *redis.Client
	initialized bool
	ctx         context.Context
	mutex       sync.RWMutex
	accounts    map[int]*types.Account // Cache for accounts
	txAccounts  map[int]*types.Account // Accounts in the current transaction
	inTx        bool
	keyPrefix   string
	addr        string
	password    string
	db          int
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(config map[string]interface{}) (Storage, error) {
	// Get the Redis address from config
	addrInterface, ok := config["address"]
	if !ok {
		return nil, fmt.Errorf("%w: address is required", ErrInvalidConfiguration)
	}

	addr, ok := addrInterface.(string)
	if !ok {
		return nil, fmt.Errorf("%w: address must be a string", ErrInvalidConfiguration)
	}

	// Get the Redis password from config (optional)
	var password string
	passwordInterface, ok := config["password"]
	if ok {
		password, ok = passwordInterface.(string)
		if !ok {
			return nil, fmt.Errorf("%w: password must be a string", ErrInvalidConfiguration)
		}
	}

	// Get the Redis database from config (optional)
	var db int
	dbInterface, ok := config["db"]
	if ok {
		dbFloat, ok := dbInterface.(float64)
		if !ok {
			return nil, fmt.Errorf("%w: db must be a number", ErrInvalidConfiguration)
		}
		db = int(dbFloat)
	}

	// Get the key prefix from config (optional)
	var keyPrefix string
	keyPrefixInterface, ok := config["key_prefix"]
	if ok {
		keyPrefix, ok = keyPrefixInterface.(string)
		if !ok {
			return nil, fmt.Errorf("%w: key_prefix must be a string", ErrInvalidConfiguration)
		}
	} else {
		keyPrefix = "account:"
	}

	return &RedisStorage{
		ctx:        context.Background(),
		accounts:   make(map[int]*types.Account),
		txAccounts: make(map[int]*types.Account),
		keyPrefix:  keyPrefix,
		addr:       addr,
		password:   password,
		db:         db,
	}, nil
}

// Initialize initializes the storage
func (s *RedisStorage) Initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.initialized {
		return ErrAlreadyInitialized
	}

	// Create a Redis client
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.addr,
		Password: s.password,
		DB:       s.db,
	})

	// Ping the Redis server to check if it's available
	if err := s.client.Ping(s.ctx).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	s.initialized = true
	return nil
}

// Close closes the storage
func (s *RedisStorage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// Clear transaction state
	s.txAccounts = make(map[int]*types.Account)
	s.inTx = false

	// Close the Redis client
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	s.client = nil
	s.initialized = false
	return nil
}

// accountKey generates a key for an account
func (s *RedisStorage) accountKey(id int) string {
	return s.keyPrefix + strconv.Itoa(id)
}

// GetAccount retrieves an account by ID
func (s *RedisStorage) GetAccount(id int) (*types.Account, error) {
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

	// Check the cache
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

	// Get the account from Redis
	key := s.accountKey(id)
	val, err := s.client.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Deserialize the account
	var account types.Account
	if err := json.Unmarshal([]byte(val), &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	// Cache the account
	s.accounts[id] = &account

	// If in a transaction, make a copy for the transaction
	if s.inTx {
		accCopy := &types.Account{
			ID:      account.ID,
			Balance: account.Balance,
		}
		s.txAccounts[id] = accCopy
		return accCopy, nil
	}

	return &account, nil
}

// AccountExists checks if an account exists
func (s *RedisStorage) AccountExists(id int) (bool, error) {
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

	// Check the cache
	if _, exists := s.accounts[id]; exists {
		return true, nil
	}

	// Check if the account exists in Redis
	key := s.accountKey(id)
	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}

	return exists > 0, nil
}

// UpdateBalance updates an account's balance
func (s *RedisStorage) UpdateBalance(id int, delta int) error {
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

	// If in a transaction, just update the transaction account
	if s.inTx {
		s.txAccounts[id] = account
		return nil
	}

	// Serialize the account
	accountData, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	// Update the account in Redis
	key := s.accountKey(id)
	if err := s.client.Set(s.ctx, key, accountData, 0).Err(); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	// Update the cache
	s.accounts[id] = account

	return nil
}

// CreateAccount creates a new account
func (s *RedisStorage) CreateAccount(id int, initialBalance int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

	// If in a transaction, just store the account in the transaction
	if s.inTx {
		s.txAccounts[id] = account
		return nil
	}

	// Serialize the account
	accountData, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	// Store the account in Redis
	key := s.accountKey(id)
	if err := s.client.Set(s.ctx, key, accountData, 0).Err(); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	// Cache the account
	s.accounts[id] = account

	return nil
}

// Commit commits any pending changes
func (s *RedisStorage) Commit() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// If not in a transaction, nothing to commit
	if !s.inTx {
		return nil
	}

	// Use a Redis pipeline to batch the updates
	pipe := s.client.Pipeline()

	// Apply all transaction changes
	for id, account := range s.txAccounts {
		// Serialize the account
		accountData, err := json.Marshal(account)
		if err != nil {
			return fmt.Errorf("failed to marshal account: %w", err)
		}

		// Store the account in Redis
		key := s.accountKey(id)
		pipe.Set(s.ctx, key, accountData, 0)

		// Update the cache
		s.accounts[id] = account
	}

	// Execute the pipeline
	if _, err := pipe.Exec(s.ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clear transaction state
	s.txAccounts = make(map[int]*types.Account)
	s.inTx = false

	return nil
}

// Rollback rolls back any pending changes
func (s *RedisStorage) Rollback() error {
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
func (s *RedisStorage) BeginTransaction() error {
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
func (s *RedisStorage) GetBalance(id int) (int, error) {
	account, err := s.GetAccount(id)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

// GetAllAccounts gets all accounts
func (s *RedisStorage) GetAllAccounts() ([]*types.Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, ErrNotInitialized
	}

	// Get all account keys from Redis
	keys, err := s.client.Keys(s.ctx, s.keyPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get account keys: %w", err)
	}

	// If no accounts found, return an empty slice
	if len(keys) == 0 {
		return []*types.Account{}, nil
	}

	// Get all accounts in a single operation
	vals, err := s.client.MGet(s.ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	// Parse the accounts
	accounts := make([]*types.Account, 0, len(vals))
	for _, val := range vals {
		if val == nil {
			continue
		}

		var account types.Account
		if err := json.Unmarshal([]byte(val.(string)), &account); err != nil {
			return nil, fmt.Errorf("failed to unmarshal account: %w", err)
		}

		// If in a transaction, check if the account has been modified
		if s.inTx {
			if acc, exists := s.txAccounts[account.ID]; exists {
				accounts = append(accounts, acc)
				continue
			}
		}

		accounts = append(accounts, &account)
	}

	// If in a transaction, add any new accounts
	if s.inTx {
		for id, acc := range s.txAccounts {
			// Check if the account is already in the list
			found := false
			for _, a := range accounts {
				if a.ID == id {
					found = true
					break
				}
			}
			if !found {
				accounts = append(accounts, acc)
			}
		}
	}

	return accounts, nil
}

func init() {
	// Register the Redis storage backend
	RegisterStorage("redis", NewRedisStorage)
}
