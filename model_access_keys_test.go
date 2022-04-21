package bux

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newAccessKey(t *testing.T) {
	t.Run("valid access key", func(t *testing.T) {
		key := newAccessKey(testXPubID)
		require.NotNil(t, key)
		assert.Equal(t, testXPubID, key.XpubID)
		assert.Equal(t, ModelAccessKey.String(), key.GetModelName())
		assert.Equal(t, 64, len(key.GetID()))
		assert.Equal(t, 64, len(key.Key))

		privateKey, _ := bitcoin.PrivateKeyFromString(key.Key)
		assert.IsType(t, bec.PrivateKey{}, *privateKey)
		publicKey := privateKey.PubKey()
		assert.IsType(t, bec.PublicKey{}, *publicKey)
		id := utils.Hash(hex.EncodeToString(publicKey.SerialiseCompressed()))
		assert.Equal(t, id, key.ID)
	})

	t.Run("save", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		key := newAccessKey(testXPubID, append(
			client.DefaultModelOptions(),
			New(),
			WithMetadatas(Metadata{
				"test-key": "test-value",
			}),
		)...)
		assert.Equal(t, 64, len(key.Key))
		err := key.Save(ctx)
		id := key.ID
		require.NoError(t, err)

		var accessKey *AccessKey
		accessKey, err = GetAccessKey(ctx, id, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Equal(t, id, accessKey.ID)
		assert.Equal(t, testXPubID, accessKey.XpubID)
		assert.Equal(t, "", accessKey.Key) // private key is lost after Save
		assert.Len(t, accessKey.Metadata, 1)
		assert.Equal(t, "test-value", accessKey.Metadata["test-key"])
	})

	t.Run("revoke", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()

		key := newAccessKey(testXPubID, append(client.DefaultModelOptions(), New())...)
		assert.Equal(t, 64, len(key.Key))
		err := key.Save(ctx)
		id := key.ID
		require.NoError(t, err)

		var accessKey *AccessKey
		accessKey, err = GetAccessKey(ctx, id, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Equal(t, id, accessKey.ID)
		assert.False(t, accessKey.RevokedAt.Valid)

		// revoke the key
		accessKey.RevokedAt.Valid = true
		accessKey.RevokedAt.Time = time.Now()
		err = accessKey.Save(ctx)
		require.NoError(t, err)

		var revokedKey *AccessKey
		revokedKey, err = GetAccessKey(ctx, id, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Equal(t, id, revokedKey.ID)
		assert.True(t, revokedKey.RevokedAt.Valid)
	})
}

// TestAccessKey_GetAccessKey will test the method GetAccessKey()
func TestAccessKey_GetAccessKey(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		accessKey, err := GetAccessKey(ctx, testXPubID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Nil(t, accessKey)
	})

	t.Run("found tx", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true)
		defer deferMe()
		opts := client.DefaultModelOptions()
		ak := newAccessKey(testXPubID, append(opts, New())...)
		txErr := ak.Save(ctx)
		require.NoError(t, txErr)
		assert.NotEqual(t, "", ak.Key)

		accessKey, err := GetAccessKey(ctx, ak.ID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.NotNil(t, accessKey)
		assert.Equal(t, ak.ID, accessKey.ID)
		assert.Equal(t, testXPubID, accessKey.XpubID)
		assert.Equal(t, "", accessKey.Key)
	})
}

// TestAccessKey_GetAccessKeys will test the method GetAccessKeys()
func TestAccessKey_GetAccessKeys(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		accessKey, err := GetAccessKeys(ctx, testXPubID, nil, nil, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Nil(t, accessKey)
	})

	t.Run("found txs", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true)
		defer deferMe()
		opts := client.DefaultModelOptions()
		ak := newAccessKey(testXPubID, append(opts, New())...)
		txErr := ak.Save(ctx)
		require.NoError(t, txErr)
		assert.NotEqual(t, "", ak.Key)

		ak2 := newAccessKey(testXPubID, append(opts, New())...)
		txErr = ak2.Save(ctx)
		require.NoError(t, txErr)
		assert.NotEqual(t, "", ak2.Key)

		accessKeys, err := GetAccessKeys(ctx, testXPubID, nil, nil, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Len(t, accessKeys, 2)
		assert.Equal(t, ak.ID, accessKeys[0].ID)
		assert.Equal(t, testXPubID, accessKeys[0].XpubID)
		assert.Equal(t, "", accessKeys[0].Key)
		assert.Equal(t, ak2.ID, accessKeys[1].ID)
		assert.Equal(t, testXPubID, accessKeys[1].XpubID)
		assert.Equal(t, "", accessKeys[1].Key)
	})

	t.Run("found txs with metadata", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true)
		defer deferMe()
		opts := client.DefaultModelOptions()
		ak := newAccessKey(testXPubID, append(opts, New(), WithMetadata("test-key", "test-value-1"))...)
		txErr := ak.Save(ctx)
		require.NoError(t, txErr)
		assert.NotEqual(t, "", ak.Key)

		ak2 := newAccessKey(testXPubID, append(opts, New(), WithMetadata("test-key", "test-value-2"))...)
		txErr = ak2.Save(ctx)
		require.NoError(t, txErr)
		assert.NotEqual(t, "", ak2.Key)

		metadata := &Metadata{"test-key": "test-value-2"}
		accessKeys, err := GetAccessKeys(ctx, testXPubID, metadata, nil, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Len(t, accessKeys, 1)
		assert.Equal(t, ak2.ID, accessKeys[0].ID)
		assert.Equal(t, testXPubID, accessKeys[0].XpubID)
		assert.Equal(t, "", accessKeys[0].Key)
	})
}
