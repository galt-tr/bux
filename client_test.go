package bux

import (
	"context"
	"testing"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
)

// todo: finish unit tests!

// TestClient_Debug will test the method Debug()
func TestClient_Debug(t *testing.T) {
	t.Parallel()

	t.Run("load basic Datastore and Cachestore with debug", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.Equal(t, false, tc.IsDebug())

		tc.Debug(true)

		assert.Equal(t, true, tc.IsDebug())
		assert.Equal(t, true, tc.Datastore().IsDebug())
		assert.Equal(t, true, tc.Cachestore().IsDebug())
		assert.Equal(t, true, tc.Taskmanager().IsDebug())
	})
}

// TestClient_IsDebug will test the method IsDebug()
func TestClient_IsDebug(t *testing.T) {
	t.Parallel()

	t.Run("basic debug checks", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.Equal(t, false, tc.IsDebug())

		tc.Debug(true)

		assert.Equal(t, true, tc.IsDebug())
	})
}

// TestClient_Version will test the method Version()
func TestClient_Version(t *testing.T) {
	t.Parallel()

	t.Run("check version", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.Equal(t, version, tc.Version())
	})
}

// TestClient_loadCache will test the method loadCache()
func TestClient_loadCache(t *testing.T) {
	// finish test
}

// TestClient_loadDatastore will test the method loadDatastore()
func TestClient_loadDatastore(t *testing.T) {
	// finish test
}

// TestClient_Cachestore will test the method Cachestore()
func TestClient_Cachestore(t *testing.T) {
	t.Parallel()

	t.Run("no options, panic", func(t *testing.T) {
		assert.Panics(t, func() {
			c := new(Client)
			assert.Nil(t, c.Cachestore())
		})
	})

	t.Run("valid cachestore", func(t *testing.T) {
		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		defer CloseClient(context.Background(), t, tc)

		assert.NotNil(t, tc.Cachestore())
		assert.IsType(t, &cachestore.Client{}, tc.Cachestore())
	})
}

// TestClient_Datastore will test the method Datastore()
func TestClient_Datastore(t *testing.T) {
	t.Parallel()

	t.Run("no options, panic", func(t *testing.T) {
		assert.Panics(t, func() {
			c := new(Client)
			assert.Nil(t, c.Datastore())
		})
	})

	t.Run("valid datastore", func(t *testing.T) {
		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		defer CloseClient(context.Background(), t, tc)

		assert.NotNil(t, tc.Datastore())
		assert.IsType(t, &datastore.Client{}, tc.Datastore())
	})
}

// TestClient_AddModels will test the method AddModels()
func TestClient_AddModels(t *testing.T) {
	// finish test
}

// TestClient_Close will test the method Close()
func TestClient_Close(t *testing.T) {
	// finish test
}

// TestClient_GetFeeUnit will test the method GetFeeUnit()
func TestClient_GetFeeUnit(t *testing.T) {
	// finish test
}

// TestClient_PaymailClient will test the method PaymailClient()
func TestClient_PaymailClient(t *testing.T) {
	t.Parallel()

	t.Run("no options, panic", func(t *testing.T) {
		assert.Panics(t, func() {
			c := new(Client)
			assert.Nil(t, c.PaymailClient())
		})
	})

	t.Run("valid paymail client", func(t *testing.T) {
		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		defer CloseClient(context.Background(), t, tc)

		assert.NotNil(t, tc.PaymailClient())
		assert.IsType(t, &paymail.Client{}, tc.PaymailClient())
	})
}

// TestClient_ModifyPaymailConfig will test the method ModifyPaymailConfig()
func TestClient_ModifyPaymailConfig(t *testing.T) {
	t.Parallel()

	t.Run("no options, panic", func(t *testing.T) {
		assert.Panics(t, func() {
			c := new(Client)
			c.ModifyPaymailConfig(nil, "", "")
		})
	})

	t.Run("update from sender and note", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithPaymailServer(&server.Configuration{
			APIVersion:      "v1",
			BSVAliasVersion: paymail.DefaultBsvAliasVersion,
			Port:            paymail.DefaultPort,
			ServiceName:     paymail.DefaultServiceName,
		}, defaultSenderPaymail, defaultAddressResolutionPurpose))

		tc, err := NewClient(context.Background(), opts...)
		require.NoError(t, err)
		assert.NotNil(t, tc.PaymailServerConfig())
		defer CloseClient(context.Background(), t, tc)

		tc.ModifyPaymailConfig(nil, "from", "note")

		assert.Equal(t, "from", tc.PaymailServerConfig().DefaultFromPaymail)
		assert.Equal(t, "note", tc.PaymailServerConfig().DefaultNote)
		assert.NotNil(t, tc.PaymailServerConfig())
	})

	t.Run("update server config", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithPaymailServer(&server.Configuration{
			APIVersion:      "v1",
			BSVAliasVersion: paymail.DefaultBsvAliasVersion,
			Port:            paymail.DefaultPort,
			ServiceName:     paymail.DefaultServiceName,
		}, defaultSenderPaymail, defaultAddressResolutionPurpose))

		tc, err := NewClient(context.Background(), opts...)
		require.NoError(t, err)
		assert.NotNil(t, tc.PaymailServerConfig())
		defer CloseClient(context.Background(), t, tc)

		assert.Equal(t, defaultSenderPaymail, tc.PaymailServerConfig().DefaultFromPaymail)
		assert.Equal(t, defaultAddressResolutionPurpose, tc.PaymailServerConfig().DefaultNote)

		tc.ModifyPaymailConfig(&server.Configuration{
			APIVersion:      "v2",
			BSVAliasVersion: paymail.DefaultBsvAliasVersion,
			Port:            paymail.DefaultPort,
			ServiceName:     paymail.DefaultServiceName,
		}, "", "")

		assert.Equal(t, defaultSenderPaymail, tc.PaymailServerConfig().DefaultFromPaymail)
		assert.Equal(t, defaultAddressResolutionPurpose, tc.PaymailServerConfig().DefaultNote)

		config := tc.PaymailServerConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "v2", config.APIVersion)
	})
}

