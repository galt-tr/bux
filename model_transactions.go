package bux

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
)

// TransactionBase is the same fields share between multiple transaction models
type TransactionBase struct {
	ID  string `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique id (hash of the transaction hex)" bson:"_id"`
	Hex string `json:"hex" toml:"hex" yaml:"hex" gorm:"<-:create;type:text;comment:This is the raw transaction hex" bson:"hex"`

	// Private for internal use
	parsedTx *bt.Tx `gorm:"-" bson:"-"` // The go-bt version of the transaction
}

// TransactionDirection String describing the direction of the transaction (in / out)
type TransactionDirection string

const (
	// TransactionDirectionIn The transaction is coming in to the wallet of the xpub
	TransactionDirectionIn TransactionDirection = "incoming"

	// TransactionDirectionOut The transaction is going out of to the wallet of the xpub
	TransactionDirectionOut TransactionDirection = "outgoing"

	// TransactionDirectionReconcile The transaction is an internal reconciliation transaction
	TransactionDirectionReconcile TransactionDirection = "reconcile"
)

// Transaction is an object representing the BitCoin transaction table
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type Transaction struct {
	// Base model
	Model `bson:",inline"`

	// Standard transaction model base fields
	TransactionBase `bson:",inline"`

	// Model specific fields
	XpubInIDs       IDs             `json:"xpub_in_ids,omitempty" toml:"xpub_in_ids" yaml:"xpub_in_ids" gorm:"<-:create;type:json" bson:"xpub_in_ids,omitempty"`
	XpubOutIDs      IDs             `json:"xpub_out_ids,omitempty" toml:"xpub_out_ids" yaml:"xpub_out_ids" gorm:"<-:create;type:json" bson:"xpub_out_ids,omitempty"`
	BlockHash       string          `json:"block_hash" toml:"block_hash" yaml:"block_hash" gorm:"<-;type:char(64);comment:This is the related block when the transaction was mined" bson:"block_hash,omitempty"`
	BlockHeight     uint64          `json:"block_height" toml:"block_height" yaml:"block_height" gorm:"<-;type:bigint;comment:This is the related block when the transaction was mined" bson:"block_height,omitempty"`
	Fee             uint64          `json:"fee" toml:"fee" yaml:"fee" gorm:"<-create;type:bigint" bson:"fee,omitempty"`
	NumberOfInputs  uint32          `json:"number_of_inputs" toml:"number_of_inputs" yaml:"number_of_inputs" gorm:"<-create;type:int" bson:"number_of_inputs,omitempty"`
	NumberOfOutputs uint32          `json:"number_of_outputs" toml:"number_of_outputs" yaml:"number_of_outputs" gorm:"<-create;type:int" bson:"number_of_outputs,omitempty"`
	DraftID         string          `json:"draft_id" toml:"draft_id" yaml:"draft_id" gorm:"<-create;type:varchar(64);index;comment:This is the related draft id" bson:"draft_id,omitempty"`
	TotalValue      uint64          `json:"total_value" toml:"total_value" yaml:"total_value" gorm:"<-create;type:bigint" bson:"total_value,omitempty"`
	XpubMetadata    XpubMetadata    `json:"-" toml:"xpub_metadata" gorm:"<-;type:json;xpub_id specific metadata" bson:"xpub_metadata,omitempty"`
	XpubOutputValue XpubOutputValue `json:"-" toml:"xpub_output_value" gorm:"<-create;type:json;xpub_id specific value" bson:"xpub_output_value,omitempty"`

	// Virtual Fields
	OutputValue int64                `json:"-" toml:"-" bson:"-,omitempty"`
	Status      SyncStatus           `json:"-" toml:"-" yaml:"-" gorm:"-" bson:"-"`
	Direction   TransactionDirection `json:"-" toml:"-" yaml:"-" gorm:"-" bson:"-"`
	// Confirmations  uint64       `json:"-" toml:"-" yaml:"-" gorm:"-" bson:"-"`

	// Private for internal use
	draftTransaction   *DraftTransaction    `gorm:"-" bson:"-"` // Related draft transaction for processing and recording
	syncTransaction    *SyncTransaction     `gorm:"-" bson:"-"` // Related record if broadcast config is detected (create new recordNew)
	transactionService transactionInterface `gorm:"-" bson:"-"` // Used for interfacing methods
	utxos              []Utxo               `gorm:"-" bson:"-"` // json:"destinations,omitempty"
	xPubID             string               `gorm:"-" bson:"-"` // XPub of the user registering this transaction
	inputUtxoChecksOff bool                 `gorm:"-" bson:"-"` // Whether to turn off the checks of the input utxos that match a draft transaction
}

