package sync_state

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/tendermint/tendermint/libs/log"
)

type JSONRPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  Result `json:"result"`
}

type Result struct {
	NodeInfo      NodeInfo      `json:"node_info"`
	SyncInfo      SyncInfo      `json:"sync_info"`
	ValidatorInfo ValidatorInfo `json:"validator_info"`
}

type NodeInfo struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	ID              string          `json:"id"`
	ListenAddr      string          `json:"listen_addr"`
	Network         string          `json:"network"`
	Version         string          `json:"version"`
	Channels        string          `json:"channels"`
	Moniker         string          `json:"moniker"`
	Other           Other           `json:"other"`
}

type ProtocolVersion struct {
	P2P   string `json:"p2p"`
	Block string `json:"block"`
	App   string `json:"app"`
}

type Other struct {
	TxIndex    string `json:"tx_index"`
	RPCAddress string `json:"rpc_address"`
}

type SyncInfo struct {
	LatestBlockHash     string    `json:"latest_block_hash"`
	LatestAppHash       string    `json:"latest_app_hash"`
	LatestBlockHeight   string    `json:"latest_block_height"`
	LatestBlockTime     time.Time `json:"latest_block_time"`
	EarliestBlockHash   string    `json:"earliest_block_hash"`
	EarliestAppHash     string    `json:"earliest_app_hash"`
	EarliestBlockHeight string    `json:"earliest_block_height"`
	EarliestBlockTime   time.Time `json:"earliest_block_time"`
	CatchingUp          bool      `json:"catching_up"`
}

type ValidatorInfo struct {
	Address     string `json:"address"`
	PubKey      PubKey `json:"pub_key"`
	VotingPower string `json:"voting_power"`
}

type PubKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (sync *JSONRPCResponse) GetCurrentBlockNumber() uint64 {
	if sync == nil {
		return 0
	}
	// convert string to uint64
	blockHeight, err := strconv.ParseUint(sync.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		return 0
	}
	return blockHeight
}

func (sync *JSONRPCResponse) IsSync() bool {
	if sync == nil {
		return false
	}

	return !sync.Result.SyncInfo.CatchingUp
}

type Client struct {
	logger        log.Logger
	syncData      *JSONRPCResponse
	Url           string
	PolishTimeSec uint

	quit    chan struct{}
	signals chan os.Signal
}

func NewHTTPRequest(logget log.Logger, url string, PolishTimeSec uint, enable bool) *Client {
	c := &Client{
		logger:        logget,
		Url:           url,
		PolishTimeSec: PolishTimeSec,
		quit:          make(chan struct{}),
		signals:       make(chan os.Signal, 1),
	}

	// Register for SIGINT (Ctrl+C) and SIGTERM (kill) signals
	signal.Notify(c.signals, syscall.SIGINT, syscall.SIGTERM)
	if enable {
		go c.loop()
	}

	return c
}

func (c *Client) GetSyncData() *JSONRPCResponse {
	return c.syncData
}

func (c *Client) loop() {
	// wait for 2 minutes for the node to start
	time.Sleep(2 * time.Minute)
	ticker := time.NewTicker(time.Duration(c.PolishTimeSec) * time.Second)
	defer ticker.Stop()
	// Perform the first tick immediately
	c.fetchSyncStatus()

	for {
		select {
		case <-ticker.C:
			c.fetchSyncStatus()
		case <-c.quit:
			c.logger.Info("Mamoru SyncProcess Shutting down...")
			return
		case <-c.signals:
			c.logger.Info("Signal received, initiating shutdown...")
			c.Close()
		}
	}
}

func (c *Client) fetchSyncStatus() {

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.logger.Info("Mamoru requesting syncData ...")
	// Send the request and get the response
	response, err := sendJSONRPCRequest(ctx, c.Url)
	if err != nil {
		c.logger.Error("Mamoru Sync", "error", err)
		return
	}
	c.logger.Info("Mamoru Sync", "response", response != nil)
	c.syncData = response
}

func (c *Client) Close() {
	close(c.quit)
}

func sendJSONRPCRequest(ctx context.Context, url string) (*JSONRPCResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request returned status code %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response JSONRPCResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
