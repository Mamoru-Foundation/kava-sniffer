package mamoru_cosmos_sdk

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
	"time"

	tmabci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"gotest.tools/v3/assert"
)

func TestListenBeginBlock(t *testing.T) {
	t.Run("TestListenBeginBlock", func(t *testing.T) {

		logger := log.TestingLogger()
		header := tmprototypes.Header{}
		ischeck := true
		ctx := sdk.NewContext(nil, header, ischeck, logger)
		sdkctx := sdk.UnwrapSDKContext(ctx)

		getTStoreFunc := func(ctx sdk.Context) sdk.KVStore {
			return sdkctx.TransientStore(sdk.NewKVStoreKey("somekey"))
		}

		ss := NewStreamingService(logger.With("module", "mamoru"), nil, getTStoreFunc)

		req := tmabci.RequestBeginBlock{Header: tmprototypes.Header{
			Version:            tmversion.Consensus{},
			ChainID:            "",
			Height:             1234,
			Time:               time.Time{},
			LastBlockId:        tmprototypes.BlockID{},
			LastCommitHash:     []byte{'a', 'b', 'c'},
			DataHash:           []byte{'a', 'b', 'c'},
			ValidatorsHash:     []byte{'a', 'b', 'c'},
			NextValidatorsHash: []byte{'a', 'b', 'c'},
			ConsensusHash:      []byte{'a', 'b', 'c'},
			AppHash:            []byte{'a', 'b', 'c'},
			LastResultsHash:    []byte{'a', 'b', 'c'}, EvidenceHash: []byte{'a', 'b', 'c'},
			ProposerAddress: []byte{'a', 'b', 'c'},
		}}
		res := tmabci.ResponseBeginBlock{}

		err := ss.ListenBeginBlock(ctx, req, res)
		assert.NilError(t, err)
		assert.Equal(t, ss.blockMetadata.RequestBeginBlock.Header.Height, req.Header.Height)
		assert.Equal(t, ss.currentBlockNumber, req.Header.Height)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.LastCommitHash, req.Header.LastCommitHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.DataHash, req.Header.DataHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.ValidatorsHash, req.Header.ValidatorsHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.NextValidatorsHash, req.Header.NextValidatorsHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.ConsensusHash, req.Header.ConsensusHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.AppHash, req.Header.AppHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.LastResultsHash, req.Header.LastResultsHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.EvidenceHash, req.Header.EvidenceHash)
		assert.DeepEqual(t, ss.blockMetadata.RequestBeginBlock.Header.ProposerAddress, req.Header.ProposerAddress)
	})
}