// newTransactionBase creates the standard transaction model base
func newTransactionBase(hex string, opts ...ModelOps) *Transaction {
	return &Transaction{
		TransactionBase: TransactionBase{
			Hex: hex,
		},
		Model:              *NewBaseModel(ModelTransaction, opts...),
		Status:             statusComplete,
		transactionService: transactionService{},
		XpubOutputValue:    map[string]int64{},
	}
}

// newTransaction will start a new transaction model
func newTransaction(txHex string, opts ...ModelOps) (tx *Transaction) {
	tx = newTransactionBase(txHex, opts...)

	// Set the ID
	if len(tx.Hex) > 0 {
		_ = tx.setID()
	}

	// Set xPub ID
	tx.setXPubID()

	return
}

// newTransactionWithDraftID will start a new transaction model and set the draft ID
func newTransactionWithDraftID(txHex, draftID string, opts ...ModelOps) (tx *Transaction) {
	tx = newTransaction(txHex, opts...)
	tx.DraftID = draftID
	return
}

// newTransactionFromIncomingTransaction will start a new transaction model using an incomingTx
func newTransactionFromIncomingTransaction(incomingTx *IncomingTransaction) *Transaction {

	// Create the base
	tx := newTransactionBase(incomingTx.Hex, incomingTx.GetOptions(true)...)
	tx.TransactionBase.parsedTx = incomingTx.TransactionBase.parsedTx
	tx.rawXpubKey = incomingTx.rawXpubKey
	tx.setXPubID()

	// Set the fields
	tx.NumberOfOutputs = uint32(len(tx.TransactionBase.parsedTx.Outputs))
	tx.NumberOfInputs = uint32(len(tx.TransactionBase.parsedTx.Inputs))
	tx.Status = statusProcessing

	// Set the ID (run the same method)
	if len(tx.Hex) > 0 {
		_ = tx.setID()
	}

	return tx
}

// getTransactionByID will get the model from a given transaction ID
func getTransactionByID(ctx context.Context, rawXpubKey, txID string, opts ...ModelOps) (*Transaction, error) {

	// Construct an empty tx
	tx := newTransaction("", opts...)
	tx.rawXpubKey = rawXpubKey
	tx.ID = txID

	// Get the record
	if err := Get(ctx, tx, nil, false, 30*time.Second); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return tx, nil
}

// setXPubID will set the xPub ID on the model
func (m *Transaction) setXPubID() {
	if len(m.rawXpubKey) > 0 && len(m.xPubID) == 0 {
		m.xPubID = utils.Hash(m.rawXpubKey)
	}
}

// getTransactionsByXpubID will get all the models for a given xpub ID
func getTransactionsByXpubID(ctx context.Context, rawXpubKey string,
	metadata *Metadata, conditions *map[string]interface{},
	pageSize, page int, opts ...ModelOps) ([]*Transaction, error) {

	xPubID := utils.Hash(rawXpubKey)
	dbConditions := map[string]interface{}{
		"$or": []map[string]interface{}{{
			"xpub_in_ids": xPubID,
		}, {
			"xpub_out_ids": xPubID,
		}},
	}

	// check for direction query
	if conditions != nil && (*conditions)["direction"] != nil {
		direction := (*conditions)["direction"].(string)
		if direction == string(TransactionDirectionIn) {
			dbConditions["xpub_output_value"] = map[string]interface{}{
				xPubID: map[string]interface{}{
					"$gt": 0,
				},
			}
		} else if direction == string(TransactionDirectionOut) {
			dbConditions["xpub_output_value"] = map[string]interface{}{
				xPubID: map[string]interface{}{
					"$lt": 0,
				},
			}
		} else if direction == string(TransactionDirectionReconcile) {
			dbConditions["xpub_output_value"] = map[string]interface{}{
				xPubID: 0,
			}
		}
		delete(*conditions, "direction")
	}

	if metadata != nil && len(*metadata) > 0 {
		and := make([]map[string]interface{}, 0)
		if _, ok := dbConditions["$and"]; ok {
			and = dbConditions["$and"].([]map[string]interface{})
		}
		for key, value := range *metadata {
			condition := map[string]interface{}{
				"$or": []map[string]interface{}{{
					metadataField: Metadata{
						key: value,
					},
				}, {
					xPubMetadataField: map[string]interface{}{
						xPubID: Metadata{
							key: value,
						},
					},
				}},
			}
			and = append(and, condition)
		}
		dbConditions["$and"] = and
	}

	if conditions != nil && len(*conditions) > 0 {
		and := make([]map[string]interface{}, 0)
		if _, ok := dbConditions["$and"]; ok {
			and = dbConditions["$and"].([]map[string]interface{})
		}
		and = append(and, *conditions)
		dbConditions["$and"] = and
	}

	return _getTransactions(ctx, dbConditions, page, pageSize, opts...)
}

