package listener

import (
	"logListener/pkg/config"
	"logListener/pkg/logging"
	"logListener/pkg/processor"
	"sync"
	"time"

	"github.com/gookit/slog"
	"gopkg.in/mcuadros/go-syslog.v2"
)

// StartLogListener initializes and starts a syslog server based on the provided
// configuration. It supports both UDP and TCP protocols and processes incoming
// log entries using a worker pool.
//
// Parameters:
// - logConfig: configuration parameters for the syslog listener including address,
// protocol, queue size, and worker processes.
func StartLogListener(logConfig config.LogConfig) {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)

	var err error
	switch logConfig.Protocol {
	case "UDP":
		err = server.ListenUDP(logConfig.Address)
	case "TCP":
		err = server.ListenTCP(logConfig.Address)
	default:
		logging.ErrorLogger.Error("Protocol not supported")
		return
	}

	if err != nil {
		logging.ErrorLogger.Errorf("Error listening: %v", err)
		return
	}

	if err := server.Boot(); err != nil {
		logging.ErrorLogger.Errorf("Error booting: %v", err)
		return
	}

	queue := make(chan string, logConfig.Queue)
	retryQueue := make(chan string, logConfig.Queue)
	var workerWG sync.WaitGroup
	batchProcessor := processor.NewBatchProcessor()
	batchProcessor.Start(logConfig, retryQueue)
	defer batchProcessor.Stop()

	for i := 0; i < logConfig.WorkerProcesses; i++ {
		workerWG.Add(1)
		go processor.StartWorker(logConfig, queue, &workerWG, i, batchProcessor)
	}

	go func() {
		for logParts := range channel {
			slog.Debug("Received log line: %v", logParts)
			if logLine, ok := logParts["content"]; ok {
				queue <- logLine.(string)
			}
		}
		close(queue)
	}()

	go processRetryQueue(retryQueue, batchProcessor)

	server.Wait()
	workerWG.Wait()

}

func processRetryQueue(retryQueue chan string, bp *processor.BatchProcessor) {
	for log := range retryQueue {
		time.Sleep(5 * time.Second) // Consider exponential backoff instead of fixed sleep
		bp.AddLog(log)
	}
}

// func processRetryQueue(ctx context.Context, retryQueue chan string, bp *processor.BatchProcessor, maxRetries int) {
// 	for {
// 		select {
// 		case log := <-retryQueue:
// 			for retry := 0; retry < maxRetries; retry++ {
// 				slog.Info("Retrying log: ", retry)
// 				time.Sleep(30 * time.Second) // Consider exponential backoff instead of fixed sleep
// 				bp.AddLog(log)
// 				if retry == maxRetries-1 {
// 					logging.Logger.Errorf("Persistent failure for log: %v after %d retries", log, maxRetries)
// 					// Optionally send to a dead-letter queue or log it for manual review
// 				}
// 			}
// 		case <-ctx.Done():
// 			return // Graceful shutdown
// 		}
// 	}
// }
