package bux

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/BuxOrg/bux/utils"
	magic "github.com/bitcoinschema/go-map"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	emptyConfigJSON = "{\"change_destinations\":[{\"created_at\":\"0001-01-01T00:00:00Z\",\"updated_at\":\"0001-01-01T00:00:00Z\",\"deleted_at\":null,\"id\":\"c775e7b757ede630cd0aa1113bd102661ab38829ca52a6422ab782862f268646\",\"xpub_id\":\"1a0b10d4eda0636aae1709e7e7080485a4d99af3ca2962c6e677cf5b53d8ab8c\",\"locking_script\":\"76a9147ff514e6ae3deb46e6644caac5cdd0bf2388906588ac\",\"type\":\"pubkeyhash\",\"chain\":1,\"num\":123,\"address\":\"1CfaQw9udYNPccssFJFZ94DN8MqNZm9nGt\",\"draft_id\":\"test-reference\"}],\"change_destinations_strategy\":\"\",\"change_minimum_satoshis\":0,\"change_number_of_destinations\":0,\"change_satoshis\":124,\"expires_in\":30000000000,\"fee\":443,\"fee_unit\":{\"satoshis\":500,\"bytes\":1000},\"from_utxos\":null,\"inputs\":null,\"miner\":\"\",\"outputs\":null,\"send_all_to\":\"\",\"sync\":null}"
	opReturn        = "006a2231394878696756345179427633744870515663554551797131707a5a56646f417574324b65657020616e20657965206f6e207468697320706c61636520666f7220736f6d65204a616d696679206c6f76652e2e2e200d746578742f6d61726b646f776e055554462d38"
	unsetConfigJSON = "{\"change_destinations\":null,\"change_destinations_strategy\":\"\",\"change_minimum_satoshis\":0,\"change_number_of_destinations\":0,\"change_satoshis\":0,\"expires_in\":0,\"fee\":0,\"fee_unit\":null,\"from_utxos\":null,\"inputs\":null,\"miner\":\"\",\"outputs\":null,\"send_all_to\":\"\",\"sync\":null}"

	opReturnParts = []string{
		"31394878696756345179427633744870515663554551797131707a5a56646f417574",
		"4b65657020616e20657965206f6e207468697320706c61636520666f7220736f6d65204a616d696679206c6f76652e2e2e20",
		"746578742f6d61726b646f776e",
		"5554462d38",
	}

	stringParts = []string{
		"19HxigV4QyBv3tHpQVcUEQyq1pzZVdoAut",
		"Keep an eye on this place for some Jamify love... ",
		"text/markdown",
		"UTF-8",
	}

	emptyConfig = TransactionConfig{
		ChangeDestinations: []*Destination{{
			Address:       testExternalAddress,
			Chain:         1,
			ID:            testDestinationID,
			LockingScript: testLockingScript,
			Num:           123,
			DraftID:       "test-reference",
			Type:          utils.ScriptTypePubKeyHash,
			XpubID:        testXPubID,
		}},
		ChangeSatoshis: 124,
		ExpiresIn:      defaultDraftTxExpiresIn,
		FeeUnit: &utils.FeeUnit{
			Satoshis: 500,
			Bytes:    1000,
		},
		Fee:     443,
		Inputs:  nil,
		Outputs: nil,
	}
)

// assertEmptyTransactionConfig will test the config
func assertEmptyTransactionConfig(t *testing.T, transactionConfig TransactionConfig) {
	assert.Nil(t, transactionConfig.ChangeDestinations)
	assert.Nil(t, transactionConfig.FeeUnit)
	assert.Nil(t, transactionConfig.Inputs)
	assert.Nil(t, transactionConfig.Outputs)
	assert.Equal(t, uint64(0), transactionConfig.ChangeSatoshis)
	assert.Equal(t, uint64(0), transactionConfig.Fee)
}

// TestTransactionConfigScan will test the db Scanner of the TransactionConfig model
func TestTransactionConfig_Scan(t *testing.T) {
	t.Parallel()

	t.Run("nil value", func(t *testing.T) {
		transactionConfig := TransactionConfig{}
		err := transactionConfig.Scan(nil)
		require.NoError(t, err)
		assertEmptyTransactionConfig(t, transactionConfig)
	})

	t.Run("empty string", func(t *testing.T) {
		transactionConfig := TransactionConfig{}
		err := transactionConfig.Scan([]byte("\"\""))
		assert.NoError(t, err)
		assertEmptyTransactionConfig(t, transactionConfig)
	})

	t.Run("empty string - incorrectly coded", func(t *testing.T) {
		transactionConfig := TransactionConfig{}
		err := transactionConfig.Scan([]byte(""))
		assert.NoError(t, err)
		assertEmptyTransactionConfig(t, transactionConfig)
	})

	t.Run("object", func(t *testing.T) {
		transactionConfig := TransactionConfig{}
		err := transactionConfig.Scan([]byte(emptyConfigJSON))
		require.NoError(t, err)
		assert.Equal(t, emptyConfig, transactionConfig)
	})
}

