package main

import (
	"fmt"
	"logListener/pkg/config"
	"logListener/pkg/listener"
	"logListener/pkg/logging"
	"logListener/pkg/manticore"
	"logListener/pkg/parser"
	"sync"

	"github.com/gookit/slog"
	"github.com/robfig/cron/v3"
)

// main is the entry point of the log listener application. It loads the configuration from a YAML file,
// starts log listeners based on the configuration, and waits for all listeners to complete.
func main() {
	logFilePath := map[logging.LogType]string{
		logging.AuditLog: "logs/audit.log",
		logging.ErrorLog: "logs/error.log",
		logging.InfoLog:  "logs/info.log",
	}

	logging.InitLogger(logFilePath)

	logging.InfoLogger.Info("Starting log listener application")

	if err := config.InitConfig("pkg/config/config.yaml"); err != nil {
		logging.ErrorLogger.Fatalf("Error loading config: %v", err)
	}

	parser.EnsureGeoIpInitialized(config.Cfg.GeoDirectory, config.Cfg)

	// Setup the cron job for cleanup
	c := cron.New()
	spec := fmt.Sprintf("@%s", config.Cfg.MantiRotation)
	_, err := c.AddFunc(spec, func() {
		slog.Info("Running %s cleanup job...", config.Cfg.MantiRotation)
		manticore.CleanupOldTables()
	})
	if err != nil {
		logging.ErrorLogger.Fatalf("Error scheduling cleanup job: %v", err)
	}
	c.Start()

	var wg sync.WaitGroup
	// Start a log listener for each log configuration
	for _, logConfig := range config.Cfg.Logs {
		logging.InfoLogger.Info("Starting log listener...", "for ", logConfig.Name)
		wg.Add(1)
		go func(logConfig config.LogConfig) {
			defer wg.Done()
			defer recoverFromPanic()
			listener.StartLogListener(logConfig)
		}(logConfig)
	}

	logging.SyncLogger()

	// Wait for all log listeners to finish
	wg.Wait()

	select {}
}

// recoverFromPanic handles any panic that occurs in the application, logging the error and allowing the program to continue running.
func recoverFromPanic() {
	if r := recover(); r != nil {
		logging.ErrorLogger.Errorf("Recovered from panic: %v", r)
	}
}
