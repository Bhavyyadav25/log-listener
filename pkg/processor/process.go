package processor

import (
	"encoding/json"
	"fmt"
	"logListener/pkg/config"
	"logListener/pkg/logging"
	"logListener/pkg/manticore"
	"logListener/pkg/parser"
	"strings"
	"sync"
	"time"

	"github.com/gookit/slog"
)

var (
	manticoreMutex sync.Mutex
	batchInterval  = 10 * time.Second
	batchSize      = 10
)

type LogBatch struct {
	TableName string
	Logs      []string
}

type BatchProcessor struct {
	mu         sync.Mutex
	batches    map[string]*LogBatch
	queue      chan string
	retryQueue chan string
	stopChan   chan struct{}
}

func NewBatchProcessor() *BatchProcessor {
	return &BatchProcessor{
		batches:    make(map[string]*LogBatch),
		queue:      make(chan string, 0), // Dynamic channel
		stopChan:   make(chan struct{}),
		retryQueue: make(chan string, 100),
	}
}

func (bp *BatchProcessor) Start(logConfig config.LogConfig, retryQueue chan string) {
	bp.retryQueue = retryQueue
	go bp.process(logConfig)
}

func (bp *BatchProcessor) Stop() {
	close(bp.stopChan)
}

func (bp *BatchProcessor) AddLog(log string) {
	bp.queue <- log
}

func (bp *BatchProcessor) process(logConfig config.LogConfig) {
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case log := <-bp.queue:
			bp.addLogToBatch(logConfig, log)
		case <-ticker.C:
			bp.processRemainingBatches(logConfig)
		case <-bp.stopChan:
			bp.processRemainingBatches(logConfig)
			return
		}
	}
}

func (bp *BatchProcessor) addLogToBatch(logConfig config.LogConfig, log string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	tableName := determineTableName(logConfig, log)
	if tableName == "" {
		tableName = logConfig.Name
	}

	if _, exists := bp.batches[tableName]; !exists {
		bp.batches[tableName] = &LogBatch{TableName: tableName, Logs: []string{}}
	}

	batch := bp.batches[tableName]
	batch.Logs = append(batch.Logs, log)
	slog.Debugf("Added to batch: %v, batch size: %d", batch, len(batch.Logs))

	if len(batch.Logs) >= batchSize {
		bp.insertLogBatch(logConfig, batch)
		bp.batches[tableName] = &LogBatch{TableName: tableName, Logs: []string{}}
	}
}

func (bp *BatchProcessor) processRemainingBatches(logConfig config.LogConfig) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	for _, batch := range bp.batches {
		if len(batch.Logs) > 0 {
			bp.insertLogBatch(logConfig, batch)
			batch.Logs = []string{}
		}
	}
}

// func (bp *BatchProcessor) insertLogBatch(logConfig config.LogConfig, batch *LogBatch) {
// 	manticoreMutex.Lock()
// 	defer manticoreMutex.Unlock()

// 	if err := manticore.EnsureConnection(); err != nil {
// 		bp.queueRetryLogs(batch.Logs)
// 		logging.ErrorLogger.Error(err)
// 		return
// 	}

// 	if err := createTableIfNotExists(logConfig, batch.TableName); err != nil {
// 		bp.queueRetryLogs(batch.Logs)
// 		logging.ErrorLogger.Error(err)
// 		return
// 	}

// 	var s strings.Builder
// 	s.WriteString(fmt.Sprintf("INSERT INTO %s(data) VALUES ", batch.TableName))
// 	for i, log := range batch.Logs {
// 		if i > 0 {
// 			s.WriteString(",")
// 		}
// 		s.WriteString(fmt.Sprintf("('%s')", escapeSingleQuotes(log)))
// 	}
// 	s.WriteString(";")

// 	logging.AuditLogger.Info("Executing batch insert: %s", s.String())
// 	if _, err := manticore.ExecuteQuery(s.String()); err != nil {
// 		bp.queueRetryLogs(batch.Logs)
// 		logging.ErrorLogger.Error(err)
// 	}
// }

func (bp *BatchProcessor) insertLogBatch(logConfig config.LogConfig, batch *LogBatch) {
	manticoreMutex.Lock()
	defer manticoreMutex.Unlock()

	if err := manticore.EnsureConnection(); err != nil {
		bp.queueRetryLogs(batch.Logs)
		logging.ErrorLogger.Error(err)
		return
	}

	if err := createTableIfNotExists(logConfig, batch.TableName); err != nil {
		bp.queueRetryLogs(batch.Logs)
		logging.ErrorLogger.Error(err)
		return
	}

	var s strings.Builder
	s.WriteString(fmt.Sprintf("INSERT INTO %s(data) VALUES ", batch.TableName))
	for i, log := range batch.Logs {
		if i > 0 {
			s.WriteString(",")
		}
		s.WriteString(fmt.Sprintf("('%s')", EscapeSingleQuotes(log)))
	}
	s.WriteString(";")

	logging.AuditLogger.Infof("Executing batch insert: %s", s.String())
	if _, err := manticore.ExecuteQuery(s.String()); err == nil {
		// Only clear batch after successful insertion
		bp.batches[batch.TableName] = &LogBatch{TableName: batch.TableName, Logs: []string{}}
	} else {
		bp.queueRetryLogs(batch.Logs)
		logging.ErrorLogger.Error(err)
	}
}

