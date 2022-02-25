package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
)

// NewAccessKey will create a new access key for the given xpub
//
// opts are options and can include "metadata"
func (c *Client) NewAccessKey(ctx context.Context, rawXpubKey string, opts ...ModelOps) (*AccessKey, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_access_key")

	// Validate that the value is an xPub
	_, err := utils.ValidateXPub(rawXpubKey)
	if err != nil {
		return nil, err
	}

	// Get the xPub (by key - converts to id)
	var xPub *Xpub
	if xPub, err = getXpub(
		ctx, rawXpubKey, // Pass the context and key everytime (for now)
		c.DefaultModelOptions()..., // Passing down the Datastore and client information into the model
	); err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	// Create the model & set the default options (gives options from client->model)
	accessKey := newAccessKey(
		xPub.ID, c.DefaultModelOptions(append(opts, New())...)...,
	)

	// Save the model
	if err = accessKey.Save(ctx); err != nil {
		return nil, err
	}

	// Return the created model
	return accessKey, nil
}

// GetAccessKey will get an existing access key from the Datastore
func (c *Client) GetAccessKey(ctx context.Context, rawXpubKey, id string) (*AccessKey, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_access_key")

	// Get the access key
	accessKey, err := GetAccessKey(
		ctx, id,
		c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	} else if accessKey == nil {
		return nil, ErrMissingXpub
	}

	// make sure this is the correct accessKey
	if accessKey.XpubID != utils.Hash(rawXpubKey) {
		return nil, utils.ErrXpubNoMatch
	}

	// Return the model
	return accessKey, nil
}

// GetAccessKeys will get all existing access keys from the Datastore
//
// metadataConditions is the metadata to match to the access keys being returned
func (c *Client) GetAccessKeys(ctx context.Context, xPubID string, metadataConditions *Metadata, opts ...ModelOps) ([]*AccessKey, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_access_keys")

	// Get the access key
	accessKeys, err := GetAccessKeys(
		ctx,
		xPubID,
		metadataConditions,
		c.DefaultModelOptions(opts...)...,
	)
	if err != nil {
		return nil, err
	} else if accessKeys == nil {
		return nil, datastore.ErrNoResults
	}

	// Return the models
	return accessKeys, nil
}

// RevokeAccessKey will revoke an access key by its id
//
// opts are options and can include "metadata"
func (c *Client) RevokeAccessKey(ctx context.Context, rawXpubKey, id string, opts ...ModelOps) (*AccessKey, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "new_access_key")

	// Validate that the value is an xPub
	_, err := utils.ValidateXPub(rawXpubKey)
	if err != nil {
		return nil, err
	}

	// Get the xPub (by key - converts to id)
	var xPub *Xpub
	if xPub, err = getXpub(
		ctx, rawXpubKey, // Pass the context and key everytime (for now)
		c.DefaultModelOptions()..., // Passing down the Datastore and client information into the model
	); err != nil {
		return nil, err
	} else if xPub == nil {
		return nil, ErrMissingXpub
	}

	var accessKey *AccessKey
	if accessKey, err = GetAccessKey(
		ctx, id, c.DefaultModelOptions(opts...)...,
	); err != nil {
		return nil, err
	}

	// make sure this is the correct accessKey
	if accessKey.XpubID != utils.Hash(rawXpubKey) {
		return nil, utils.ErrXpubNoMatch
	}

	accessKey.RevokedAt.Valid = true
	accessKey.RevokedAt.Time = time.Now()

	// Save the model
	if err = accessKey.Save(ctx); err != nil {
		return nil, err
	}

	// Return the updated model
	return accessKey, nil
}
