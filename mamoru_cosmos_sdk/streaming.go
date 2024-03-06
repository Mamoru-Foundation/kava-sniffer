package mamoru_cosmos_sdk

import (
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"

	"github.com/Mamoru-Foundation/mamoru-sniffer-go/mamoru_sniffer/cosmos"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

var _ baseapp.StreamingService = (*StreamingService)(nil)

type StreamingService struct {
	logger             log.Logger
	blockMetadata      types.BlockMetadata
	currentBlockNumber int64
	storeListeners     []*types.MemoryListener

	sniffer *Sniffer
}

func NewStreamingService(logger log.Logger, sniffer *Sniffer) *StreamingService {
	logger.Info("Mamoru StreamingService start")

	return &StreamingService{
		sniffer: sniffer,
		logger:  logger,
	}
}

func (ss *StreamingService) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	ss.blockMetadata = types.BlockMetadata{}
	ss.blockMetadata.RequestBeginBlock = &req
	ss.blockMetadata.ResponseBeginBlock = &res
	ss.currentBlockNumber = req.Header.Height
	ss.logger.Info("Mamoru ListenBeginBlock", "height", ss.currentBlockNumber)

	return nil
}

func (ss *StreamingService) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	ss.logger.Info("Mamoru ListenDeliverTx", "height", ss.currentBlockNumber)
	ss.blockMetadata.DeliverTxs = append(ss.blockMetadata.DeliverTxs, &types.BlockMetadata_DeliverTx{
		Request:  &req,
		Response: &res,
	})

	return nil
}

func (ss *StreamingService) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	ss.blockMetadata.RequestEndBlock = &req
	ss.blockMetadata.ResponseEndBlock = &res
	ss.logger.Info("Mamoru ListenEndBlock", "height", ss.currentBlockNumber)

	return nil
}