func (bp *BatchProcessor) queueRetryLogs(logs []string) {
	for _, log := range logs {
		bp.retryQueue <- log
	}
}

func isValidJSON(str string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

func createTableIfNotExists(logConfig config.LogConfig, tableName string) error {
	slog.Debug("Creating table if not exists: ", manticore.CurrentIndex[tableName])
	if _, ok := manticore.CurrentIndex[tableName]; !ok {
		var tableCreate string
		if logConfig.CreateTable != "" {
			tableCreate = strings.Replace(logConfig.CreateTable, "{table_name}", tableName, 1)
		} else {
			tableCreate = "create table " + tableName + "(id bigint, data json) rt_mem_limit='10M'"
		}

		if _, err := manticore.ExecuteQuery(tableCreate); err != nil {
			return err
		}
	}
	return nil
}

func EscapeSingleQuotes(input string) string {
	return strings.ReplaceAll(input, "'", "''")
}

func determineTableName(logConfig config.LogConfig, logLine string) string {
	var logData map[string]interface{}
	err := json.Unmarshal([]byte(logLine), &logData)
	if err != nil {
		logging.ErrorLogger.Error(err)
		return ""
	}

	tablePrefix := ""
	for _, rule := range logConfig.Rules {
		fieldValue, exists := logData[rule.Condition.Field]
		if exists {
			if (rule.Condition.IsEmpty && fieldValue == "") || (!rule.Condition.IsEmpty && fieldValue != "") || (rule.Condition.Contains != "" && strings.Contains(fieldValue.(string), rule.Condition.Contains)) {
				tablePrefix = rule.Condition.Name
				break
			} else {
				logging.ErrorLogger.Error(fmt.Sprintf("Field %s does not match rule %s", rule.Condition.Field, rule.Condition.Name))
			}
		}
	}

	if tablePrefix == "" {
		tablePrefix = logConfig.Name
	}

	dateSuffix := ""
	if logConfig.AppendDate {
		dateSuffix = fmt.Sprintf("_%s", time.Now().Format("2006_01_02"))
	}
	return fmt.Sprintf("%s%s", tablePrefix, dateSuffix)
}

func StartWorker(logConfig config.LogConfig, queue chan string, workerWG *sync.WaitGroup, workerID int, bp *BatchProcessor) {
	defer workerWG.Done()
	defer recoverFromWorkerPanic()

	for line := range queue {
		slog.Debugf("Worker %d processing log content: %v", workerID, line)
		if logConfig.IsJSON {
			if isValidJSON(line) {
				slog.Debug("Valid JSON")
				updatedLine, err := updateTimestamp(line)
				if err != nil {
					logging.ErrorLogger.Errorf("Error while updating timestamp: %v", err)
				}
				modifiedLog, err := parser.ProcessLogWithGeo(updatedLine, logConfig)
				if err != nil {
					logging.ErrorLogger.Errorf("Error processing log with geo: %v", err)
				}
				bp.AddLog(modifiedLog)
			} else {
				slog.Debug("Invalid JSON")
			}
		} else {
			parsedLog, err := parser.ParseErrorLog(line)
			if err != nil {
				logging.ErrorLogger.Error(err)
				continue
			}

			parsedLog.Timestamp = time.Now().Unix()
			modifiedJSON, err := json.Marshal(parsedLog)
			if err != nil {
				logging.ErrorLogger.Error("error marshalling JSON:", err)
				continue
			}

			modifiedLog, err := parser.ProcessLogWithGeo(string(modifiedJSON), logConfig)
			if err != nil {
				logging.ErrorLogger.Errorf("Error processing log with geo: %v", err)
			}

			bp.AddLog(modifiedLog)
		}
	}
}

func recoverFromWorkerPanic() {
	if r := recover(); r != nil {
		logging.ErrorLogger.Errorf("Recovered from worker panic: %v", r)
	}
}

func updateTimestamp(jsonStr string) (string, error) {
	var data map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	// if time is not ok, deal with it!
	timestamp, exists := data["timestamp"]
	if !exists {
		data["timestamp"] = time.Now().Unix()
	} else if exists {
		if tsStr, ok := timestamp.(string); ok {
			parsedTime, err := time.Parse(time.RFC3339, tsStr)
			if err != nil {
				return "", err
			}
			data["timestamp"] = parsedTime.Unix()
		}
	}

	updatedJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(updatedJSON), nil
}
