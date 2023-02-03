package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/namsral/flag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger      *zap.Logger
	atomicLevel zap.AtomicLevel
)

func setupLogging() {
	logLevel := os.Getenv("LOG_LEVEL")
	logLevel = strings.ToUpper(logLevel)

	atomicLevel = zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	logger = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atomicLevel,
	))

	switch logLevel {
	case "DEBUG":
		atomicLevel.SetLevel(zap.DebugLevel)
	case "INFO":
		atomicLevel.SetLevel(zap.InfoLevel)
	case "WARN":
		atomicLevel.SetLevel(zap.WarnLevel)
	case "ERROR":
		atomicLevel.SetLevel(zap.ErrorLevel)
	case "FATAL":
		atomicLevel.SetLevel(zap.FatalLevel)
	case "PANIC":
		atomicLevel.SetLevel(zap.PanicLevel)
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
	}
}

// ZapLeveledLogger allow the Zap logger to be compatible with retryablehttp by implementing its LeveledLogger interface.
type ZapLeveledLogger struct {
	logger *zap.SugaredLogger
}

func (z *ZapLeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	z.logger.Errorw(msg, keysAndValues...)
}

func (z *ZapLeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	z.logger.Infow(msg, keysAndValues...)
}

func (z *ZapLeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	z.logger.Debugw(msg, keysAndValues...)
}

func (z *ZapLeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	z.logger.Warnw(msg, keysAndValues...)
}

func main() {
	// Parse CLI flag configuration
	workerCount := flag.Int("worker_count", 10, "Number of event workers")
	slsURL := flag.String("sls_url", "http://localhost:8376", "SLS URL")
	testNetwork := flag.String("test_network", "CHN", "SLS network to test against")
	flag.Parse()

	// Setup logging
	setupLogging()

	// Setup signal handler
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// workers := []*Worker{}

	var workerWg sync.WaitGroup
	workerCtx, workerCancel := context.WithCancel(context.Background())
	for id := 0; id < *workerCount; id++ {
		worker := &Worker{
			id:     id,
			logger: logger.With(zap.Int("WorkerID", id)),
			ctx:    workerCtx,
			wg:     &workerWg,

			slsURL:         *slsURL,
			slsTestNetwork: *testNetwork,
		}
		// workers = append(workers, worker)

		go worker.Start()
	}

	// Wait for signal
	sig := <-sigchan
	fmt.Printf("Caught signal %v: terminating\n", sig)

	// Stop the workers
	logger.Info("Stopping workers")
	workerCancel()
	workerWg.Wait()
	logger.Info("All workers completed")
}
