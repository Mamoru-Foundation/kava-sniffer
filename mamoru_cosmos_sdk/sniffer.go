package mamoru_cosmos_sdk

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/go-kit/log/level"
	"github.com/go-kit/log/term"

	"github.com/Mamoru-Foundation/mamoru-sniffer-go/mamoru_sniffer"
	"github.com/Mamoru-Foundation/mamoru-sniffer-go/mamoru_sniffer/cosmos"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/kava-labs/kava/mamoru_cosmos_sdk/sync_state"
)

const (
	PolishTimeSec   = 10
	DefaultTNApiUrl = "http://localhost:26657/status"
)

var snifferConnectFunc = cosmos.CosmosConnect

func InitConnectFunc(f func() (*cosmos.SnifferCosmos, error)) {
	snifferConnectFunc = f
}

func init() {
	mamoru_sniffer.InitLogger(func(entry mamoru_sniffer.LogEntry) {
		kvs := mapToInterfaceSlice(entry.Ctx)
		msg := "Mamoru core: " + entry.Message
		var tmLogger = log.NewTMLoggerWithColorFn(os.Stdout, func(keyvals ...interface{}) term.FgBgColor {
			if keyvals[0] != level.Key() {
				panic(fmt.Sprintf("expected level key to be first, got %v", keyvals[0]))
			}
			switch keyvals[1].(level.Value).String() {
			case "debug":
				return term.FgBgColor{Fg: term.Green}
			case "error":
				return term.FgBgColor{Fg: term.DarkRed}
			default:
				return term.FgBgColor{}
			}
		})

		switch entry.Level {
		case mamoru_sniffer.LogLevelDebug:
			tmLogger.Debug(msg, kvs...)
		case mamoru_sniffer.LogLevelInfo:
			tmLogger.Info(msg, kvs...)
		case mamoru_sniffer.LogLevelWarning:
			tmLogger.With("Warn").Error(msg, kvs...)
		case mamoru_sniffer.LogLevelError:
			tmLogger.Error(msg, kvs...)
		}
	})
}

func mapToInterfaceSlice(m map[string]string) []interface{} {
	var result []interface{}
	for key, value := range m {
		result = append(result, key, value)
	}

	return result
}

type Sniffer struct {
	mu     sync.Mutex
	logger log.Logger
	client *cosmos.SnifferCosmos
	sync   *sync_state.Client
}

func NewSniffer(logger log.Logger) *Sniffer {
	tmApiUrl := getEnv("MAMORU_TM_API_URL", DefaultTNApiUrl)
	httpClient := sync_state.NewHTTPRequest(logger, tmApiUrl, PolishTimeSec, isSnifferEnabled())

	return &Sniffer{
		logger: logger,
		sync:   httpClient,
	}
}

// IsSynced returns true if the sniffer is synced with the chain
func (s *Sniffer) IsSynced() bool {
	s.logger.Info("Mamoru Sniffer sync", "sync", s.sync.GetSyncData().IsSync(),
		"block", s.sync.GetSyncData().GetCurrentBlockNumber())

	return s.sync.GetSyncData().IsSync()
}

func (s *Sniffer) CheckRequirements() bool {
	return isSnifferEnabled() && s.IsSynced() && s.connect()
}

func (s *Sniffer) Client() *cosmos.SnifferCosmos {
	return s.client
}

func (s *Sniffer) connect() bool {
	if s.client != nil {
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	s.client, err = snifferConnectFunc()
	if err != nil {
		s.logger.Error("Mamoru Sniffer connect", "err", err)
		return false
	}

	return true
}

func isSnifferEnabled() bool {
	val, _ := strconv.ParseBool(getEnv("MAMORU_SNIFFER_ENABLE", "false"))
	return val
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