// _getTransactions get all transactions for the given conditions
// NOTE: this function should only be used internally, it allows to query the whole transaction table
func _getTransactions(ctx context.Context, conditions map[string]interface{}, page, pageSize int, opts ...ModelOps) ([]*Transaction, error) {
	var models []Transaction
	if err := getModels(
		ctx, NewBaseModel(
			ModelNameEmpty, opts...).Client().Datastore(),
		&models, conditions,
		pageSize, page,
		"", "", defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	// Loop and enrich
	transactions := make([]*Transaction, 0)
	for index := range models {
		models[index].enrich(ModelTransaction, opts...)
		models[index].setXPubID()
		tx := &models[index]
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// InputUtxoChecksOff will set the inputUtxoChecksOff option on the transaction
func (m *Transaction) InputUtxoChecksOff(check bool) {
	m.inputUtxoChecksOff = check
}

// GetModelName will get the name of the current model
func (m *Transaction) GetModelName() string {
	return ModelTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *Transaction) GetModelTableName() string {
	return tableTransactions
}

// Save will Save the model into the Datastore
func (m *Transaction) Save(ctx context.Context) (err error) {

	// Prepare the metadata
	if len(m.Metadata) > 0 {
		// set the metadata to be xpub specific, but only if we have a valid xpub ID
		if m.xPubID != "" {
			// was metadata set via opts ?
			if m.XpubMetadata == nil {
				m.XpubMetadata = make(XpubMetadata)
			}
			if _, ok := m.XpubMetadata[m.xPubID]; !ok {
				m.XpubMetadata[m.xPubID] = make(Metadata)
			}
			for key, value := range m.Metadata {
				m.XpubMetadata[m.xPubID][key] = value
			}
			// todo will this overwrite the global metadata ?
			m.Metadata = nil
		}
	}

	return Save(ctx, m)
}

// GetID will get the ID
func (m *Transaction) GetID() string {
	return m.ID
}

// setID will set the ID from the transaction hex
func (m *Transaction) setID() (err error) {

	// Parse the hex (if not already parsed)
	if m.TransactionBase.parsedTx == nil {
		if m.TransactionBase.parsedTx, err = bt.NewTxFromString(m.Hex); err != nil {
			return
		}
	}

	// Set the true transaction ID
	m.ID = m.TransactionBase.parsedTx.TxID()

	return
}

// getValue calculates the value of the transaction
func (m *Transaction) getValues() (outputValue uint64, fee uint64) {

	// Parse the outputs
	for _, output := range m.TransactionBase.parsedTx.Outputs {
		outputValue += output.Satoshis
	}

	// Remove the "change" from the transaction if found
	// todo: this will NOT work for an "external" tx that is coming into our system?
	if m.draftTransaction != nil {
		outputValue -= m.draftTransaction.Configuration.ChangeSatoshis
		fee = m.draftTransaction.Configuration.Fee
	} else { // external transaction

		// todo: missing inputs in some tests?
		var inputValue uint64
		for _, input := range m.TransactionBase.parsedTx.Inputs {
			inputValue += input.PreviousTxSatoshis
		}

		if inputValue > 0 {
			fee = inputValue - outputValue
			outputValue -= fee
		}

		// todo: outputs we know are accumulated
	}

	// remove the fee from the value
	if outputValue > fee {
		outputValue -= fee
	}

	return
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *Transaction) BeforeCreating(ctx context.Context) error {

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Test for required field(s)
	if len(m.Hex) == 0 {
		return ErrMissingFieldHex
	}

	// Set the xPubID
	m.setXPubID()

	// Set the ID - will also parse and verify the tx
	err := m.setID()
	if err != nil {
		return err
	}

	// 	m.xPubID is the xpub of the user registering the transaction
	if m.xPubID != "" && m.DraftID != "" {

		// Only get the draft if we haven't already
		if m.draftTransaction == nil {
			if m.draftTransaction, err = getDraftTransactionID(
				ctx, m.xPubID, m.DraftID, m.GetOptions(false)...,
			); err != nil {
				return err
			} else if m.draftTransaction == nil {
				return ErrDraftNotFound
			}
		}
	}

	// Validations and broadcast config check
	if m.draftTransaction != nil {

		// Do we have a broadcast config? Create the new record
		if m.draftTransaction.Configuration.Sync != nil {
			m.syncTransaction = newSyncTransaction(
				m.GetID(),
				m.draftTransaction.Configuration.Sync,
				m.GetOptions(true)...,
			)
		}
	}

	// If we are external and the user disabled incoming transaction checking, check outputs
	if m.isExternal() && !m.Client().IsITCEnabled() {
		// Check that the transaction has >= 1 known destination
		if !m.TransactionBase.hasOneKnownDestination(ctx, m.GetOptions(false)...) {
			return ErrNoMatchingOutputs
		}
	}

	// Process the UTXOs
	if err = m.processUtxos(ctx); err != nil {
		return err
	}

	// Set the values from the inputs/outputs and draft tx
	m.TotalValue, m.Fee = m.getValues()

	// Add values if found
	if m.TransactionBase.parsedTx != nil {
		m.NumberOfInputs = uint32(len(m.TransactionBase.parsedTx.Inputs))
		m.NumberOfOutputs = uint32(len(m.TransactionBase.parsedTx.Outputs))
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *Transaction) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	// Pre-build the options
	opts := m.GetOptions(false)

	// todo: run these in go routines?

	// update the xpub balances
	for xPubID, balance := range m.XpubOutputValue {
		// todo move this into a function on the xpub model
		// todo: turn this in to job/task to run? (go routine)
		xPub, err := getXpubByID(ctx, xPubID, opts...)
		if err != nil {
			return err
		} else if xPub == nil {
			return ErrMissingRequiredXpub
		}
		err = xPub.IncrementBalance(ctx, balance)
		if err != nil {
			return err
		}
	}

	// update the draft transaction (if linked to reference) to complete
	if m.draftTransaction != nil {
		m.draftTransaction.Status = DraftStatusComplete
		m.draftTransaction.FinalTxID = m.ID
		err := m.draftTransaction.Save(ctx)
		if err != nil {
			m.DebugLog("error updating draft transaction: " + err.Error())
		}
	}

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// ChildModels will get any related sub models
func (m *Transaction) ChildModels() (childModels []ModelInterface) {

	// Add the UTXOs if found
	for index := range m.utxos {
		childModels = append(childModels, &m.utxos[index])
	}

	// Add the broadcast transaction record
	if m.syncTransaction != nil {
		childModels = append(childModels, m.syncTransaction)
	}

	return
}

// processUtxos will process the inputs and outputs for UTXOs
func (m *Transaction) processUtxos(ctx context.Context) error {

	if err := m.processInputs(ctx); err != nil {
		return err
	}

	return m.processOutputs(ctx)
}

// processTxOutputs will process the transaction outputs
func (m *Transaction) processOutputs(ctx context.Context) (err error) {

	// Pre-build the options
	opts := m.GetOptions(false)
	newOpts := append(opts, New())
	var destination *Destination

	// check all the outputs for a known destination
	numberOfOutputsProcessed := 0
	for index := range m.TransactionBase.parsedTx.Outputs {
		amount := m.TransactionBase.parsedTx.Outputs[index].Satoshis
		lockingScript := m.TransactionBase.parsedTx.Outputs[index].LockingScript.String()

		// only save outputs with a satoshi value attached to it
		if amount > 0 {

			// only Save utxos for known destinations
			// todo: optimize this SQL SELECT by requesting all the scripts at once (vs in this loop)
			if destination, err = m.transactionService.getDestinationByLockingScript(
				ctx, lockingScript, opts...,
			); err != nil {
				return
			} else if destination != nil {

				// Add value of output to xPub ID
				if _, ok := m.XpubOutputValue[destination.XpubID]; !ok {
					m.XpubOutputValue[destination.XpubID] = 0
				}
				m.XpubOutputValue[destination.XpubID] += int64(amount)

				// Append the UTXO model
				m.utxos = append(m.utxos, *newUtxo(
					destination.XpubID, m.ID, lockingScript, uint32(index),
					amount, newOpts...,
				))

				// Add the xPub ID
				if !utils.StringInSlice(destination.XpubID, m.XpubOutIDs) {
					m.XpubOutIDs = append(m.XpubOutIDs, destination.XpubID)
				}

				numberOfOutputsProcessed++
			}
		}
	}

	/*
		if m.isExternal() && numberOfOutputsProcessed == 0 {
			// ERR BS transaction
			fmt.Println("finish this")
		} else if !m.isExternal() {
			// check outputs of draft vs current model (excluding change and fee)
			// Change addresses should still be internal addresses
			fmt.Println("finish this")
		}
	*/

	return
}

func (m *Transaction) isExternal() bool {
	return m.draftTransaction == nil
}

// processTxInputs will process the transaction inputs
func (m *Transaction) processInputs(ctx context.Context) (err error) {

	// Pre-build the options
	opts := m.GetOptions(false)

	var utxo *Utxo

	// check whether we are spending an internal utxo
	for index := range m.TransactionBase.parsedTx.Inputs {

		// todo: optimize this SQL SELECT to get all utxos in one query?
		if utxo, err = m.transactionService.getUtxo(ctx,
			hex.EncodeToString(m.TransactionBase.parsedTx.Inputs[index].PreviousTxID()),
			m.TransactionBase.parsedTx.Inputs[index].PreviousTxOutIndex,
			opts...,
		); err != nil {
			return
		} else if utxo != nil {

			isSpent := len(utxo.SpendingTxID.String) > 0
			if isSpent {
				return ErrUtxoAlreadySpent
			}

			if !m.inputUtxoChecksOff {
				// check whether the utxo is spent
				isReserved := len(utxo.DraftID.String) > 0
				matchesDraft := m.draftTransaction != nil && utxo.DraftID.String == m.draftTransaction.ID

				// Check whether the spending transaction was reserved by the draft transaction (in the utxo)
				if !isReserved {
					return ErrUtxoNotReserved
				}
				if !matchesDraft {
					return ErrDraftIDMismatch
				}
			}

			// Update the output value
			if _, ok := m.XpubOutputValue[utxo.XpubID]; !ok {
				m.XpubOutputValue[utxo.XpubID] = 0
			}
			m.XpubOutputValue[utxo.XpubID] -= int64(utxo.Satoshis)

			// Mark utxo as spent
			utxo.SpendingTxID.Valid = true
			utxo.SpendingTxID.String = m.ID
			m.utxos = append(m.utxos, *utxo)

			// Add the xPub ID
			if !utils.StringInSlice(utxo.XpubID, m.XpubInIDs) {
				m.XpubInIDs = append(m.XpubInIDs, utxo.XpubID)
			}
		}
	}

	return
}

// IsXpubAssociated will check if this key is associated to this transaction
func (m *Transaction) IsXpubAssociated(rawXpubKey string) bool {

	// Hash the raw key
	xPubID := utils.Hash(rawXpubKey)
	if len(xPubID) == 0 {
		return false
	}

	// On the input side
	for _, id := range m.XpubInIDs {
		if id == xPubID {
			return true
		}
	}

	// On the output side
	for _, id := range m.XpubOutIDs {
		if id == xPubID {
			return true
		}
	}
	return false
}

// Display filter the model for display
func (m *Transaction) Display() interface{} {

	// In case it was not set
	m.setXPubID()

	if len(m.XpubMetadata) > 0 && len(m.XpubMetadata[m.xPubID]) > 0 {
		if m.Metadata == nil {
			m.Metadata = make(Metadata)
		}
		for key, value := range m.XpubMetadata[m.xPubID] {
			m.Metadata[key] = value
		}
	}

	m.OutputValue = int64(0)
	if len(m.XpubOutputValue) > 0 && m.XpubOutputValue[m.xPubID] != 0 {
		m.OutputValue = m.XpubOutputValue[m.xPubID]
	}

	if m.OutputValue > 0 {
		m.Direction = TransactionDirectionIn
	} else {
		m.Direction = TransactionDirectionOut
	}

	m.XpubInIDs = nil
	m.XpubOutIDs = nil
	m.XpubMetadata = nil
	m.XpubOutputValue = nil
	return m
}

// Migrate model specific migration on startup
func (m *Transaction) Migrate(client datastore.ClientInterface) error {

	tableName := client.GetTableName(tableTransactions)
	if client.Engine() == datastore.MySQL {
		if err := m.migrateMySQL(client, tableName); err != nil {
			return err
		}
	} else if client.Engine() == datastore.PostgreSQL {
		if err := m.migratePostgreSQL(client, tableName); err != nil {
			return err
		}
	}

	return client.IndexMetadata(tableName, xPubMetadataField)
}

// migratePostgreSQL is specific migration SQL for Postgresql
func (m *Transaction) migratePostgreSQL(client datastore.ClientInterface, tableName string) error {

	tx := client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_in_ids ON ` +
		tableName + ` USING gin (xpub_in_ids jsonb_ops)`)
	if tx.Error != nil {
		return tx.Error
	}

	if tx = client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_out_ids ON ` +
		tableName + ` USING gin (xpub_out_ids jsonb_ops)`); tx.Error != nil {
		return tx.Error
	}

	return nil
}

// migrateMySQL is specific migration SQL for MySQL
func (m *Transaction) migrateMySQL(client datastore.ClientInterface, tableName string) error {

	idxName := "idx_" + tableName + "_xpub_in_ids"
	idxExists, err := client.IndexExists(tableName, idxName)
	if err != nil {
		return err
	}
	if !idxExists {
		tx := client.Execute("ALTER TABLE `" + tableName + "`" +
			" ADD INDEX " + idxName + " ( (CAST(xpub_in_ids AS CHAR(64) ARRAY)) )")
		if tx.Error != nil {
			m.Client().Logger().Error(context.Background(), "failed creating json index on mysql: "+tx.Error.Error())
			return nil // nolint: nilerr // error is not needed
		}
	}

	idxName = "idx_" + tableName + "_xpub_out_ids"
	if idxExists, err = client.IndexExists(
		tableName, idxName,
	); err != nil {
		return err
	}
	if !idxExists {
		tx := client.Execute("ALTER TABLE `" + tableName + "`" +
			" ADD INDEX " + idxName + " ( (CAST(xpub_out_ids AS CHAR(64) ARRAY)) )")
		if tx.Error != nil {
			m.Client().Logger().Error(context.Background(), "failed creating json index on mysql: "+tx.Error.Error())
			return nil // nolint: nilerr // error is not needed
		}
	}

	return nil
}

// hasOneKnownDestination will check if the transaction has at least one known destination
//
// This is used to validate if an external transaction should be recorded into the engine
func (m *TransactionBase) hasOneKnownDestination(ctx context.Context, opts ...ModelOps) bool {

	// todo: this can be optimized searching X records at a time vs loop->query->loop->query
	lockingScript := ""
	for index := range m.parsedTx.Outputs {
		lockingScript = m.parsedTx.Outputs[index].LockingScript.String()
		destination, err := getDestinationByLockingScript(ctx, lockingScript, opts...)
		if err != nil {
			destination = newDestination("", lockingScript, opts...)
			destination.Client().Logger().Error(ctx, "error getting destination: "+err.Error())
		} else if destination != nil && destination.LockingScript == lockingScript {
			return true
		}
	}
	return false
}
