package chainstate

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonicpow/go-minercraft"
)

// TestWithNewRelic will test the method WithNewRelic()
func TestWithNewRelic(t *testing.T) {

	t.Run("get opts", func(t *testing.T) {
		opt := WithNewRelic()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply opts", func(t *testing.T) {
		opts := []ClientOps{WithNewRelic()}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, true, c.IsNewRelicEnabled())
	})
}

// TestWithDebugging will test the method WithDebugging()
func TestWithDebugging(t *testing.T) {

	t.Run("get opts", func(t *testing.T) {
		opt := WithDebugging()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply opts", func(t *testing.T) {
		opts := []ClientOps{WithDebugging()}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, true, c.IsDebug())
	})
}

// TestWithHTTPClient will test the method WithHTTPClient()
func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithHTTPClient(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithHTTPClient(nil)
		opt(options)
		assert.Nil(t, options.config.httpClient)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		customClient := &http.Client{}
		opt := WithHTTPClient(customClient)
		opt(options)
		assert.Equal(t, customClient, options.config.httpClient)
	})
}

// TestWithMinercraft will test the method WithMinercraft()
func TestWithMinercraft(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithMinercraft(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithMinercraft(nil)
		opt(options)
		assert.Nil(t, options.config.minercraft)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		customClient := &minerCraftTxOnChain{}
		opt := WithMinercraft(customClient)
		opt(options)
		assert.Equal(t, customClient, options.config.minercraft)
	})
}

// TestWithWhatsOnChain will test the method WithWhatsOnChain()
func TestWithWhatsOnChain(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithWhatsOnChain(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithWhatsOnChain(nil)
		opt(options)
		assert.Nil(t, options.config.whatsOnChain)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		customClient := &whatsOnChainTxOnChain{}
		opt := WithWhatsOnChain(customClient)
		opt(options)
		assert.Equal(t, customClient, options.config.whatsOnChain)
	})
}

// TestWithMatterCloud will test the method WithMatterCloud()
func TestWithMatterCloud(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithMatterCloud(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithMatterCloud(nil)
		opt(options)
		assert.Nil(t, options.config.matterCloud)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		customClient := &matterCloudTxOnChain{}
		opt := WithMatterCloud(customClient)
		opt(options)
		assert.Equal(t, customClient, options.config.matterCloud)
	})
}

// TestWithMatterCloudAPIKey will test the method WithMatterCloudAPIKey()
func TestWithMatterCloudAPIKey(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithMatterCloudAPIKey("")
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying empty string", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithMatterCloudAPIKey("")
		opt(options)
		assert.Equal(t, "", options.config.matterCloudAPIKey)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithMatterCloudAPIKey(matterCloudSig1)
		opt(options)
		assert.Equal(t, matterCloudSig1, options.config.matterCloudAPIKey)
	})
}

// TestWithNowNodes will test the method WithNowNodes()
func TestWithNowNodes(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithNowNodes(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithNowNodes(nil)
		opt(options)
		assert.Nil(t, options.config.nowNodes)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		customClient := &nowNodesTxOnChain{}
		opt := WithNowNodes(customClient)
		opt(options)
		assert.Equal(t, customClient, options.config.nowNodes)
	})
}

// TestWithNowNodesAPIKey will test the method WithNowNodesAPIKey()
func TestWithNowNodesAPIKey(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithNowNodesAPIKey("")
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying empty string", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithNowNodesAPIKey("")
		opt(options)
		assert.Equal(t, "", options.config.nowNodesAPIKey)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithNowNodesAPIKey(onChainExample1TxID)
		opt(options)
		assert.Equal(t, onChainExample1TxID, options.config.nowNodesAPIKey)
	})
}

// TestWithBroadcastMiners will test the method WithBroadcastMiners()
func TestWithBroadcastMiners(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithBroadcastMiners(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{mAPI: &mAPIConfig{}},
		}
		opt := WithBroadcastMiners(nil)
		opt(options)
		assert.Nil(t, options.config.mAPI.broadcastMiners)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{mAPI: &mAPIConfig{}},
		}
		miners := []*minercraft.Miner{minerTaal}
		opt := WithBroadcastMiners([]*minercraft.Miner{minerTaal})
		opt(options)
		assert.Equal(t, miners, options.config.mAPI.broadcastMiners)
	})
}

// TestWithQueryMiners will test the method WithQueryMiners()
func TestWithQueryMiners(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithQueryMiners(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{mAPI: &mAPIConfig{}},
		}
		opt := WithQueryMiners(nil)
		opt(options)
		assert.Nil(t, options.config.mAPI.queryMiners)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{mAPI: &mAPIConfig{}},
		}
		miners := []*minercraft.Miner{minerTaal}
		opt := WithQueryMiners([]*minercraft.Miner{minerTaal})
		opt(options)
		assert.Equal(t, miners, options.config.mAPI.queryMiners)
	})
}

// TestWithCustomMiners will test the method WithCustomMiners()
func TestWithCustomMiners(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithCustomMiners(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{mAPI: &mAPIConfig{}},
		}
		opt := WithCustomMiners(nil)
		opt(options)
		assert.Nil(t, options.config.mAPI.miners)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{mAPI: &mAPIConfig{}},
		}
		miners := []*minercraft.Miner{minerTaal}
		opt := WithCustomMiners([]*minercraft.Miner{minerTaal})
		opt(options)
		assert.Equal(t, miners, options.config.mAPI.miners)
	})
}

// TestWithQueryTimeout will test the method WithQueryTimeout()
func TestWithQueryTimeout(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithQueryTimeout(0)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying empty value", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithQueryTimeout(0)
		opt(options)
		assert.Equal(t, time.Duration(0), options.config.queryTimeout)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithQueryTimeout(10 * time.Second)
		opt(options)
		assert.Equal(t, 10*time.Second, options.config.queryTimeout)
	})
}

// TestWithNetwork will test the method WithNetwork()
func TestWithNetwork(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithNetwork("")
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying empty string", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithNetwork("")
		opt(options)
		assert.Equal(t, Network(""), options.config.network)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithNetwork(TestNet)
		opt(options)
		assert.Equal(t, TestNet, options.config.network)
	})
}

// TestWithLogger will test the method WithLogger()
func TestWithLogger(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithLogger(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		opt := WithLogger(nil)
		opt(options)
		assert.Nil(t, options.logger)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{
			config: &syncConfig{},
		}
		customClient := newLogger()
		opt := WithLogger(customClient)
		opt(options)
		assert.Equal(t, customClient, options.logger)
	})
}
