package mamoru_cosmos_sdk

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"
	"os"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// TestNewSniffer tests the NewSniffer function
func TestNewSniffer(t *testing.T) {
	snifferTest := NewSniffer(log.NewTMLogger(log.NewSyncWriter(os.Stdout)))
	if snifferTest == nil {
		t.Error("NewSniffer returned nil")
	}
}

// TestIsSnifferEnable tests the isSnifferEnable method
func TestIsSnifferEnable(t *testing.T) {

	// Set environment variable for testing
	t.Setenv("MAMORU_SNIFFER_ENABLE", "true")
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	_ = NewSniffer(logger)
	if !isSnifferEnabled() {
		t.Error("Expected sniffer to be enabled")
	}

	// Test with invalid value
	t.Setenv("MAMORU_SNIFFER_ENABLE", "not_a_bool")
	if isSnifferEnabled() {
		t.Error("Expected sniffer to be disabled with invalid env value")
	}
}

// smoke test for the sniffer
func TestSnifferSmoke(t *testing.T) {
	t.Skip()
	t.Setenv("MAMORU_SNIFFER_ENABLE", "true")
	t.Setenv("MAMORU_CHAIN_TYPE", "ETH_TESTNET")
	t.Setenv("MAMORU_CHAIN_ID", "validationchain")
	t.Setenv("MAMORU_STATISTICS_SEND_INTERVAL_SECS", "1")
	t.Setenv("MAMORU_ENDPOINT", "http://localhost:9090")
	t.Setenv("MAMORU_PRIVATE_KEY", "/6Hi8mqAFp14m3pySNYDjXhUysZok0X6jaMWvwZGdd8=")
	//InitConnectFunc(func() (*cosmos.SnifferCosmos, error) {
	//	return nil, nil
	//})
	logger := log.TestingLogger()
	sniffer := NewSniffer(logger)
	if sniffer == nil {
		t.Error("NewSniffer returned nil")
	}
	header := tmprototypes.Header{}
	ischeck := true
	ctx := sdk.NewContext(nil, header, ischeck, logger)
	sdkctx := sdk.UnwrapSDKContext(ctx)

	getTStoreFunc := func(ctx sdk.Context) sdk.KVStore {
		return sdkctx.TransientStore(sdk.NewKVStoreKey("somekey"))
	}

	streamingService := NewStreamingService(logger, sniffer, getTStoreFunc)
	regBB := abci.RequestBeginBlock{}
	resBB := abci.ResponseBeginBlock{}
	err := streamingService.ListenBeginBlock(ctx, regBB, resBB)
	assert.NilError(t, err)
	regDT := abci.RequestDeliverTx{}
	resDT := abci.ResponseDeliverTx{}
	err = streamingService.ListenDeliverTx(ctx, regDT, resDT)
	assert.NilError(t, err)
	regEB := abci.RequestEndBlock{}
	resEB := abci.ResponseEndBlock{}
	err = streamingService.ListenEndBlock(ctx, regEB, resEB)
	assert.NilError(t, err)

	resC := abci.ResponseCommit{}
	err = streamingService.ListenCommit(ctx, resC)
	assert.NilError(t, err)
}
