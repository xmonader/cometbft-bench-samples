package storage

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"

	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	tbtypes "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// TigerBeetleStorage implements the Storage interface using TigerBeetle
type TigerBeetleStorage struct {
	client      tigerbeetle.Client
	initialized bool
	addresses   []string
	clusterID   tbtypes.Uint128
	mutex       sync.RWMutex
	accounts    map[int]*types.Account // Cache for accounts
	txAccounts  map[int]*types.Account // Accounts in the current transaction
	inTx        bool
}

// NewTigerBeetleStorage creates a new TigerBeetle storage instance
func NewTigerBeetleStorage(config map[string]interface{}) (Storage, error) {
	// Get the addresses from config
	addressesInterface, ok := config["addresses"]
	if !ok {
		return nil, fmt.Errorf("%w: addresses is required", ErrInvalidConfiguration)
	}

	addressesSlice, ok := addressesInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: addresses must be an array of strings", ErrInvalidConfiguration)
	}

	addresses := make([]string, len(addressesSlice))
	for i, addr := range addressesSlice {
		addrStr, ok := addr.(string)
		if !ok {
			return nil, fmt.Errorf("%w: address must be a string", ErrInvalidConfiguration)
		}
		addresses[i] = addrStr
	}

	// Get the cluster ID from config
	clusterIDInterface, ok := config["cluster_id"]
	if !ok {
		return nil, fmt.Errorf("%w: cluster_id is required", ErrInvalidConfiguration)
	}

	var clusterID tbtypes.Uint128
	switch v := clusterIDInterface.(type) {
	case float64:
		clusterID = tbtypes.ToUint128(uint64(v))
	case string:
		var err error
		clusterID, err = tbtypes.HexStringToUint128(v)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid cluster_id format: %v", ErrInvalidConfiguration, err)
		}
	default:
		return nil, fmt.Errorf("%w: cluster_id must be a number or hex string", ErrInvalidConfiguration)
	}

	// Create the data directory if it doesn't exist
	dataDir, ok := config["data_dir"].(string)
	if ok {
		if err := createDirIfNotExists(filepath.Dir(dataDir)); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	return &TigerBeetleStorage{
		addresses:  addresses,
		clusterID:  clusterID,
		accounts:   make(map[int]*types.Account),
		txAccounts: make(map[int]*types.Account),
	}, nil
}

// Initialize initializes the storage
func (s *TigerBeetleStorage) Initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.initialized {
		return ErrAlreadyInitialized
	}

	// Create a TigerBeetle client
	client, err := tigerbeetle.NewClient(s.clusterID, s.addresses)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	s.client = client
	s.initialized = true
	return nil
}

// Close closes the storage
func (s *TigerBeetleStorage) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// Clear transaction state
	s.txAccounts = make(map[int]*types.Account)
	s.inTx = false

	// Close the client
	s.client.Close()
	s.client = nil
	s.initialized = false
	return nil
}

// accountID converts an account ID to a TigerBeetle account ID
func accountID(id int) tbtypes.Uint128 {
	var bytes [16]byte
	binary.BigEndian.PutUint64(bytes[8:], uint64(id))
	return tbtypes.BytesToUint128(bytes)
}

// GetAccount retrieves an account by ID
func (s *TigerBeetleStorage) GetAccount(id int) (*types.Account, error) {
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

	// Lookup the account in TigerBeetle
	tbID := accountID(id)
	accounts, err := s.client.LookupAccounts([]tbtypes.Uint128{tbID})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup account: %w", err)
	}

	// Check if the account exists
	if len(accounts) == 0 {
		return nil, ErrAccountNotFound
	}

	// Convert the TigerBeetle account to our account type
	tbAccount := accounts[0]

	// In TigerBeetle, the balance is CreditsPosted - DebitsPosted
	creditsPosted := tbAccount.CreditsPosted.BigInt()
	debitsPosted := tbAccount.DebitsPosted.BigInt()
	balanceBigInt := new(big.Int).Sub(&creditsPosted, &debitsPosted)

	// Convert to int (assuming balance fits in int)
	balance := int(balanceBigInt.Int64())

	account := &types.Account{
		ID:      id,
		Balance: balance,
	}

	// Cache the account
	s.accounts[id] = account

	// If in a transaction, make a copy for the transaction
	if s.inTx {
		accCopy := &types.Account{
			ID:      account.ID,
			Balance: account.Balance,
		}
		s.txAccounts[id] = accCopy
		return accCopy, nil
	}

	return account, nil
}

// AccountExists checks if an account exists
func (s *TigerBeetleStorage) AccountExists(id int) (bool, error) {
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

	// Lookup the account in TigerBeetle
	tbID := accountID(id)
	accounts, err := s.client.LookupAccounts([]tbtypes.Uint128{tbID})
	if err != nil {
		return false, fmt.Errorf("failed to lookup account: %w", err)
	}

	return len(accounts) > 0, nil
}