// TestTransactionConfigValue will test the db Valuer of the TransactionConfig model
func TestTransactionConfig_Value(t *testing.T) {
	t.Parallel()

	t.Run("empty object", func(t *testing.T) {
		transactionConfig := TransactionConfig{}
		value, err := transactionConfig.Value()
		require.NoError(t, err)
		assert.Equal(t, unsetConfigJSON, value)
	})

	t.Run("full config", func(t *testing.T) {
		transactionConfig := emptyConfig
		value, err := transactionConfig.Value()
		require.NoError(t, err)
		assert.Equal(t, emptyConfigJSON, value)
	})
}

// TestTransactionConfig_processAddressOutput will test the method processAddressOutput()
func TestTransactionConfig_processAddressOutput(t *testing.T) {
	// t.Parallel() mocking does not allow parallel tests

	sats := uint64(1000)
	address := "1CfaQw9udYNPccssFJFZ94DN8MqNZm9nGt"

	t.Run("valid address output", func(t *testing.T) {
		out := &TransactionOutput{
			Satoshis: sats,
			To:       address,
		}
		require.NotNil(t, out)

		err := out.processAddressOutput()
		require.NoError(t, err)

		assert.Equal(t, 1, len(out.Scripts))
		assert.Equal(t, address, out.Scripts[0].Address)
		assert.Equal(t, sats, out.Scripts[0].Satoshis)
		assert.Equal(t, "76a9147ff514e6ae3deb46e6644caac5cdd0bf2388906588ac", out.Scripts[0].Script)
	})

	t.Run("invalid address", func(t *testing.T) {
		out := &TransactionOutput{
			Satoshis: sats,
			To:       "123456",
		}
		require.NotNil(t, out)

		err := out.processAddressOutput()
		require.Error(t, err)
	})
}

