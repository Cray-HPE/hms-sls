package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	sls_client "github.com/Cray-HPE/hms-sls/v2/pkg/sls-client"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

type Worker struct {
	id     int
	logger *zap.Logger
	ctx    context.Context
	wg     *sync.WaitGroup

	slsURL         string
	slsTestNetwork string
}

func (w *Worker) Start() {
	w.wg.Add(1)
	defer w.wg.Done()

	logger := w.logger
	logger.Info("Starting worker")

	// Setup HTTP client
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = &ZapLeveledLogger{
		logger: logger.Sugar(),
	}

	slsClient := sls_client.NewSLSClient(w.slsURL, httpClient.StandardClient(), fmt.Sprintf("sls-benchmark-%d", w.id))

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("Worker done")
			return
		default:
		}

		start := time.Now()
		_, err := slsClient.GetAllHardware(w.ctx)
		if err != nil {
			logger.Error("Failed to retrieve hardware information", zap.Error(err))
		} else {
			logger.Info("Received hardware information", zap.Duration("legnth", time.Since(start)))
		}

		start = time.Now()
		_, err = slsClient.GetNetwork(w.ctx, w.slsTestNetwork)
		if err != nil {
			logger.Error("Failed to retrieve network information", zap.Error(err))
		} else {
			logger.Info("Received Network information", zap.Duration("legnth", time.Since(start)))
		}

		// time.Sleep(time.Second)
	}

}
