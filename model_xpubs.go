package bux

import (
	"context"
	"errors"
	"fmt"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
)

// Xpub is an object representing the BitCoin xPub table
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type Xpub struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID              string `json:"id" toml:"id" yaml:"hash" gorm:"<-:create;type:char(64);primaryKey;comment:This is the sha256(xpub) hash" bson:"_id"`
	CurrentBalance  uint64 `json:"current_balance" toml:"current_balance" yaml:"current_balance" gorm:"<-;comment:The current balance of unspent satoshis" bson:"current_balance"`
	NextInternalNum uint32 `json:"next_internal_num" toml:"next_internal_num" yaml:"next_internal_num" gorm:"<-;type:int;comment:The next index number for the internal xPub derivation" bson:"next_internal_num"`
	NextExternalNum uint32 `json:"next_external_num" toml:"next_external_num" yaml:"next_external_num" gorm:"<-;type:int;comment:The next index number for the external xPub derivation" bson:"next_external_num"`

	destinations []Destination `gorm:"-" bson:"-"` // json:"destinations,omitempty"
}

// newXpub will start a new xPub model
func newXpub(key string, opts ...ModelOps) *Xpub {
	return &Xpub{
		ID:    utils.Hash(key),
		Model: *NewBaseModel(ModelXPub, append(opts, WithXPub(key))...),
	}
}

// newXpub will start a new xPub model
func newXpubUsingID(xPubID string, opts ...ModelOps) *Xpub {
	return &Xpub{
		ID:    xPubID,
		Model: *NewBaseModel(ModelXPub, opts...),
	}
}

// getXpub will get the xPub with the given conditions
func getXpub(ctx context.Context, key string, opts ...ModelOps) (*Xpub, error) {

	// Get the record
	xPub := newXpub(key, opts...)
	if err := Get(
		ctx, xPub, nil, false, defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return xPub, nil
}

// getXpubByID will get the xPub with the given conditions
func getXpubByID(ctx context.Context, xPubID string, opts ...ModelOps) (*Xpub, error) {

	// Get the record
	xPub := newXpubUsingID(xPubID, opts...)
	if err := Get(
		ctx,
		xPub,
		nil,
		false,
		defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return xPub, nil
}

// GetModelName will get the name of the current model
func (m *Xpub) GetModelName() string {
	return ModelXPub.String()
}

// GetModelTableName will get the db table name of the current model
func (m *Xpub) GetModelTableName() string {
	return tableXPubs
}

// Save will Save the model into the Datastore
func (m *Xpub) Save(ctx context.Context) error {
	return Save(ctx, m)
}

// GetID will get the ID
func (m *Xpub) GetID() string {
	return m.ID
}

// getNewDestination will get a new destination, adding to the xpub and incrementing num / address
func (m *Xpub) getNewDestination(ctx context.Context, chain uint32, destinationType string,
	metadata *map[string]interface{}) (*Destination, error) {

	// Check the type
	// todo: support more types of destinations
	if destinationType != utils.ScriptTypePubKeyHash {
		return nil, ErrUnsupportedDestinationType
	}

	// Increment the next num
	num, err := m.IncrementNextNum(ctx, chain)
	if err != nil {
		return nil, err
	}

	// Create the new address
	var destination *Destination
	if destination, err = newAddress(
		m.rawXpubKey, chain, num, m.GetOptions(true)...,
	); err != nil {
		return nil, err
	}

	// Check if metadata is set
	if metadata != nil {
		destination.Metadata = *metadata
	}

	// Add the destination to the xPub
	m.destinations = append(m.destinations, *destination)

	return destination, nil
}

// IncrementBalance will atomically update the balance of the xPub
func (m *Xpub) IncrementBalance(ctx context.Context, balanceIncrement int64) error {
	newBalance, err := IncrementField(ctx, m, currentBalanceField, balanceIncrement)
	if err != nil {
		return err
	}
	m.CurrentBalance = uint64(newBalance)
	return nil
}

// IncrementNextNum will atomically update the num of the given chain of the xPub and return it
func (m *Xpub) IncrementNextNum(ctx context.Context, chain uint32) (uint32, error) {
	var err error
	var newNum int64

	// overwrite the model when incrementing on the DB
	fieldName := nextExternalNumField
	if chain == utils.ChainInternal {
		fieldName = nextInternalNumField
	}

	// Try to increment the field
	incrementXPub := newXpubUsingID(m.ID, m.GetOptions(false)...)
	if newNum, err = IncrementField(
		ctx, incrementXPub, fieldName, 1,
	); err != nil {
		return 0, err
	}

	// return the previous number, which was next num
	return uint32(newNum - 1), err
}

// ChildModels will get any related sub models
func (m *Xpub) ChildModels() (childModels []ModelInterface) {
	for index := range m.destinations {
		childModels = append(childModels, &m.destinations[index])
	}
	return
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *Xpub) BeforeCreating(_ context.Context) error {

	m.DebugLog("starting: [" + m.name.String() + "] BeforeCreating hook...")

	// Validate that the xPub key is correct
	if _, err := utils.ValidateXPub(m.rawXpubKey); err != nil {
		return err
	}

	// Make sure we have an ID
	if len(m.ID) == 0 {
		return ErrMissingFieldID
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *Xpub) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	// todo: run these in go routines?

	// Store in the cache (if enabled)
	if err := saveToCache(
		ctx, fmt.Sprintf("%s-id-%s", m.GetModelName(), m.GetID()), m, 0,
	); err != nil {
		return err
	}

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// AfterUpdated will fire after a successful update into the Datastore
func (m *Xpub) AfterUpdated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterUpdated hook...")

	// Store in the cache (if enabled)
	if err := saveToCache(
		ctx, fmt.Sprintf("%s-id-%s", m.GetModelName(), m.GetID()), m, 0,
	); err != nil {
		return err
	}

	m.DebugLog("end: " + m.Name() + " AfterUpdated hook")
	return nil
}

// Migrate model specific migration on startup
func (m *Xpub) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableXPubs), metadataField)
}
