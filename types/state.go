package types

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Account represents a user account with a balance
type Account struct {
	ID      int `json:"id"`
	Balance int `json:"balance"`
}

// State represents the application state
type State struct {
	Accounts map[int]*Account `json:"accounts"`
	mutex    sync.RWMutex     `json:"-"` // Mutex for thread safety, not serialized
}

// NewState creates a new application state
func NewState() *State {
	return &State{
		Accounts: make(map[int]*Account),
	}
}

// GetAccount returns an account by ID, creating it if it doesn't exist
func (s *State) GetAccount(id int) *Account {
	s.mutex.RLock()
	acc, exists := s.Accounts[id]
	s.mutex.RUnlock()

	if !exists {
		// Create a new account with zero balance
		acc = &Account{
			ID:      id,
			Balance: 0,
		}
		s.mutex.Lock()
		s.Accounts[id] = acc
		s.mutex.Unlock()
	}

	return acc
}

// UpdateBalance updates an account's balance
func (s *State) UpdateBalance(id int, delta int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	acc, exists := s.Accounts[id]
	if !exists {
		acc = &Account{
			ID:      id,
			Balance: 0,
		}
		s.Accounts[id] = acc
	}

	newBalance := acc.Balance + delta
	if newBalance < 0 {
		return fmt.Errorf("insufficient balance for account %d: %d < %d", id, acc.Balance, -delta)
	}

	acc.Balance = newBalance
	return nil
}

// Serialize serializes the state to JSON
func (s *State) Serialize() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return json.Marshal(s)
}

// LoadState loads state from JSON
func LoadState(data []byte) (*State, error) {
	var state State
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	// Initialize the mutex
	state.mutex = sync.RWMutex{}

	// If accounts is nil, initialize it
	if state.Accounts == nil {
		state.Accounts = make(map[int]*Account)
	}

	return &state, nil
}

// String returns a string representation of the state
func (s *State) String() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling state: %v", err)
	}
	return string(data)
}
