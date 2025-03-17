package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// StateStore manages the application state
type StateStore struct {
	state      *types.State
	stateMutex sync.RWMutex
	stateFile  string
}

// NewStateStore creates a new state store
func NewStateStore(stateFile string) *StateStore {
	return &StateStore{
		state:     types.NewState(),
		stateFile: stateFile,
	}
}

// GetState returns the current state
func (s *StateStore) GetState() *types.State {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.state
}

// UpdateState updates the state with a function
func (s *StateStore) UpdateState(updateFn func(*types.State) error) error {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	if err := updateFn(s.state); err != nil {
		return err
	}

	return nil
}

// SaveState saves the state to disk
func (s *StateStore) SaveState() error {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()

	if s.stateFile == "" {
		return nil // No state file configured, skip saving
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Serialize state
	data, err := s.state.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	// Write to temporary file first
	tempFile := s.stateFile + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state to temp file: %w", err)
	}

	// Rename to actual file (atomic operation)
	if err := os.Rename(tempFile, s.stateFile); err != nil {
		return fmt.Errorf("failed to rename temp state file: %w", err)
	}

	return nil
}

// LoadState loads the state from disk
func (s *StateStore) LoadState() error {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	if s.stateFile == "" {
		return nil // No state file configured, skip loading
	}

	// Check if file exists
	if _, err := os.Stat(s.stateFile); os.IsNotExist(err) {
		// File doesn't exist, use default state
		s.state = types.NewState()
		return nil
	}

	// Read file
	data, err := ioutil.ReadFile(s.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse state
	state, err := types.LoadState(data)
	if err != nil {
		return fmt.Errorf("failed to parse state: %w", err)
	}

	s.state = state
	return nil
}

// GetAccount gets an account by ID
func (s *StateStore) GetAccount(id int) *types.Account {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.state.GetAccount(id)
}

// UpdateBalance updates an account's balance
func (s *StateStore) UpdateBalance(id int, delta int) error {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	return s.state.UpdateBalance(id, delta)
}

// String returns a string representation of the state
func (s *StateStore) String() string {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.state.String()
}
