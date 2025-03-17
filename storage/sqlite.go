package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// SQLiteStorage implements the Storage interface using SQLite
type SQLiteStorage struct {
	db          *sql.DB
	initialized bool
	dbPath      string
	tx          *sql.Tx
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(config map[string]interface{}) (Storage, error) {
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

	return &SQLiteStorage{
		dbPath: dbPath,
	}, nil
}

// Initialize initializes the storage
func (s *SQLiteStorage) Initialize() error {
	if s.initialized {
		return ErrAlreadyInitialized
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Create the accounts table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY,
			balance INTEGER NOT NULL
		)
	`)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create accounts table: %w", err)
	}

	s.db = db
	s.initialized = true
	return nil
}

// Close closes the storage
func (s *SQLiteStorage) Close() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	if s.tx != nil {
		s.tx.Rollback()
		s.tx = nil
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	s.initialized = false
	return nil
}

// GetAccount retrieves an account by ID
func (s *SQLiteStorage) GetAccount(id int) (*types.Account, error) {
	if !s.initialized {
		return nil, ErrNotInitialized
	}

	var query string
	var queryArgs []interface{}
	var rows *sql.Rows
	var err error

	query = "SELECT id, balance FROM accounts WHERE id = ?"
	queryArgs = []interface{}{id}

	// Use the transaction if one is active
	if s.tx != nil {
		rows, err = s.tx.Query(query, queryArgs...)
	} else {
		rows, err = s.db.Query(query, queryArgs...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query account: %w", err)
	}
	defer rows.Close()

	// Check if the account exists
	if !rows.Next() {
		return nil, ErrAccountNotFound
	}

	// Parse the account
	var account types.Account
	if err := rows.Scan(&account.ID, &account.Balance); err != nil {
		return nil, fmt.Errorf("failed to scan account: %w", err)
	}

	return &account, nil
}

// AccountExists checks if an account exists
func (s *SQLiteStorage) AccountExists(id int) (bool, error) {
	if !s.initialized {
		return false, ErrNotInitialized
	}

	var query string
	var queryArgs []interface{}
	var row *sql.Row

	query = "SELECT 1 FROM accounts WHERE id = ? LIMIT 1"
	queryArgs = []interface{}{id}

	// Use the transaction if one is active
	if s.tx != nil {
		row = s.tx.QueryRow(query, queryArgs...)
	} else {
		row = s.db.QueryRow(query, queryArgs...)
	}

	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}

	return true, nil
}

// UpdateBalance updates an account's balance
func (s *SQLiteStorage) UpdateBalance(id int, delta int) error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// Check if the account exists
	exists, err := s.AccountExists(id)
	if err != nil {
		return err
	}

	// If the account doesn't exist and delta is positive, create it
	if !exists {
		if delta <= 0 {
			return ErrAccountNotFound
		}
		return s.CreateAccount(id, delta)
	}

	// Get the current balance
	account, err := s.GetAccount(id)
	if err != nil {
		return err
	}

	// Check if the balance would go negative
	newBalance := account.Balance + delta
	if newBalance < 0 {
		return ErrInsufficientBalance
	}

	// Update the balance
	var query string
	var queryArgs []interface{}
	var result sql.Result

	query = "UPDATE accounts SET balance = ? WHERE id = ?"
	queryArgs = []interface{}{newBalance, id}

	// Use the transaction if one is active
	if s.tx != nil {
		result, err = s.tx.Exec(query, queryArgs...)
	} else {
		result, err = s.db.Exec(query, queryArgs...)
	}

	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	// Check if the update was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrAccountNotFound
	}

	return nil
}

// CreateAccount creates a new account
func (s *SQLiteStorage) CreateAccount(id int, initialBalance int) error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// Check if the account already exists
	exists, err := s.AccountExists(id)
	if err != nil {
		return err
	}
	if exists {
		return ErrAccountAlreadyExists
	}

	// Create the account
	var query string
	var queryArgs []interface{}
	var result sql.Result

	query = "INSERT INTO accounts (id, balance) VALUES (?, ?)"
	queryArgs = []interface{}{id, initialBalance}

	// Use the transaction if one is active
	if s.tx != nil {
		result, err = s.tx.Exec(query, queryArgs...)
	} else {
		result, err = s.db.Exec(query, queryArgs...)
	}

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	// Check if the insert was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("failed to create account: no rows affected")
	}

	return nil
}

// Commit commits any pending changes
func (s *SQLiteStorage) Commit() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// If no transaction is active, nothing to commit
	if s.tx == nil {
		return nil
	}

	// Commit the transaction
	if err := s.tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrTransactionFailed, err)
	}

	// Clear the transaction
	s.tx = nil

	return nil
}

// Rollback rolls back any pending changes
func (s *SQLiteStorage) Rollback() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// If no transaction is active, nothing to rollback
	if s.tx == nil {
		return nil
	}

	// Rollback the transaction
	if err := s.tx.Rollback(); err != nil {
		return fmt.Errorf("%w: %v", ErrTransactionFailed, err)
	}

	// Clear the transaction
	s.tx = nil

	return nil
}

// BeginTransaction begins a new transaction
func (s *SQLiteStorage) BeginTransaction() error {
	if !s.initialized {
		return ErrNotInitialized
	}

	// If a transaction is already active, commit it first
	if s.tx != nil {
		if err := s.Commit(); err != nil {
			return err
		}
	}

	// Start a new transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	s.tx = tx
	return nil
}

// GetBalance gets an account's current balance
func (s *SQLiteStorage) GetBalance(id int) (int, error) {
	account, err := s.GetAccount(id)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

// GetAllAccounts gets all accounts
func (s *SQLiteStorage) GetAllAccounts() ([]*types.Account, error) {
	if !s.initialized {
		return nil, ErrNotInitialized
	}

	var query string
	var rows *sql.Rows
	var err error

	query = "SELECT id, balance FROM accounts"

	// Use the transaction if one is active
	if s.tx != nil {
		rows, err = s.tx.Query(query)
	} else {
		rows, err = s.db.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	// Parse the accounts
	var accounts []*types.Account
	for rows.Next() {
		var account types.Account
		if err := rows.Scan(&account.ID, &account.Balance); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}

func init() {
	// Register the SQLite storage backend
	RegisterStorage("sqlite", NewSQLiteStorage)
}
