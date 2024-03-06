package sync_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/assert"
)

var mockSyncStatusResponse = `{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "node_info": {
      "protocol_version": {
        "p2p": "8",
        "block": "11",
        "app": "0"
      },
      "id": "adde11e52e960a6b204ec0fdfbfbf65049d325bf",
      "listen_addr": "tcp://0.0.0.0:26655",
      "network": "kava_2222-10",
      "version": "0.34.27",
      "channels": "40202122233038606100",
      "moniker": "mamoru-kava",
      "other": {
        "tx_index": "on",
        "rpc_address": "tcp://127.0.0.1:26658"
      }
    },
    "sync_info": {
      "latest_block_hash": "320B1CBF4D15D5E5BA89E3D36121385874251C3BD1847C2D6CE47BFCFD4F4D09",
      "latest_app_hash": "FA397677181078430BEC3E10F829111D554963A94618DDCC2B3DAD73F3FFA54D",
      "latest_block_height": "9042629",
      "latest_block_time": "2024-03-18T10:54:16.728631892Z",
      "earliest_block_hash": "72CD24385249F6BF6F1ECD92E9B9EDA6A5DD241D74A7501EC818BCE132D32E0F",
      "earliest_app_hash": "785C3EBA43E200377C68E907ADD843EC529B55A931F39C38B9035D34E1C6A1F0",
      "earliest_block_height": "9037588",
      "earliest_block_time": "2024-03-18T02:08:51.032328852Z",
      "catching_up": true
    },
    "validator_info": {
      "address": "60321514488E8840E437FE6E991D76ABFB15C5A7",
      "pub_key": {
        "type": "tendermint/PubKeyEd25519",
        "value": "BFVaUzOk0pmh4hIilFF+9A20fgtUq3o5ngyAUgcWanc="
      },
      "voting_power": "0"
    }
  }
}`

func TestNewHTTPRequest(t *testing.T) {
	logger := log.NewNopLogger()
	client := NewHTTPRequest(logger, "http://localhost", 10, true)
	assert.NotNil(t, client)
}

func TestGetCurrentBlockNumber(t *testing.T) {
	sync := &JSONRPCResponse{
		Result: Result{
			SyncInfo: SyncInfo{
				LatestBlockHeight: "100",
			},
		},
	}
	assert.Equal(t, uint64(100), sync.GetCurrentBlockNumber())
}

func TestGetCurrentBlockNumberWithInvalidHeight(t *testing.T) {
	sync := &JSONRPCResponse{
		Result: Result{
			SyncInfo: SyncInfo{
				LatestBlockHeight: "invalid",
			},
		},
	}
	assert.Equal(t, uint64(0), sync.GetCurrentBlockNumber())
}

func TestIsSync(t *testing.T) {
	sync := &JSONRPCResponse{
		Result: Result{
			SyncInfo: SyncInfo{
				CatchingUp: false,
			},
		},
	}
	assert.True(t, sync.IsSync())
}

func TestSendJSONRPCRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte(mockSyncStatusResponse))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := sendJSONRPCRequest(ctx, server.URL)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestSendJSONRPCRequestFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := sendJSONRPCRequest(ctx, "http://invalid-url")
	assert.Error(t, err)
}

func TestFetchSyncStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte(mockSyncStatusResponse))
	}))
	defer server.Close()

	logger := log.NewNopLogger()
	client := NewHTTPRequest(logger, server.URL, 10, true)
	client.fetchSyncStatus()

	assert.NotNil(t, client.GetSyncData())
}

func TestFetchSyncStatusWithInvalidURL(t *testing.T) {
	logger := log.NewNopLogger()
	client := NewHTTPRequest(logger, "http://invalid-url", 10, true)
	client.fetchSyncStatus()

	assert.Nil(t, client.GetSyncData())
}

func TestClose(t *testing.T) {
	logger := log.NewNopLogger()
	client := NewHTTPRequest(logger, "http://localhost", 10, true)
	client.Close()

	_, ok := <-client.quit
	assert.False(t, ok)
}