// UpdateBalance updates an account's balance
func (s *TigerBeetleStorage) UpdateBalance(id int, delta int) error {
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

	// Create a transfer to update the balance
	transfer := tbtypes.Transfer{
		ID:              tbtypes.ToUint128(uint64(id)),
		DebitAccountID:  accountID(0), // System account
		CreditAccountID: accountID(id),
		Amount:          tbtypes.ToUint128(uint64(delta)),
		PendingID:       tbtypes.ToUint128(0),
		UserData128:     tbtypes.ToUint128(0),
		UserData64:      0,
		UserData32:      0,
		Timeout:         0,
		Ledger:          0,
		Code:            0,
		Flags:           0,
		Timestamp:       0,
	}

	// Execute the transfer
	result, err := s.client.CreateTransfers([]tbtypes.Transfer{transfer})
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	// Check for errors
	if len(result) > 0 {
		return fmt.Errorf("failed to update account balance: %v", result[0])
	}

	// Update the cache
	s.accounts[id] = account

	return nil
}

// CreateAccount creates a new account
func (s *TigerBeetleStorage) CreateAccount(id int, initialBalance int) error {
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

	// Create a TigerBeetle account
	tbAccount := tbtypes.Account{
		ID:             accountID(id),
		DebitsPending:  tbtypes.ToUint128(0),
		DebitsPosted:   tbtypes.ToUint128(0),
		CreditsPending: tbtypes.ToUint128(0),
		CreditsPosted:  tbtypes.ToUint128(uint64(initialBalance)),
		UserData128:    tbtypes.ToUint128(0),
		UserData64:     0,
		UserData32:     0,
		Reserved:       0,
		Ledger:         0,
		Code:           0,
		Flags:          0,
		Timestamp:      0,
	}

	// Create the account in TigerBeetle
	result, err := s.client.CreateAccounts([]tbtypes.Account{tbAccount})
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	// Check for errors
	if len(result) > 0 {
		return fmt.Errorf("failed to create account: %v", result[0])
	}

	// Cache the account
	s.accounts[id] = account

	return nil
}

// Commit commits any pending changes
func (s *TigerBeetleStorage) Commit() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return ErrNotInitialized
	}

	// If not in a transaction, nothing to commit
	if !s.inTx {
		return nil
	}

	// Create accounts that don't exist
	for id, account := range s.txAccounts {
		if _, exists := s.accounts[id]; !exists {
			// Create a TigerBeetle account
			tbAccount := tbtypes.Account{
				ID:             accountID(id),
				DebitsPending:  tbtypes.ToUint128(0),
				DebitsPosted:   tbtypes.ToUint128(0),
				CreditsPending: tbtypes.ToUint128(0),
				CreditsPosted:  tbtypes.ToUint128(uint64(account.Balance)),
				UserData128:    tbtypes.ToUint128(0),
				UserData64:     0,
				UserData32:     0,
				Reserved:       0,
				Ledger:         0,
				Code:           0,
				Flags:          0,
				Timestamp:      0,
			}

			// Create the account in TigerBeetle
			result, err := s.client.CreateAccounts([]tbtypes.Account{tbAccount})
			if err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}

			// Check for errors
			if len(result) > 0 {
				return fmt.Errorf("failed to create account: %v", result[0])
			}
		} else {
			// Update the account balance
			delta := account.Balance - s.accounts[id].Balance
			if delta != 0 {
				// Create a transfer to update the balance
				transfer := tbtypes.Transfer{
					ID:              tbtypes.ToUint128(uint64(id)),
					DebitAccountID:  accountID(0), // System account
					CreditAccountID: accountID(id),
					Amount:          tbtypes.ToUint128(uint64(delta)),
					PendingID:       tbtypes.ToUint128(0),
					UserData128:     tbtypes.ToUint128(0),
					UserData64:      0,
					UserData32:      0,
					Timeout:         0,
					Ledger:          0,
					Code:            0,
					Flags:           0,
					Timestamp:       0,
				}

				// Execute the transfer
				result, err := s.client.CreateTransfers([]tbtypes.Transfer{transfer})
				if err != nil {
					return fmt.Errorf("failed to update account balance: %w", err)
				}

				// Check for errors
				if len(result) > 0 {
					return fmt.Errorf("failed to update account balance: %v", result[0])
				}
			}
		}

		// Update the cache
		s.accounts[id] = account
	}

	// Clear transaction state
	s.txAccounts = make(map[int]*types.Account)
	s.inTx = false

	return nil
}

// Rollback rolls back any pending changes
func (s *TigerBeetleStorage) Rollback() error {
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
func (s *TigerBeetleStorage) BeginTransaction() error {
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
func (s *TigerBeetleStorage) GetBalance(id int) (int, error) {
	account, err := s.GetAccount(id)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

// GetAllAccounts gets all accounts
func (s *TigerBeetleStorage) GetAllAccounts() ([]*types.Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, ErrNotInitialized
	}

	// TigerBeetle doesn't have a way to get all accounts, so we just return the cached accounts
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
	// Register the TigerBeetle storage backend
	RegisterStorage("tigerbeetle", NewTigerBeetleStorage)
}