// TestClient_PaymailServerConfig will test the method PaymailServerConfig()
func TestClient_PaymailServerConfig(t *testing.T) {
	t.Parallel()

	t.Run("no options, panic", func(t *testing.T) {
		assert.Panics(t, func() {
			c := new(Client)
			assert.Nil(t, c.PaymailServerConfig())
		})
	})

	t.Run("valid paymail server config", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithPaymailServer(&server.Configuration{
			APIVersion:      "v1",
			BSVAliasVersion: paymail.DefaultBsvAliasVersion,
			Port:            paymail.DefaultPort,
			ServiceName:     paymail.DefaultServiceName,
		}, defaultSenderPaymail, defaultAddressResolutionPurpose))

		tc, err := NewClient(context.Background(), opts...)
		require.NoError(t, err)
		defer CloseClient(context.Background(), t, tc)

		assert.NotNil(t, tc.PaymailServerConfig())
		assert.IsType(t, &paymailServerOptions{}, tc.PaymailServerConfig())
	})
}

// TestPaymailOptions_Client will test the method Client()
func TestPaymailOptions_Client(t *testing.T) {
	t.Parallel()

	t.Run("no client", func(t *testing.T) {
		p := new(paymailOptions)
		assert.Nil(t, p.Client())
	})

	t.Run("valid paymail client", func(t *testing.T) {
		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		assert.NotNil(t, tc.PaymailClient())
		defer CloseClient(context.Background(), t, tc)

		assert.IsType(t, &paymail.Client{}, tc.PaymailClient())
		assert.NotNil(t, tc.PaymailClient())
		assert.IsType(t, &paymail.Client{}, tc.PaymailClient())
	})
}

// TestPaymailOptions_FromSender will test the method FromSender()
func TestPaymailOptions_FromSender(t *testing.T) {
	t.Parallel()

	t.Run("no sender, use default", func(t *testing.T) {
		p := &paymailOptions{
			serverConfig: &paymailServerOptions{},
		}
		assert.Equal(t, defaultSenderPaymail, p.FromSender())
	})

	t.Run("custom sender set", func(t *testing.T) {
		p := &paymailOptions{
			serverConfig: &paymailServerOptions{
				DefaultFromPaymail: "from@domain.com",
			},
		}
		assert.Equal(t, "from@domain.com", p.FromSender())
	})
}

// TestPaymailOptions_Note will test the method Note()
func TestPaymailOptions_Note(t *testing.T) {
	t.Parallel()

	t.Run("no note, use default", func(t *testing.T) {
		p := &paymailOptions{
			serverConfig: &paymailServerOptions{},
		}
		assert.Equal(t, defaultAddressResolutionPurpose, p.Note())
	})

	t.Run("custom note set", func(t *testing.T) {
		p := &paymailOptions{
			serverConfig: &paymailServerOptions{
				DefaultNote: "from this person",
			},
		}
		assert.Equal(t, "from this person", p.Note())
	})
}

// TestPaymailOptions_ServerConfig will test the method ServerConfig()
func TestPaymailOptions_ServerConfig(t *testing.T) {
	t.Parallel()

	t.Run("no server config", func(t *testing.T) {
		p := new(paymailOptions)
		assert.Nil(t, p.ServerConfig())
	})

	t.Run("valid server config", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithPaymailServer(&server.Configuration{
			APIVersion:      "v1",
			BSVAliasVersion: paymail.DefaultBsvAliasVersion,
			Port:            paymail.DefaultPort,
			ServiceName:     paymail.DefaultServiceName,
		}, defaultSenderPaymail, defaultAddressResolutionPurpose))

		tc, err := NewClient(context.Background(), opts...)
		require.NoError(t, err)
		defer CloseClient(context.Background(), t, tc)

		assert.NotNil(t, tc.PaymailServerConfig())
		assert.IsType(t, &paymailServerOptions{}, tc.PaymailServerConfig())
	})
}
