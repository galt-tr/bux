package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/tonicpow/go-minercraft"
)

// chainStateBase is the base interface / methods
type chainStateBase struct {
}

func (c *chainStateBase) Broadcast(context.Context, string, string, time.Duration) error {
	return nil
}

func (c *chainStateBase) QueryTransaction(context.Context, string,
	chainstate.RequiredIn, time.Duration) (*chainstate.TransactionInfo, error) {
	return nil, nil
}

func (c *chainStateBase) QueryTransactionFastest(context.Context, string, chainstate.RequiredIn,
	time.Duration) (*chainstate.TransactionInfo, error) {
	return nil, nil
}

func (c *chainStateBase) BroadcastMiners() []*minercraft.Miner {
	return nil
}

func (c *chainStateBase) Close(context.Context) {}

func (c *chainStateBase) Debug(bool) {}

func (c *chainStateBase) DebugLog(string) {}

func (c *chainStateBase) HTTPClient() chainstate.HTTPInterface {
	return nil
}

func (c *chainStateBase) IsDebug() bool {
	return false
}

func (c *chainStateBase) IsNewRelicEnabled() bool {
	return true
}

func (c *chainStateBase) MatterCloud() mattercloud.ClientInterface {
	return nil
}

func (c *chainStateBase) Minercraft() minercraft.ClientInterface {
	return nil
}

func (c *chainStateBase) NowNodes() nownodes.ClientInterface {
	return nil
}

func (c *chainStateBase) Miners() []*minercraft.Miner {
	return nil
}

func (c *chainStateBase) Network() chainstate.Network {
	return chainstate.MainNet
}

func (c *chainStateBase) QueryMiners() []*minercraft.Miner {
	return nil
}

func (c *chainStateBase) QueryTimeout() time.Duration {
	return 10 * time.Second
}

func (c *chainStateBase) WhatsOnChain() whatsonchain.ClientInterface {
	return nil
}

type chainStateEverythingInMempool struct {
	chainStateBase
}

func (c *chainStateEverythingInMempool) Broadcast(context.Context, string, string, time.Duration) error {
	return nil
}

func (c *chainStateEverythingInMempool) QueryTransaction(_ context.Context, id string,
	_ chainstate.RequiredIn, _ time.Duration) (*chainstate.TransactionInfo, error) {

	minerID, _ := utils.RandomHex(32)
	return &chainstate.TransactionInfo{
		BlockHash:     "",
		BlockHeight:   0,
		Confirmations: 0,
		ID:            id,
		MinerID:       minerID,
		Provider:      "some-miner-name",
	}, nil
}

func (c *chainStateEverythingInMempool) QueryTransactionFastest(_ context.Context, id string, _ chainstate.RequiredIn,
	_ time.Duration) (*chainstate.TransactionInfo, error) {

	minerID, _ := utils.RandomHex(32)
	return &chainstate.TransactionInfo{
		BlockHash:     "",
		BlockHeight:   0,
		Confirmations: 0,
		ID:            id,
		MinerID:       minerID,
		Provider:      "some-miner-name",
	}, nil
}

type chainStateEverythingOnChain struct {
	chainStateEverythingInMempool
}

func (c *chainStateEverythingOnChain) QueryTransaction(_ context.Context, id string,
	_ chainstate.RequiredIn, _ time.Duration) (*chainstate.TransactionInfo, error) {

	hash, _ := utils.RandomHex(32)
	return &chainstate.TransactionInfo{
		BlockHash:     hash,
		BlockHeight:   600000,
		Confirmations: 10,
		ID:            id,
		MinerID:       "",
		Provider:      "whatsonchain",
	}, nil
}

func (c *chainStateEverythingOnChain) QueryTransactionFastest(_ context.Context, id string, _ chainstate.RequiredIn,
	_ time.Duration) (*chainstate.TransactionInfo, error) {

	hash, _ := utils.RandomHex(32)
	return &chainstate.TransactionInfo{
		BlockHash:     hash,
		BlockHeight:   600000,
		Confirmations: 10,
		ID:            id,
		MinerID:       "",
		Provider:      "whatsonchain",
	}, nil
}
