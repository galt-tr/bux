package chainstate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_Transaction will test the method QueryTransaction()
func TestClient_Transaction(t *testing.T) {
	t.Parallel()

	t.Run("error - missing id", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)

		info, err := c.QueryTransaction(
			context.Background(), "", RequiredOnChain, defaultQueryTimeOut,
		)
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidTransactionID)
	})

	t.Run("error - missing requirements", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			"", defaultQueryTimeOut,
		)
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidRequirements)
	})

	t.Run("valid - all four", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNowNodes(&nowNodesTxOnChain{}),
		)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, minerTaal.Name, info.Provider)
		assert.Equal(t, "030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e", info.MinerID)
	})

	t.Run("mAPI not found - woc, mattercloud, nownodes", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}), // NOT going to find the TX
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNowNodes(&nowNodesTxOnChain{}),
		)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, providerWhatsOnChain, info.Provider)
	})

	t.Run("mAPI, WOC not found - mattercloud, nownodes", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}),     // NOT going to find the TX
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // NOT going to find the TX
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNowNodes(&nowNodesTxOnChain{}),
		)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, providerMatterCloud, info.Provider)
	})

	t.Run("mAPI, WOC, mattercloud not found - nownodes", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}),     // NOT going to find the TX
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // NOT going to find the TX
			WithMatterCloud(&matterCloudTxNotFound{}),   // NOT going to find the TX
			WithNowNodes(&nowNodesTxOnChain{}),
		)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, providerNowNodes, info.Provider)
	})

	t.Run("error - all not found", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}),     // NOT going to find the TX
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // NOT going to find the TX
			WithMatterCloud(&matterCloudTxNotFound{}),   // NOT going to find the TX
			WithNowNodes(&nowNodesTxNotFound{}),         // NOT going to find the TX
		)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrTransactionNotFound)
	})

	t.Run("valid - stn network", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNetwork(StressTestNet),
		)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Contains(t, []string{providerWhatsOnChain, providerMatterCloud}, info.Provider)
	})

	t.Run("valid - test network", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNetwork(TestNet),
		)

		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
	})
}

// TestClient_TransactionFastest will test the method QueryTransactionFastest()
func TestClient_TransactionFastest(t *testing.T) {
	t.Parallel()

	t.Run("error - missing id", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)

		info, err := c.QueryTransactionFastest(
			context.Background(), "", RequiredOnChain, defaultQueryTimeOut,
		)
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidTransactionID)
	})

	t.Run("error - missing requirements", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)

		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			"", defaultQueryTimeOut,
		)
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidRequirements)
	})

	t.Run("valid - all four", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNowNodes(&nowNodesTxOnChain{}),
		)

		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
	})

	t.Run("mAPI not found - woc, mattercloud, nownodes", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}), // NOT going to find the TX
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNowNodes(&nowNodesTxOnChain{}),
		)

		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
	})

	t.Run("error - all not found", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}),     // NOT going to find the TX
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // NOT going to find the TX
			WithMatterCloud(&matterCloudTxNotFound{}),   // NOT going to find the TX
			WithNowNodes(&nowNodesTxNotFound{}),         // NOT going to find the TX
		)

		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrTransactionNotFound)
	})

	t.Run("valid - stn network", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithNetwork(StressTestNet),
		)

		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Contains(t, []string{providerWhatsOnChain, providerMatterCloud}, info.Provider)
	})
}
