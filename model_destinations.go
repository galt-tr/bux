package bux

import (
	"context"
	"errors"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
)

// Destination is an object representing the BitCoin destination table
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type Destination struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID            string `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the hash of the locking script" bson:"_id"`
	XpubID        string `json:"xpub_id" toml:"xpub_id" yaml:"xpub_id" gorm:"<-:create;type:char(64);index;comment:This is the related xPub" bson:"xpub_id"`
	LockingScript string `json:"locking_script" toml:"locking_script" yaml:"locking_script" gorm:"<-:create;type:text;comment:This is Bitcoin output script in hex" bson:"locking_script"`
	Type          string `json:"type" toml:"type" yaml:"type" gorm:"<-:create;type:text;comment:Type of output" bson:"type"`
	Chain         uint32 `json:"chain" toml:"chain" yaml:"chain" gorm:"<-:create;type:int;comment:This is the (chain)/num location of the address related to the xPub" bson:"chain"`
	Num           uint32 `json:"num" toml:"num" yaml:"num" gorm:"<-:create;type:int;comment:This is the chain/(num) location of the address related to the xPub" bson:"num"`
	Address       string `json:"address" toml:"address" yaml:"address" gorm:"<-:create;type:varchar(35);index;comment:This is the BitCoin address" bson:"address"`
	DraftID       string `json:"draft_id" toml:"draft_id" yaml:"draft_id" gorm:"<-:create;type:varchar(64);index;comment:This is the related draft id (if internal tx)" bson:"draft_id,omitempty"`
}

// newDestination will start a new Destination model for a locking script
func newDestination(xPubID, lockingScript string, opts ...ModelOps) *Destination {

	// Determine the type if the locking script is provided
	destinationType := ""
	address := ""
	if len(lockingScript) > 0 {
		destinationType = utils.GetDestinationType(lockingScript)
		if destinationType == utils.ScriptTypePubKeyHash {
			// try to extract the address from the locking script
			address, _ = bitcoin.GetAddressFromScript(lockingScript)
		}
	}

	// Return the model
	return &Destination{
		ID:            utils.Hash(lockingScript),
		LockingScript: lockingScript,
		Model:         *NewBaseModel(ModelDestination, opts...),
		Type:          destinationType,
		XpubID:        xPubID,
		Address:       address,
	}
}

// newAddress will start a new Destination model for a legacy Bitcoin address
func newAddress(rawXpubKey string, chain, num uint32, opts ...ModelOps) (*Destination, error) {

	// Create the model
	destination := &Destination{
		Chain: chain,
		Model: *NewBaseModel(ModelDestination, opts...),
		Num:   num,
	}

	// Set the default address
	err := destination.setAddress(rawXpubKey)
	if err != nil {
		return nil, err
	}

	// Set the locking script
	if destination.LockingScript, err = bitcoin.ScriptFromAddress(
		destination.Address,
	); err != nil {
		return nil, err
	}

	// Determine the type if the locking script is provided
	destination.Type = utils.GetDestinationType(destination.LockingScript)
	destination.ID = utils.Hash(destination.LockingScript)

	// Return the destination (address)
	return destination, nil
}

// getDestinationByAddress will get the destination by the given address
func getDestinationByAddress(ctx context.Context, address string, opts ...ModelOps) (*Destination, error) {

	// Construct an empty model
	destination := newDestination("", "", opts...)
	conditions := map[string]interface{}{
		"address": address,
	}

	// Get the record
	if err := Get(ctx, destination, conditions, true, defaultDatabaseReadTimeout); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return destination, nil
}

// getDestinationByLockingScript will get the destination by the given locking script
func getDestinationByLockingScript(ctx context.Context, lockingScript string, opts ...ModelOps) (*Destination, error) {

	// Construct an empty model
	destination := newDestination("", lockingScript, opts...)

	// Get the record
	if err := Get(ctx, destination, nil, true, defaultDatabaseReadTimeout); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return destination, nil
}

// getDestinationsByXpubID will get the destination(s) by the given xPubID
func getDestinationsByXpubID(ctx context.Context, xPubID string, usingMetadata *Metadata,
	pageSize, page int, opts ...ModelOps) ([]*Destination, error) {

	// Construct an empty model
	var models []Destination
	conditions := map[string]interface{}{
		xPubIDField: xPubID,
	}

	if usingMetadata != nil {
		conditions[metadataField] = usingMetadata
	}

	// Get the records
	if err := getModels(
		ctx, NewBaseModel(ModelNameEmpty, opts...).Client().Datastore(),
		&models, conditions, pageSize, page, "", "", defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	// Loop and enrich
	destinations := make([]*Destination, 0)
	for index := range models {
		models[index].enrich(ModelDestination, opts...)
		destinations = append(destinations, &models[index])
	}

	return destinations, nil
}

// GetModelName will get the name of the current model
func (m *Destination) GetModelName() string {
	return ModelDestination.String()
}

// GetModelTableName will get the db table name of the current model
func (m *Destination) GetModelTableName() string {
	return tableDestinations
}

// Save will Save the model into the Datastore
func (m *Destination) Save(ctx context.Context) (err error) {
	return Save(ctx, m)
}

// GetID will get the model ID
func (m *Destination) GetID() string {
	return m.ID
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *Destination) BeforeCreating(_ context.Context) error {

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Set the ID and Type (from LockingScript) (if not set)
	if len(m.LockingScript) > 0 && (len(m.ID) == 0 || len(m.Type) == 0) {
		m.ID = utils.Hash(m.LockingScript)
		m.Type = utils.GetDestinationType(m.LockingScript)
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")

	return nil
}

// setAddress will derive and set the address based on the chain (internal vs external)
func (m *Destination) setAddress(rawXpubKey string) error {

	// Check the xPub
	hdKey, err := utils.ValidateXPub(rawXpubKey)
	if err != nil {
		return err
	}

	// Set the ID
	m.XpubID = utils.Hash(rawXpubKey)

	// Derive the address to ensure it is correct
	var internal, external string
	if external, internal, err = utils.DeriveAddresses(
		hdKey, m.Num,
	); err != nil {
		return err
	}

	if m.Chain == utils.ChainExternal {
		// Set to external
		m.Address = external
	} else {
		// Default is internal
		m.Address = internal
	}

	return nil
}

// Migrate model specific migration on startup
func (m *Destination) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableDestinations), metadataField)
}
