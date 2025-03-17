package app

import (
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/xmonader/test_batched_tx_tendermint/crypto"
	"github.com/xmonader/test_batched_tx_tendermint/types"
)

// Application implements the ABCI application interface
type Application struct {
	abci.BaseApplication

	stateStore          *StateStore
	txProcessor         *TransactionProcessor
	logger              log.Logger
	pendingTransactions map[string]*types.Transaction
}

// NewApplication creates a new ABCI application
func NewApplication(stateFile string, logger log.Logger) *Application {
	stateStore := NewStateStore(stateFile)
	txProcessor := NewTransactionProcessor(stateStore)

	return &Application{
		stateStore:          stateStore,
		txProcessor:         txProcessor,
		logger:              logger,
		pendingTransactions: make(map[string]*types.Transaction),
	}
}

// Info returns information about the application state
func (app *Application) Info(req abci.InfoRequest) abci.InfoResponse {
	return abci.InfoResponse{
		Data:             "Batched Transaction ABCI App",
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  0,
		LastBlockAppHash: []byte{},
	}
}

// InitChain initializes the blockchain with validators and initial app state
func (app *Application) InitChain(req abci.InitChainRequest) abci.InitChainResponse {
	// Load state from disk if available
	if err := app.stateStore.LoadState(); err != nil {
		app.logger.Error("Failed to load state", "error", err)
	}

	// For demo purposes, let's create some initial accounts with balances
	_ = app.stateStore.UpdateBalance(1, 1000) // User 1 gets 1000 tokens
	_ = app.stateStore.UpdateBalance(2, 500)  // User 2 gets 500 tokens
	_ = app.stateStore.UpdateBalance(3, 200)  // User 3 gets 200 tokens

	app.logger.Info("Initialized chain", "validators", len(req.Validators))
	return abci.InitChainResponse{}
}

// CheckTx validates a transaction before adding it to the mempool
func (app *Application) CheckTx(req abci.CheckTxRequest) abci.CheckTxResponse {
	// Parse the transaction
	tx, err := types.ParseTransaction(req.Tx)
	if err != nil {
		return abci.CheckTxResponse{
			Code: 1,
			Log:  fmt.Sprintf("Invalid transaction format: %v", err),
		}
	}

	// Validate the transaction
	if err := app.txProcessor.ValidateTransaction(tx); err != nil {
		return abci.CheckTxResponse{
			Code: 2,
			Log:  fmt.Sprintf("Invalid transaction: %v", err),
		}
	}

	// Store the transaction for later processing
	txHash := fmt.Sprintf("%x", req.Tx)
	app.pendingTransactions[txHash] = tx

	return abci.CheckTxResponse{
		Code: 0,
		Log:  "Transaction is valid",
	}
}

// FinalizeBlock processes transactions and updates the application state
func (app *Application) FinalizeBlock(req abci.FinalizeBlockRequest) abci.FinalizeBlockResponse {
	var txResults []*abci.ExecTxResult

	// Process each transaction
	for _, tx := range req.Txs {
		result := app.processTx(tx)
		txResults = append(txResults, result)
	}

	return abci.FinalizeBlockResponse{
		TxResults: txResults,
	}
}

// processTx processes a single transaction
func (app *Application) processTx(txBytes []byte) *abci.ExecTxResult {
	// Parse the transaction
	tx, err := types.ParseTransaction(txBytes)
	if err != nil {
		return &abci.ExecTxResult{
			Code: 1,
			Log:  fmt.Sprintf("Invalid transaction format: %v", err),
		}
	}

	// Process the transaction
	if err := app.txProcessor.ProcessTransaction(tx); err != nil {
		return &abci.ExecTxResult{
			Code: 2,
			Log:  fmt.Sprintf("Failed to process transaction: %v", err),
		}
	}

	// Log the transaction details
	app.logger.Info("Processed transaction",
		"operations", len(tx.Operations),
		"sender", tx.Operations[0].From)

	// Return success
	return &abci.ExecTxResult{
		Code: 0,
		Log:  fmt.Sprintf("Processed %d operations", len(tx.Operations)),
	}
}

// Commit commits the current state and returns a hash of the state
func (app *Application) Commit() abci.CommitResponse {
	// Save state to disk
	if err := app.stateStore.SaveState(); err != nil {
		app.logger.Error("Failed to save state", "error", err)
	}

	// Clear pending transactions
	app.pendingTransactions = make(map[string]*types.Transaction)

	// In a real application, you would compute a state hash here
	return abci.CommitResponse{
		RetainHeight: 0, // Don't prune any heights
	}
}

// Query handles queries to the application state
func (app *Application) Query(req abci.QueryRequest) abci.QueryResponse {
	switch req.Path {
	case "state":
		// Return the entire state
		data, err := app.stateStore.GetState().Serialize()
		if err != nil {
			return abci.QueryResponse{
				Code: 1,
				Log:  fmt.Sprintf("Failed to serialize state: %v", err),
			}
		}
		return abci.QueryResponse{
			Code:  0,
			Value: data,
		}

	case "account":
		// Parse account ID from data
		var accountID int
		if err := json.Unmarshal(req.Data, &accountID); err != nil {
			return abci.QueryResponse{
				Code: 1,
				Log:  fmt.Sprintf("Invalid account ID format: %v", err),
			}
		}

		// Get account
		account := app.stateStore.GetAccount(accountID)
		data, err := json.Marshal(account)
		if err != nil {
			return abci.QueryResponse{
				Code: 2,
				Log:  fmt.Sprintf("Failed to serialize account: %v", err),
			}
		}

		return abci.QueryResponse{
			Code:  0,
			Value: data,
		}

	default:
		return abci.QueryResponse{
			Code: 3,
			Log:  fmt.Sprintf("Unknown query path: %s", req.Path),
		}
	}
}

// RegisterUserKey registers a public key for a user
func (app *Application) RegisterUserKey(userID int, pubKeyBase64 string) error {
	pubKey, exists := app.txProcessor.GetUserKey(userID)
	if exists {
		return fmt.Errorf("user %d already has a registered key", userID)
	}

	// Parse the public key
	var err error
	pubKey, err = crypto.PublicKeyFromBase64(pubKeyBase64)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Register the key
	app.txProcessor.RegisterUserKey(userID, pubKey)
	return nil
}
