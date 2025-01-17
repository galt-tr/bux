package bux

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math"
	"math/big"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
)

// DraftTransaction is an object representing the draft BitCoin transaction prior to the final transaction
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type DraftTransaction struct {
	// Base model
	Model `bson:",inline"`

	// Standard transaction model base fields
	TransactionBase `bson:",inline"`

	// Model specific fields
	XpubID        string            `json:"xpub_id" toml:"xpub_id" yaml:"xpub_id" gorm:"<-:create;type:char(64);index;comment:This is the related xPub" bson:"xpub_id"`
	ExpiresAt     time.Time         `json:"expires_at" toml:"expires_at" yaml:"expires_at" gorm:"<-:create;comment:Time when the draft expires" bson:"expires_at"`
	Configuration TransactionConfig `json:"configuration" toml:"configuration" yaml:"configuration" gorm:"<-;type:text;comment:This is the configuration struct in JSON" bson:"configuration"`
	Status        DraftStatus       `json:"status" toml:"status" yaml:"status" gorm:"<-;type:varchar(10);index;comment:This is the status of the draft" bson:"status"`
	FinalTxID     string            `json:"final_tx_id,omitempty" toml:"final_tx_id" yaml:"final_tx_id" gorm:"<-;type:char(64);index;comment:This is the final tx ID" bson:"final_tx_id,omitempty"`
}

// newDraftTransaction will start a new draft tx
func newDraftTransaction(rawXpubKey string, config *TransactionConfig, opts ...ModelOps) *DraftTransaction {

	// Random GUID
	id, _ := utils.RandomHex(32)

	// Set the fee (if not found)
	if config.FeeUnit == nil {
		config.FeeUnit = defaultFee
	}

	// Set the expires time (default)
	expiresAt := time.Now().UTC().Add(defaultDraftTxExpiresIn)
	if config.ExpiresIn > 0 {
		expiresAt = time.Now().UTC().Add(config.ExpiresIn)
	}

	return &DraftTransaction{
		TransactionBase: TransactionBase{ID: id},
		Configuration:   *config,
		ExpiresAt:       expiresAt,
		Model: *NewBaseModel(
			ModelDraftTransaction,
			append(opts, WithXPub(rawXpubKey))...,
		),
		Status: DraftStatusDraft,
		XpubID: utils.Hash(rawXpubKey),
	}
}

// getDraftTransactionID will get the draft transaction with the given conditions
func getDraftTransactionID(ctx context.Context, xPubID, id string,
	opts ...ModelOps) (*DraftTransaction, error) {

	// Get the record
	config := &TransactionConfig{}
	conditions := map[string]interface{}{
		xPubIDField: xPubID,
		idField:     id,
		// statusField:      DraftStatusDraft,
	}
	draftTransaction := newDraftTransaction("", config, opts...)
	draftTransaction.ID = "" // newDraftTransaction always sets an ID, need to remove for querying
	if err := Get(ctx, draftTransaction, conditions, false, defaultDatabaseReadTimeout); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return draftTransaction, nil
}

// GetModelName will get the name of the current model
func (m *DraftTransaction) GetModelName() string {
	return ModelDraftTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *DraftTransaction) GetModelTableName() string {
	return tableDraftTransactions
}

// Save will Save the model into the Datastore
func (m *DraftTransaction) Save(ctx context.Context) (err error) {
	if err = Save(ctx, m); err != nil {

		// todo: run in a go routine?
		// un-reserve the utxos
		if utxoErr := UnReserveUtxos(
			ctx, m.XpubID, m.ID, m.GetOptions(false)...,
		); utxoErr != nil {
			err = errors.Wrap(err, utxoErr.Error())
		}
	}
	return
}

// GetID will get the model ID
func (m *DraftTransaction) GetID() string {
	return m.ID
}