func (ss *StreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit) error {
	if ss.sniffer == nil || !ss.sniffer.CheckRequirements() {
		return nil
	}

	ss.blockMetadata.ResponseCommit = &res
	ss.logger.Info("Mamoru ListenCommit", "height", ss.currentBlockNumber)

	var eventCount uint64 = 0
	var txCount uint64 = 0
	builder := cosmos.NewCosmosCtxBuilder()

	blockHeight := uint64(ss.blockMetadata.RequestEndBlock.Height)
	block := cosmos.Block{
		Seq:                           blockHeight,
		Height:                        ss.blockMetadata.RequestEndBlock.Height,
		Hash:                          hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Hash),
		VersionBlock:                  ss.blockMetadata.RequestBeginBlock.Header.Version.Block,
		VersionApp:                    ss.blockMetadata.RequestBeginBlock.Header.Version.App,
		ChainId:                       ss.blockMetadata.RequestBeginBlock.Header.ChainID,
		Time:                          ss.blockMetadata.RequestBeginBlock.Header.Time.Unix(),
		LastBlockIdHash:               hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.LastBlockId.Hash),
		LastBlockIdPartSetHeaderTotal: ss.blockMetadata.RequestBeginBlock.Header.LastBlockId.PartSetHeader.Total,
		LastBlockIdPartSetHeaderHash:  hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.LastBlockId.PartSetHeader.Hash),
		LastCommitHash:                hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.LastCommitHash),
		DataHash:                      hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.DataHash),
		ValidatorsHash:                hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.ValidatorsHash),
		NextValidatorsHash:            hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.NextValidatorsHash),
		ConsensusHash:                 hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.ConsensusHash),
		AppHash:                       hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.AppHash),
		LastResultsHash:               hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.LastResultsHash),
		EvidenceHash:                  hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.EvidenceHash),
		ProposerAddress:               hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Header.ProposerAddress),
		LastCommitInfoRound:           ss.blockMetadata.RequestBeginBlock.LastCommitInfo.Round,
	}

	if ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates != nil {
		block.ConsensusParamUpdatesBlockMaxBytes = ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates.Block.MaxBytes
		block.ConsensusParamUpdatesBlockMaxGas = ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates.Block.MaxGas
		block.ConsensusParamUpdatesEvidenceMaxAgeNumBlocks = ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates.Evidence.MaxAgeNumBlocks
		block.ConsensusParamUpdatesEvidenceMaxAgeDuration = ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates.Evidence.MaxAgeDuration.Milliseconds()
		block.ConsensusParamUpdatesEvidenceMaxBytes = ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates.Evidence.MaxBytes
		block.ConsensusParamUpdatesValidatorPubKeyTypes = strings.Join(ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates.Validator.PubKeyTypes[:], ",") //todo  []string to string
		block.ConsensusParamUpdatesVersionApp = ss.blockMetadata.ResponseEndBlock.ConsensusParamUpdates.Version.GetAppVersion()
	}

	builder.SetBlock(block)

	for _, beginBlock := range ss.blockMetadata.ResponseBeginBlock.Events {
		eventCount++
		builder.AppendEvents([]cosmos.Event{
			{
				Seq:       blockHeight,
				EventType: beginBlock.Type,
			},
		})
		for _, attribute := range beginBlock.Attributes {
			builder.AppendEventAttributes([]cosmos.EventAttribute{
				{
					Seq:      blockHeight,
					EventSeq: blockHeight,
					Key:      string(attribute.Key),
					Value:    string(attribute.Value),
					Index:    attribute.Index,
				},
			})
		}
	}

	for _, validatorUpdate := range ss.blockMetadata.ResponseEndBlock.ValidatorUpdates {
		builder.AppendValidatorUpdates([]cosmos.ValidatorUpdate{
			{
				Seq:    blockHeight,
				PubKey: validatorUpdate.PubKey.GetEd25519(),
				Power:  validatorUpdate.Power,
			},
		})
	}

	for _, voteInfo := range ss.blockMetadata.RequestBeginBlock.LastCommitInfo.Votes {
		builder.AppendVoteInfos([]cosmos.VoteInfo{
			{
				Seq:              blockHeight,
				BlockSeq:         blockHeight,
				ValidatorAddress: sdktypes.ValAddress(voteInfo.Validator.Address).String(),
				ValidatorPower:   voteInfo.Validator.Power,
				SignedLastBlock:  voteInfo.SignedLastBlock,
			},
		})
	}

	for _, misbehavior := range ss.blockMetadata.RequestBeginBlock.ByzantineValidators {
		builder.AppendMisbehaviors([]cosmos.Misbehavior{
			{
				Seq:              blockHeight,
				BlockSeq:         blockHeight,
				Typ:              misbehavior.Type.String(),
				ValidatorPower:   misbehavior.Validator.Power,
				ValidatorAddress: sdktypes.ValAddress(misbehavior.Validator.Address).String(),
				Height:           misbehavior.Height,
				Time:             misbehavior.Time.Unix(),
				TotalVotingPower: misbehavior.TotalVotingPower,
			},
		})
	}

	for _, tx := range ss.blockMetadata.DeliverTxs {
		txCount++
		builder.AppendTxs([]cosmos.Transaction{
			{
				Seq:       blockHeight,
				Tx:        tx.Request.Tx,
				Code:      tx.Response.Code,
				Data:      tx.Response.Data,
				Log:       tx.Response.Log,
				Info:      tx.Response.Info,
				GasWanted: tx.Response.GasWanted,
				GasUsed:   tx.Response.GasUsed,
				Codespace: tx.Response.Codespace,
			},
		})

		for _, event := range tx.Response.Events {
			eventCount++
			builder.AppendEvents([]cosmos.Event{
				{
					Seq:       blockHeight,
					EventType: event.Type,
				},
			})
			for _, attribute := range event.Attributes {
				builder.AppendEventAttributes([]cosmos.EventAttribute{
					{
						Seq:      blockHeight,
						EventSeq: blockHeight,
						Key:      string(attribute.Key),
						Value:    string(attribute.Value),
						Index:    attribute.Index,
					},
				})
			}
		}
	}

	for _, event := range ss.blockMetadata.ResponseEndBlock.Events {
		eventCount++
		builder.AppendEvents([]cosmos.Event{
			{
				Seq:       blockHeight,
				EventType: event.Type,
			},
		})
		for _, attribute := range event.Attributes {
			builder.AppendEventAttributes([]cosmos.EventAttribute{
				{
					Seq:      blockHeight,
					EventSeq: blockHeight,
					Key:      string(attribute.Key),
					Value:    string(attribute.Value),
					Index:    attribute.Index,
				},
			})
		}
	}

	builder.SetBlockData(strconv.FormatUint(blockHeight, 10), hex.EncodeToString(ss.blockMetadata.RequestBeginBlock.Hash))

	statTxs := txCount
	statEvn := eventCount
	eventCount = 0
	txCount = 0

	builder.SetStatistics(uint64(1), statTxs, statEvn, 0)

	cosmosCtx := builder.Finish()

	ss.logger.Info("Mamoru Send", "height", ss.currentBlockNumber, "txs", statTxs, "events", statEvn)

	if client := ss.sniffer.Client(); client != nil {
		client.ObserveCosmosData(cosmosCtx)
	}

	return nil
}

func (ss *StreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}

func (ss *StreamingService) Listeners() map[types.StoreKey][]types.WriteListener {
	listeners := make(map[types.StoreKey][]types.WriteListener, len(ss.storeListeners))
	//for _, listener := range ss.storeListeners {
	//	listeners[listener.StoreKey()] = []types.WriteListener{listener}
	//}
	return listeners
}

func (ss StreamingService) Close() error {
	return nil
}
