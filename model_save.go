package bux

import (
	"context"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/pkg/errors"
)

// Save will Save the model(s) into the Datastore
func Save(ctx context.Context, model ModelInterface) (err error) {

	// Create new Datastore transaction
	// @siggi: we need this to be in a callback context for Mongo
	// NOTE: a DB error is not being returned from here
	return model.Client().Datastore().NewTx(ctx, func(tx *datastore.Transaction) (err error) {

		// Fire the before hooks (parent model)
		if model.IsNew() {
			if err = model.BeforeCreating(ctx); err != nil {
				return
			}
		} else {
			if err = model.BeforeUpdating(ctx); err != nil {
				return
			}
		}

		// Set the record's timestamps
		model.SetRecordTime(model.IsNew())

		// Start the list of models to Save
		modelsToSave := append(make([]ModelInterface, 0), model)

		// Add any child models (fire before hooks)
		if children := model.ChildModels(); len(children) > 0 {
			for _, child := range children {
				if child.IsNew() {
					if err = child.BeforeCreating(ctx); err != nil {
						return
					}
				} else {
					if err = child.BeforeUpdating(ctx); err != nil {
						return
					}
				}

				// Set the record's timestamps
				child.SetRecordTime(child.IsNew())
			}

			// Add to list for saving
			modelsToSave = append(modelsToSave, children...)
		}

		// Logs for saving models
		model.DebugLog(fmt.Sprintf("saving %d models...", len(modelsToSave)))

		// Save all models (or fail!)
		for _, modelToSave := range modelsToSave {
			modelToSave.DebugLog("starting to Save model: " + modelToSave.Name())
			if err = modelToSave.Client().Datastore().SaveModel(
				ctx, modelToSave, tx, modelToSave.IsNew(), false,
			); err != nil {
				return
			}
		}

		// Commit all the model(s) if needed
		if tx.CanCommit() {
			model.DebugLog("committing db transaction...")
			if err = tx.Commit(); err != nil {
				return
			}
		}

		// Fire after hooks (only on commit success)
		var afterErr error
		for _, modelToSave := range modelsToSave {
			if modelToSave.IsNew() {
				modelToSave.NotNew() // NOTE: calling it before this method... after created assumes its been saved already
				afterErr = modelToSave.AfterCreated(ctx)
			} else {
				afterErr = modelToSave.AfterUpdated(ctx)
			}
			if afterErr != nil {
				if err == nil { // First error - set the error
					err = afterErr
				} else { // Got more than one error, wrap it!
					err = errors.Wrap(err, afterErr.Error())
				}
			}
			// modelToSave.NotNew() // NOTE: moved to above from here
		}

		return
	})
}

// saveToCache will Save the model to the cache using the given key
//
// ttl of 0 will cache forever
func saveToCache(ctx context.Context, key string, model ModelInterface, ttl time.Duration) error {
	// NOTE: this check is in place in-case a model does not load it's Parent Client
	if model.Client() != nil {
		c := model.Client().Cachestore()
		if c != nil {
			return c.SetModel(ctx, key, model, ttl)
		}
	}
	model.DebugLog("ignoring SetModel: client or cachestore is missing")
	return nil
}