// TestTransactionConfig_processOutput will test the method processOutput()
func TestTransactionConfig_processOutput(t *testing.T) {
	// t.Parallel() mocking does not allow parallel tests

	sats := uint64(1000)
	paymailAddress := "TeSTeR@" + testDomain

	t.Run("error - no address or paymail given", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		out := &TransactionOutput{
			Satoshis: sats,
			To:       "",
		}

		err := out.processOutput(
			context.Background(), nil, client,
			defaultSenderPaymail, defaultAddressResolutionPurpose,
			true,
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrOutputValueNotRecognized)
	})

	t.Run("error - invalid paymail given", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		out := &TransactionOutput{
			Satoshis: sats,
			To:       testAlias + "@",
		}

		err := out.processOutput(
			context.Background(), nil, client,
			defaultSenderPaymail, defaultAddressResolutionPurpose,
			true,
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPaymailAddressIsInvalid)
	})

	t.Run("basic paymail address resolution - valid response", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		out := &TransactionOutput{
			Satoshis: sats,
			To:       paymailAddress,
		}
		require.NotNil(t, out)

		mockValidResponse(http.StatusOK, false, testDomain)

		err = out.processOutput(
			context.Background(), tc.Cachestore(), client,
			defaultSenderPaymail, defaultAddressResolutionPurpose,
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, sats, out.Satoshis)
		assert.Equal(t, testAlias+"@"+testDomain, out.To)
		assert.Equal(t, defaultSenderPaymail, out.PaymailP4.FromPaymail)
		assert.Equal(t, testAlias, out.PaymailP4.Alias)
		assert.Equal(t, testDomain, out.PaymailP4.Domain)
		assert.Equal(t, defaultAddressResolutionPurpose, out.PaymailP4.Note)
		assert.Equal(t, ResolutionTypeBasic, out.PaymailP4.ResolutionType)
		assert.Equal(t, "", out.PaymailP4.ReferenceID)
		assert.Equal(t, "", out.PaymailP4.ReceiveEndpoint)
	})

	t.Run("basic $handle -> paymail address resolution - valid response", func(t *testing.T) {
		handle := "$TeSTeR"
		handleDomain := "handcash.io"

		client := newTestPaymailClient(t, []string{testDomain, handleDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		out := &TransactionOutput{
			Satoshis: sats,
			To:       handle,
		}
		require.NotNil(t, out)

		mockValidResponse(http.StatusOK, false, handleDomain)

		err = out.processOutput(
			context.Background(), tc.Cachestore(), client,
			defaultSenderPaymail, defaultAddressResolutionPurpose,
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, sats, out.Satoshis)
		assert.Equal(t, testAlias+"@"+handleDomain, out.To)
		assert.Equal(t, defaultSenderPaymail, out.PaymailP4.FromPaymail)
		assert.Equal(t, testAlias, out.PaymailP4.Alias)
		assert.Equal(t, handleDomain, out.PaymailP4.Domain)
		assert.Equal(t, defaultAddressResolutionPurpose, out.PaymailP4.Note)
		assert.Equal(t, ResolutionTypeBasic, out.PaymailP4.ResolutionType)
		assert.Equal(t, "", out.PaymailP4.ReferenceID)
		assert.Equal(t, "", out.PaymailP4.ReceiveEndpoint)
	})

	t.Run("basic 1handle -> paymail address resolution - valid response", func(t *testing.T) {
		handle := "1TeSTeR"
		handleDomain := "relayx.io"

		client := newTestPaymailClient(t, []string{testDomain, handleDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		out := &TransactionOutput{
			Satoshis: sats,
			To:       handle,
		}
		require.NotNil(t, out)

		mockValidResponse(http.StatusOK, false, handleDomain)

		err = out.processOutput(
			context.Background(), tc.Cachestore(), client,
			defaultSenderPaymail, defaultAddressResolutionPurpose,
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, sats, out.Satoshis)
		assert.Equal(t, testAlias+"@"+handleDomain, out.To)
		assert.Equal(t, defaultSenderPaymail, out.PaymailP4.FromPaymail)
		assert.Equal(t, testAlias, out.PaymailP4.Alias)
		assert.Equal(t, handleDomain, out.PaymailP4.Domain)
		assert.Equal(t, defaultAddressResolutionPurpose, out.PaymailP4.Note)
		assert.Equal(t, ResolutionTypeBasic, out.PaymailP4.ResolutionType)
		assert.Equal(t, "", out.PaymailP4.ReferenceID)
		assert.Equal(t, "", out.PaymailP4.ReceiveEndpoint)
	})

	t.Run("p2p paymail address resolution - valid response", func(t *testing.T) {
		client := newTestPaymailClient(t, []string{testDomain})

		tc, err := NewClient(
			context.Background(),
			DefaultClientOpts(t, true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		out := &TransactionOutput{
			Satoshis: sats,
			To:       paymailAddress,
		}
		require.NotNil(t, out)

		mockValidResponse(http.StatusOK, true, testDomain)

		err = out.processOutput(
			context.Background(), tc.Cachestore(), client,
			defaultSenderPaymail, defaultAddressResolutionPurpose,
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, sats, out.Satoshis)
		assert.Equal(t, testAlias+"@"+testDomain, out.To)
		assert.Equal(t, "", out.PaymailP4.FromPaymail)
		assert.Equal(t, testAlias, out.PaymailP4.Alias)
		assert.Equal(t, testDomain, out.PaymailP4.Domain)
		assert.Equal(t, "", out.PaymailP4.Note)
		assert.Equal(t, ResolutionTypeP2P, out.PaymailP4.ResolutionType)
		assert.Equal(t, "z0bac4ec-6f15-42de-9ef4-e60bfdabf4f7", out.PaymailP4.ReferenceID)
		assert.Equal(t, testServerURL+"/receive-transaction/{alias}@{domain.tld}", out.PaymailP4.ReceiveEndpoint)
	})
}

// TestTransactionConfig_processOpReturnOutput will test the method processOpReturnOutput()
func TestTransactionConfig_processOpReturnOutput(t *testing.T) {
	t.Run("empty op_return", func(t *testing.T) {
		output := &TransactionOutput{
			OpReturn: &OpReturn{},
		}
		err := output.processOpReturnOutput()
		require.ErrorIs(t, err, ErrInvalidOpReturnOutput)
	})

	t.Run("op_return hex", func(t *testing.T) {
		output := &TransactionOutput{
			OpReturn: &OpReturn{
				Hex: opReturn,
			},
		}
		err := output.processOpReturnOutput()
		require.NoError(t, err)
		assert.Equal(t, 1, len(output.Scripts))
		assert.Equal(t, opReturn, output.Scripts[0].Script)
		assert.Equal(t, "", output.Scripts[0].Address)
		assert.Equal(t, uint64(0), output.Scripts[0].Satoshis)
	})

	t.Run("op_return hexParts", func(t *testing.T) {
		output := &TransactionOutput{
			OpReturn: &OpReturn{
				HexParts: opReturnParts,
			},
		}
		err := output.processOpReturnOutput()
		require.NoError(t, err)
		assert.Equal(t, 1, len(output.Scripts))
		assert.Equal(t, opReturn, output.Scripts[0].Script)
		assert.Equal(t, "", output.Scripts[0].Address)
		assert.Equal(t, uint64(0), output.Scripts[0].Satoshis)
	})

	t.Run("op_return stringParts", func(t *testing.T) {
		output := &TransactionOutput{
			OpReturn: &OpReturn{
				StringParts: stringParts,
			},
		}
		err := output.processOpReturnOutput()
		require.NoError(t, err)
		assert.Equal(t, 1, len(output.Scripts))
		assert.Equal(t, opReturn, output.Scripts[0].Script)
		assert.Equal(t, "", output.Scripts[0].Address)
		assert.Equal(t, uint64(0), output.Scripts[0].Satoshis)
	})

	t.Run("op_return stringParts map", func(t *testing.T) {
		mapAppName := "tonicpow"
		output := &TransactionOutput{
			OpReturn: &OpReturn{
				StringParts: []string{
					magic.Prefix,
					magic.Set,
					magic.MapAppKey,
					mapAppName,
					magic.MapTypeKey,
					"offer_click",
					"offer_config_id",
					fmt.Sprintf("%d", 23),
					"offer_session_id",
					"f54fa5c0431b37727991dab02ca0a96c0f9e2e546fd79a6e40677593f2ec8dd9",
				},
			},
		}
		err := output.processOpReturnOutput()
		require.NoError(t, err)
		assert.Equal(t, 1, len(output.Scripts))
		// https://whatsonchain.com/tx/a7a1e4cf4f7e891103bebc07f6e8ae125a67aaf16775d92a07b776d8a9a55b5d
		expected := "006a223150755161374b36324d694b43747373534c4b79316b683536575755374d74555235035345540361707008746f6e6963706f7704747970650b6f666665725f636c69636b0f6f666665725f636f6e6669675f6964023233106f666665725f73657373696f6e5f69644066353466613563303433316233373732373939316461623032636130613936633066396532653534366664373961366534303637373539336632656338646439"
		assert.Equal(t, expected, output.Scripts[0].Script)
		assert.Equal(t, "", output.Scripts[0].Address)
		assert.Equal(t, uint64(0), output.Scripts[0].Satoshis)
	})

	t.Run("op_return map", func(t *testing.T) {
		mapAppName := "tonicpow"
		output := &TransactionOutput{
			OpReturn: &OpReturn{
				Map: &MapProtocol{
					App:  mapAppName,
					Type: "offer_click",
					Keys: map[string]interface{}{
						"offer_config_id":  fmt.Sprintf("%d", 23),
						"offer_session_id": "f54fa5c0431b37727991dab02ca0a96c0f9e2e546fd79a6e40677593f2ec8dd9",
					},
				},
			},
		}
		err := output.processOpReturnOutput()
		require.NoError(t, err)
		assert.Equal(t, 1, len(output.Scripts))
		// https://whatsonchain.com/tx/a7a1e4cf4f7e891103bebc07f6e8ae125a67aaf16775d92a07b776d8a9a55b5d
		expected := "006a223150755161374b36324d694b43747373534c4b79316b683536575755374d74555235035345540361707008746f6e6963706f7704747970650b6f666665725f636c69636b0f6f666665725f636f6e6669675f6964023233106f666665725f73657373696f6e5f69644066353466613563303433316233373732373939316461623032636130613936633066396532653534366664373961366534303637373539336632656338646439"
		expected2 := "006a223150755161374b36324d694b43747373534c4b79316b683536575755374d74555235035345540361707008746f6e6963706f7704747970650b6f666665725f636c69636b106f666665725f73657373696f6e5f696440663534666135633034333162333737323739393164616230326361306139366330663965326535343666643739613665343036373735393366326563386464390f6f666665725f636f6e6669675f6964023233"
		// the order of a map is not guaranteed, but both MAP outputs are actually valid
		if output.Scripts[0].Script != expected && output.Scripts[0].Script != expected2 {
			assert.Nil(t, output.Scripts[0].Script)
		}
		assert.Equal(t, "", output.Scripts[0].Address)
		assert.Equal(t, uint64(0), output.Scripts[0].Satoshis)
	})
}
