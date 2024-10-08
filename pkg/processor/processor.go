package processor

// import (
// 	"encoding/json"
// 	"fmt"
// 	"logListener/pkg/config"
// 	"logListener/pkg/manticore"
// 	"logListener/pkg/parser"
// 	"reflect"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/gookit/slog"
// )

// var manticoreMutex sync.Mutex

// func StartWorker(logConfig config.LogConfig, queue chan interface{}, workerWG *sync.WaitGroup, workerID int) {
// 	defer workerWG.Done()

// 	logging.Logger.Info("Worker %d started", workerID)
// 	for line := range queue {
// 		slog.Debugf("Worker %d processing log content: %v", workerID, line)

// 		//TODO: Tag based tablename or use default
// 		//TODO: Add log rotation for manticore
// 		if logConfig.IsJSON {
// 			if isValidJSON(line.(string)) {
// 				slog.Debug("Valid JSON")
// 				processLogLine(logConfig, line.(string))
// 			} else {
// 				parsedLog, err := parser.ParseErrorLog(line.(string))
// 				if err != nil {
// 					logging.Logger.Error(err)
// 					continue
// 				}

// 				parsedLog.Timestamp = time.Now().Format("2006-01-02T15:04:05Z07:00")
// 				modifiedJSON, err := json.Marshal(parsedLog)
// 				if err != nil {
// 					logging.Logger.Error("error marshalling JSON:", err)
// 					continue
// 				}

// 				processLogLine(logConfig, string(modifiedJSON))
// 			}
// 		}
// 	}
// 	logging.Logger.Info("Worker %d stopped", workerID)
// }

// func processLogLine(logConfig config.LogConfig, line string) {
// 	tableName := determineTableName(logConfig, line)
// 	if tableName == "" {
// 		tableName = logConfig.Name
// 	}

// 	manticoreMutex.Lock()
// 	defer manticoreMutex.Unlock()

// 	manticore.IsManticore = false

// 	if err := manticore.EnsureManticoreConnection(); err != nil {
// 		logging.Logger.Error(err)
// 		return
// 	}

// 	if err := createTableIfNotExists(logConfig, tableName); err != nil {
// 		logging.Logger.Error(err)
// 		return
// 	}

// 	if err := insertLogLine(logConfig, line, tableName); err != nil {
// 		logging.Logger.Error(err)
// 		return
// 	}
// }

// func isValidJSON(str string) bool {
// 	var js map[string]interface{}
// 	return json.Unmarshal([]byte(str), &js) == nil
// }

// func createTableIfNotExists(logConfig config.LogConfig, tableName string) error {
// 	if _, ok := manticore.CurrentIndex[tableName]; !ok {
// 		var tableCreate string
// 		if logConfig.CreateTable != "" {
// 			tableCreate = strings.Replace(logConfig.CreateTable, "{table_name}", tableName, 1)
// 		} else {
// 			tableCreate = "create table " + tableName + "(id bigint,data json) rt_mem_limit='10M'"
// 		}

// 		if _, err := manticore.ExecManticore(tableCreate); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func insertLogLine(logConfig config.LogConfig, line, tableName string) error {
// 	var s strings.Builder
// 	currentDate := ""
// 	columnString := "data"
// 	logDate, err := timeStamp(line)
// 	if err != nil {
// 		return err
// 	}

// 	finalLog, err := updateTimestamp(line)
// 	if err != nil {
// 		logging.Logger.Error(err)
// 	}

// 	slog.Info("Inserting into table: ", finalLog)

// 	if currentDate == "" {
// 		currentDate = logDate
// 		s.WriteString("INSERT INTO ")
// 		s.WriteString(tableName)
// 		s.WriteString("(" + columnString + ") VALUES ('" + finalLog + "')")
// 	}
// 	// TODO: Add mutliple queries together with timeout
// 	slog.Print(s.String())

// 	if _, err := manticore.ExecManticore(s.String()); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func timeStamp(jsonStr string) (string, error) {
// 	var js map[string]interface{}
// 	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
// 		return "", err
// 	}

// 	if timestamp, exists := js["timestamp"]; exists {
// 		if tsStr, ok := timestamp.(string); ok {
// 			parsedTime, err := time.Parse(time.RFC3339, tsStr)
// 			if err != nil {
// 				return "", err
// 			}
// 			return parsedTime.Format("2006_01_02"), nil
// 		} else {
// 			return "", fmt.Errorf("timestamp is not a string, but a %s", reflect.TypeOf(timestamp))
// 		}
// 	} else {
// 		return time.Now().Format("2006_01_02"), nil
// 	}
// }

// func updateTimestamp(jsonStr string) (string, error) {
// 	var data map[string]interface{}

// 	err := json.Unmarshal([]byte(jsonStr), &data)
// 	if err != nil {
// 		return "", err
// 	}

// 	// if time is not ok, deal with it!
// 	timestamp, exists := data["timestamp"]
// 	if !exists {
// 		data["timestamp"] = time.Now().Unix()
// 	} else if exists {
// 		if tsStr, ok := timestamp.(string); ok {
// 			parsedTime, err := time.Parse(time.RFC3339, tsStr)
// 			if err != nil {
// 				return "", err
// 			}
// 			data["timestamp"] = parsedTime.Unix()
// 		}
// 	}

// 	updatedJSON, err := json.Marshal(data)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(updatedJSON), nil
// }

// func determineTableName(logConfig config.LogConfig, logLine string) string {
// 	var logData map[string]interface{}
// 	err := json.Unmarshal([]byte(logLine), &logData)
// 	if err != nil {
// 		logging.Logger.Error(err)
// 		return ""
// 	}

// 	tablePrefix := ""
// 	for _, rule := range logConfig.Rules {
// 		fieldValue, exists := logData[rule.Condition.Field]
// 		if exists {
// 			if (rule.Condition.IsEmpty && fieldValue == "") || (!rule.Condition.IsEmpty && fieldValue != "") || (rule.Condition.Contains != "" && strings.Contains(fieldValue.(string), rule.Condition.Contains)) {
// 				tablePrefix = rule.Condition.Name
// 				break
// 			} else {
// 				logging.Logger.Error(fmt.Sprintf("Field %s does not match rule %s", rule.Condition.Field, rule.Condition.Name))
// 			}
// 		}
// 	}

// 	if tablePrefix == "" {
// 		tablePrefix = logConfig.Name
// 	}

// 	dateSuffix := ""
// 	if logConfig.AppendDate {
// 		dateSuffix = time.Now().Format("2006_01_02")
// 	}
// 	return fmt.Sprintf("%s_%s", tablePrefix, dateSuffix)
// }
