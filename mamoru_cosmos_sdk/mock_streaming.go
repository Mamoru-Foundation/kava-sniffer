package mamoru_cosmos_sdk

import (
	"context"
	abci "github.com/cometbft/cometbft/abci/types"
	tmlog "github.com/cometbft/cometbft/libs/log"
	"sync"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	//"github.com/Mamoru-Foundation/mamoru-sniffer-go/mamoru_sniffer"
)

var _ baseapp.StreamingService = (*MockStreamingService)(nil)

// MockStreamingService mock streaming service
type MockStreamingService struct {
	logger             tmlog.Logger
	currentBlockNumber int64
	storeListeners     []*types.MemoryListener
}

func NewMockStreamingService(logger tmlog.Logger) *MockStreamingService {
	logger.Info("Mamoru MockStreamingService start")

	//_ = mamoru_sniffer.LogLevel(1)

	storeKeys := []types.StoreKey{sdk.NewKVStoreKey("mamoru")}
	listeners := make([]*types.MemoryListener, len(storeKeys))
	for i, key := range storeKeys {
		listeners[i] = types.NewMemoryListener(key)
	}
	return &MockStreamingService{
		logger:         logger,
		storeListeners: listeners,
	}
}

func (ss *MockStreamingService) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	ss.currentBlockNumber = req.Header.Height
	ss.logger.Info("Mamoru Mock ListenBeginBlock", "height", ss.currentBlockNumber)

	return nil
}

func (ss *MockStreamingService) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	ss.logger.Info("Mamoru Mock ListenDeliverTx", "height", ss.currentBlockNumber)

	return nil
}

func (ss *MockStreamingService) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	ss.logger.Info("Mamoru Mock ListenEndBlock", "height", ss.currentBlockNumber)

	return nil
}

func (ss *MockStreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit) error {
	ss.logger.Info("Mamoru Mock ListenCommit", "height", ss.currentBlockNumber)

	return nil
}

func (ss *MockStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}

func (ss *MockStreamingService) Listeners() map[types.StoreKey][]types.WriteListener {
	listeners := make(map[types.StoreKey][]types.WriteListener, len(ss.storeListeners))
	for _, listener := range ss.storeListeners {
		listeners[listener.StoreKey()] = []types.WriteListener{listener}
	}
	return listeners
}

func (ss MockStreamingService) Close() error {
	return nil
}