// processConfigOutputs will process all the outputs,
// doing any lookups and creating locking scripts
func (m *DraftTransaction) processConfigOutputs(ctx context.Context) error {

	// Get the client
	c := m.Client()

	// Special case where we are sending all funds to a single (address, paymail, handle)
	if len(m.Configuration.SendAllTo) > 0 {
		m.Configuration.Outputs = []*TransactionOutput{{
			To: m.Configuration.SendAllTo,
		}}
		if err := m.Configuration.Outputs[0].processOutput(
			ctx, c.Cachestore(),
			c.PaymailClient(),
			c.PaymailServerConfig().DefaultFromPaymail,
			c.PaymailServerConfig().DefaultNote,
			false,
		); err != nil {
			return err
		}
	} else {
		// Loop all outputs and process
		for index := range m.Configuration.Outputs {

			// Start the output script slice
			if m.Configuration.Outputs[index].Scripts == nil {
				m.Configuration.Outputs[index].Scripts = make([]*ScriptOutput, 0)
			}

			// Process the outputs
			if err := m.Configuration.Outputs[index].processOutput(
				ctx, c.Cachestore(),
				c.PaymailClient(),
				c.PaymailServerConfig().DefaultFromPaymail,
				c.PaymailServerConfig().DefaultNote,
				true,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// createTransactionHex will create the transaction with the given inputs and outputs
func (m *DraftTransaction) createTransactionHex(ctx context.Context) (err error) {

	// Check that we have outputs
	if len(m.Configuration.Outputs) == 0 && m.Configuration.SendAllTo == "" {
		return ErrMissingTransactionOutputs
	}

	// Get the total satoshis needed to make this transaction
	satoshisNeeded := m.getTotalSatoshis()

	// Set opts
	opts := m.GetOptions(false)

	// Process the outputs first
	// if an error occurs in processing the outputs, we have at least not made any reservations yet
	if err = m.processConfigOutputs(ctx); err != nil {
		return
	}

	var inputUtxos *[]*bt.UTXO
	var satoshisReserved uint64
	if m.Configuration.SendAllTo != "" {
		var spendableUtxos []*Utxo
		if spendableUtxos, err = GetSpendableUtxos(
			ctx, m.XpubID, utils.ScriptTypePubKeyHash, m.Configuration.FromUtxos, opts...,
		); err != nil {
			return err
		}
		for _, utxo := range spendableUtxos {
			// Reserve the utxos
			utxo.DraftID.Valid = true
			utxo.DraftID.String = m.ID
			utxo.ReservedAt.Valid = true
			utxo.ReservedAt.Time = time.Now().UTC()

			// Save the UTXO
			if err = utxo.Save(ctx); err != nil {
				return err
			}

			m.Configuration.Outputs[0].Satoshis += utxo.Satoshis
		}

		// Get the inputUtxos (in bt.UTXO format) and the total amount of satoshis from the utxos
		if inputUtxos, satoshisReserved, err = m.getInputsFromUtxos(
			spendableUtxos,
		); err != nil {
			return
		}

		if err = m.processUtxos(
			ctx, spendableUtxos,
		); err != nil {
			return err
		}
	} else {
		// Reserve and Get utxos for the transaction
		var reservedUtxos []*Utxo
		feePerByte := float64(m.Configuration.FeeUnit.Satoshis / m.Configuration.FeeUnit.Bytes)

		reserveSatoshis := satoshisNeeded + m.estimateFee(m.Configuration.FeeUnit) + dustLimit
		if reservedUtxos, err = ReserveUtxos(
			ctx, m.XpubID, m.ID, reserveSatoshis, feePerByte, m.Configuration.FromUtxos, opts...,
		); err != nil {
			return
		}

		// Get the inputUtxos (in bt.UTXO format) and the total amount of satoshis from the utxos
		if inputUtxos, satoshisReserved, err = m.getInputsFromUtxos(
			reservedUtxos,
		); err != nil {
			return
		}

		if err = m.processUtxos(
			ctx, reservedUtxos,
		); err != nil {
			return err
		}
	}

	// start a new transaction from the reservedUtxos
	tx := bt.NewTx()
	if err = tx.FromUTXOs(*inputUtxos...); err != nil {
		return
	}

	// Estimate the fee for the transaction
	fee := m.estimateFee(m.Configuration.FeeUnit)
	if m.Configuration.SendAllTo != "" {
		if m.Configuration.Outputs[0].Satoshis <= dustLimit {
			return ErrOutputValueTooLow
		}

		m.Configuration.Outputs[0].Satoshis -= fee
		m.Configuration.Outputs[0].Scripts[0].Satoshis = m.Configuration.Outputs[0].Satoshis
		m.Configuration.Fee = fee
	} else {
		if satoshisReserved < satoshisNeeded+fee {
			return ErrNotEnoughUtxos
		}

		// if we have a remainder, add that to an output to our own wallet address
		satoshisChange := satoshisReserved - satoshisNeeded - fee
		m.Configuration.Fee = fee
		if satoshisChange > 0 {
			if err = m.setChangeDestination(
				ctx, satoshisChange,
			); err != nil {
				return
			}
		}
	}

	// Add the outputs to the bt transaction
	if err = m.addOutputsToTx(tx); err != nil {
		return
	}

	// Create the final hex (without signatures)
	m.Hex = tx.String()

	return
}

// processUtxos will process the utxos
func (m *DraftTransaction) processUtxos(ctx context.Context, utxos []*Utxo) error {
	// Get destinations
	for _, utxo := range utxos {
		destination, err := getDestinationByLockingScript(
			ctx, utxo.ScriptPubKey, m.GetOptions(false)...,
		)
		if err != nil {
			return err
		}
		m.Configuration.Inputs = append(
			m.Configuration.Inputs, &TransactionInput{
				Utxo:        *utxo,
				Destination: *destination,
			})
	}

	return nil
}

// estimateSize will loop the inputs and outputs and estimate the size of the transaction
func (m *DraftTransaction) estimateSize() uint64 {
	size := defaultOverheadSize
	for _, input := range m.Configuration.Inputs {
		size += utils.GetInputSizeForType(input.Type)
	}
	for _, output := range m.Configuration.Outputs {
		for _, s := range output.Scripts {
			size += utils.GetOutputSize(s.Script)
		}
	}

	/*
		if m.Configuration.NLockTime {
			size += 16
		}
	*/

	return size
}

// estimateFee will loop the inputs and outputs and estimate the required fee
func (m *DraftTransaction) estimateFee(unit *utils.FeeUnit) uint64 {
	size := m.estimateSize()
	return uint64(math.Ceil(float64(size) * (float64(unit.Satoshis) / float64(unit.Bytes))))
}

// addOutputs will add the given outputs to the bt.Tx
func (m *DraftTransaction) addOutputsToTx(tx *bt.Tx) (err error) {
	var s *bscript.Script
	for _, output := range m.Configuration.Outputs {
		for _, sc := range output.Scripts {
			if s, err = bscript.NewFromHexString(
				sc.Script,
			); err != nil {
				return
			}

			if sc.ScriptType == utils.ScriptTypeNullData {
				// op_return output - only one allowed to have 0 satoshi value ???
				if sc.Satoshis > 0 {
					return ErrInvalidOpReturnOutput
				}

				tx.AddOutput(&bt.Output{
					LockingScript: s,
					Satoshis:      0,
				})
			} else {
				// sending to a p2pkh
				if sc.Satoshis == 0 {
					return ErrOutputValueTooLow
				}

				if err = tx.AddP2PKHOutputFromScript(
					s, sc.Satoshis,
				); err != nil {
					return
				}
			}
		}
	}
	return
}

// setChangeDestination will make a new change destination
func (m *DraftTransaction) setChangeDestination(ctx context.Context, satoshisChange uint64) error {

	m.Configuration.ChangeSatoshis = satoshisChange

	numberOfDestinations := m.Configuration.ChangeNumberOfDestinations
	if numberOfDestinations <= 0 {
		numberOfDestinations = 1 // todo get from config
	}
	minimumSatoshis := m.Configuration.ChangeMinimumSatoshis
	if minimumSatoshis <= 0 { // todo: protect against un-spendable amount? less than fee to miner for min tx?
		minimumSatoshis = 1250 // todo get from config
	}

	if float64(satoshisChange)/float64(numberOfDestinations) < float64(minimumSatoshis) {
		// we cannot split our change to the number of destinations given, re-calc
		numberOfDestinations = 1
	}

	if m.Configuration.ChangeDestinations == nil {
		if err := m.setChangeDestinations(
			ctx, numberOfDestinations,
		); err != nil {
			return err
		}
	}

	changeSatoshis, err := m.getChangeSatoshis(satoshisChange)
	if err != nil {
		return err
	}

	for _, destination := range m.Configuration.ChangeDestinations {
		m.Configuration.Outputs = append(m.Configuration.Outputs, &TransactionOutput{
			To: destination.Address,
			Scripts: []*ScriptOutput{{
				Address:  destination.Address,
				Satoshis: changeSatoshis[destination.LockingScript],
				Script:   destination.LockingScript,
			}},
			Satoshis: changeSatoshis[destination.LockingScript],
		})
	}

	return nil
}

// split the change satoshis amongst the change destinations according to the strategy given in config
func (m *DraftTransaction) getChangeSatoshis(satoshisChange uint64) (changeSatoshis map[string]uint64, err error) {

	changeSatoshis = make(map[string]uint64)
	var lastDestination string
	changeUsed := uint64(0)

	if m.Configuration.ChangeDestinationsStrategy == ChangeStrategyNominations {
		return nil, ErrChangeStrategyNotImplemented
	} else if m.Configuration.ChangeDestinationsStrategy == ChangeStrategyRandom {
		nDestinations := float64(len(m.Configuration.ChangeDestinations))
		var a *big.Int
		for _, destination := range m.Configuration.ChangeDestinations {
			if a, err = rand.Int(
				rand.Reader, big.NewInt(math.MaxInt64),
			); err != nil {
				return
			}
			randomChange := (((float64(a.Int64()) / (1 << 63)) * 50) + 75) / 100
			changeForDestination := uint64(randomChange * float64(satoshisChange) / nDestinations)

			changeSatoshis[destination.LockingScript] = changeForDestination
			lastDestination = destination.LockingScript
			changeUsed += changeForDestination
		}
	} else {
		// default
		changePerDestination := uint64(float64(satoshisChange) / float64(len(m.Configuration.ChangeDestinations)))
		for _, destination := range m.Configuration.ChangeDestinations {
			changeSatoshis[destination.LockingScript] = changePerDestination
			lastDestination = destination.LockingScript
			changeUsed += changePerDestination
		}
	}

	// handle remainder
	changeSatoshis[lastDestination] += satoshisChange - changeUsed

	return
}

// setChangeDestinations will set the change destinations based on the number
func (m *DraftTransaction) setChangeDestinations(ctx context.Context, numberOfDestinations int) error {

	// Set the options
	opts := m.GetOptions(false)
	optsNew := append(opts, New())

	var err error
	var xPub *Xpub
	var num uint32

	// Loop for each destination
	for i := 0; i < numberOfDestinations; i++ {
		if xPub, err = getXpub(
			ctx, m.rawXpubKey, opts...,
		); err != nil {
			return err
		} else if xPub == nil {
			return ErrMissingXpub
		}

		if num, err = xPub.IncrementNextNum(
			ctx, utils.ChainInternal,
		); err != nil {
			return err
		}

		var destination *Destination
		if destination, err = newAddress(
			m.rawXpubKey, utils.ChainInternal, num, optsNew...,
		); err != nil {
			return err
		}

		destination.DraftID = m.ID

		if err = destination.Save(ctx); err != nil {
			return err
		}

		m.Configuration.ChangeDestinations = append(m.Configuration.ChangeDestinations, destination)
	}

	return nil
}

// getInputsFromUtxos this function transforms bux utxos to bt.UTXOs
func (m *DraftTransaction) getInputsFromUtxos(reservedUtxos []*Utxo) (*[]*bt.UTXO, uint64, error) {
	// transform to bt.utxo and check if we have enough
	inputUtxos := new([]*bt.UTXO)
	satoshisReserved := uint64(0)
	var lockingScript *bscript.Script
	var err error
	for _, utxo := range reservedUtxos {

		if lockingScript, err = bscript.NewFromHexString(
			utxo.ScriptPubKey,
		); err != nil {
			return nil, 0, errors.Wrap(ErrInvalidLockingScript, err.Error())
		}

		var txIDBytes []byte
		if txIDBytes, err = hex.DecodeString(
			utxo.TransactionID,
		); err != nil {
			return nil, 0, errors.Wrap(ErrInvalidTransactionID, err.Error())
		}

		*inputUtxos = append(*inputUtxos, &bt.UTXO{
			TxID:           txIDBytes,
			Vout:           utxo.OutputIndex,
			Satoshis:       utxo.Satoshis,
			LockingScript:  lockingScript,
			SequenceNumber: bt.DefaultSequenceNumber,
		})
		satoshisReserved += utxo.Satoshis
	}

	return inputUtxos, satoshisReserved, nil
}

// getTotalSatoshis calculate the total satoshis of all outputs
func (m *DraftTransaction) getTotalSatoshis() (satoshis uint64) {
	for _, output := range m.Configuration.Outputs {
		satoshis += output.Satoshis
	}
	return
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *DraftTransaction) BeforeCreating(ctx context.Context) (err error) {

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Prepare the transaction
	if err = m.createTransactionHex(ctx); err != nil {
		return
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return
}

// AfterUpdated will fire after a successful update into the Datastore
func (m *DraftTransaction) AfterUpdated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterUpdated hook...")

	// todo: run these in go routines?

	// remove reservation from all utxos related to this draft transaction
	if m.Status == DraftStatusCanceled || m.Status == DraftStatusExpired {
		utxos, err := getUtxosByDraftID(
			ctx, m.ID,
			0, 0,
			"", "",
			m.GetOptions(false)...,
		)
		if err != nil {
			return err
		}
		for index := range utxos {
			utxos[index].DraftID.String = ""
			utxos[index].DraftID.Valid = false
			utxos[index].ReservedAt.Time = time.Time{}
			utxos[index].ReservedAt.Valid = false
			if err = utxos[index].Save(ctx); err != nil {
				return err
			}
		}
	}

	m.DebugLog("end: " + m.Name() + " AfterUpdated hook")
	return nil
}

// RegisterTasks will register the model specific tasks on client initialization
func (m *DraftTransaction) RegisterTasks() error {

	// No task manager loaded?
	tm := m.Client().Taskmanager()
	if tm == nil {
		return nil
	}

	// Register the task locally (cron task - set the defaults)
	cleanUpTask := m.Name() + "_clean_up"
	ctx := context.Background()

	// Register the task
	if err := tm.RegisterTask(&taskmanager.Task{
		Name:       cleanUpTask,
		RetryLimit: 1,
		Handler: func(client *Client) error {
			if taskErr := TaskCleanupDraftTransactions(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+cleanUpTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	return tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(cleanUpTask),
		TaskName:       cleanUpTask,
	})
}

// Migrate model specific migration on startup
func (m *DraftTransaction) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableDraftTransactions), metadataField)
}
