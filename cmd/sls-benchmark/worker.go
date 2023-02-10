// MIT License
//
// (C) Copyright [2023] Hewlett Packard Enterprise Development LP
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

		// time.Sleep(100 * time.Millisecond)
	}

}
