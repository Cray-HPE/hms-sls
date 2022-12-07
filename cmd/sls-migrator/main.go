// MIT License
//
// (C) Copyright 2022 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/namsral/flag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	base "github.com/Cray-HPE/hms-base/v2"
	sls_client "github.com/Cray-HPE/hms-sls/v2/pkg/sls-client"
)

var (
	// CLI/ENV Flags
	slsURL = flag.String("sls_url", "http://cray-sls", "URL to SLS instance.")
	dryRun = flag.Bool("dry_run", false, "Perform a dry run against SLS. Does not modify SLS")

	// Globals
	logger      *zap.Logger
	atomicLevel zap.AtomicLevel
)

func main() {
	// Setup Context
	ctx := setupContext()

	// Setup logging.
	setupLogging()

	// Configuration from environment flags
	flag.Parse()
	logger.Info("Migrator Configuration",
		zap.Stringp("sls_url", slsURL),
		zap.Boolp("dry_run", dryRun),
	)

	// Setup HTTP client
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = &ZapLeveledLogger{
		logger: logger.Sugar(),
	}

	instanceName, err := base.GetServiceInstanceName()
	if err != nil {
		logger.Warn("Can't get service instance (hostname)!  Setting to 'sls-migrator'")
		instanceName = "sls-migrator"
	}

	// Setup Migrator
	migrator := &Migrator{
		performChanges: !*dryRun,
		logger:         logger,
		slsClient:      sls_client.NewSLSClient(*slsURL, httpClient.StandardClient(), instanceName),
	}

	// Perform the migration
	if err := migrator.Run(ctx); err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to migrate SLS")
	}
}

func setupContext() context.Context {
	var cancel context.CancelFunc
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c

		// Cancel the context to cancel any in progress HTTP requests.
		cancel()
	}()

	return ctx
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

func setupLogging() {
	logLevel := os.Getenv("LOG_LEVEL")
	logLevel = strings.ToUpper(logLevel)

	atomicLevel = zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()
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
		atomicLevel.SetLevel(zap.DebugLevel)
	}
}
