package mamoru_cosmos_sdk

import (
	"context"

	tmabci "github.com/tendermint/tendermint/abci/types"
	"sync"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
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

func (ss *MockStreamingService) ListenBeginBlock(ctx context.Context, req tmabci.RequestBeginBlock, res tmabci.ResponseBeginBlock) error {
	ss.currentBlockNumber = req.Header.Height
	ss.logger.Info("Mamoru Mock ListenBeginBlock", "height", ss.currentBlockNumber)

	return nil
}

func (ss *MockStreamingService) ListenDeliverTx(ctx context.Context, req tmabci.RequestDeliverTx, res tmabci.ResponseDeliverTx) error {
	ss.logger.Info("Mamoru Mock ListenDeliverTx", "height", ss.currentBlockNumber)

	return nil
}

func (ss *MockStreamingService) ListenEndBlock(ctx context.Context, req tmabci.RequestEndBlock, res tmabci.ResponseEndBlock) error {
	ss.logger.Info("Mamoru Mock ListenEndBlock", "height", ss.currentBlockNumber)

	return nil
}

func (ss *MockStreamingService) ListenCommit(ctx context.Context, res tmabci.ResponseCommit) error {
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
