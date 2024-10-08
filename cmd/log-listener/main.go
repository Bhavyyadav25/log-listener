package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"logListener/pkg/config"
// 	"logListener/pkg/manticore"
// 	"reflect"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/gookit/slog"
// 	"gopkg.in/mcuadros/go-syslog.v2"
// )

// var manticoreMutex sync.Mutex

// func main() {
// 	if err := config.InitConfig("pkg/config/config.yaml"); err != nil {
// 		slog.Fatalf("Error loading config: %v", err)
// 	}

// 	var wg sync.WaitGroup
// 	for _, logConfig := range config.Cfg.Logs {
// 		slog.Info("Starting log listener.....")
// 		slog.Debug(logConfig)
// 		wg.Add(1)
// 		go func(logConfig config.LogConfig) {
// 			defer wg.Done()
// 			ProcessLogListener(logConfig)
// 		}(logConfig)
// 	}

// 	wg.Wait()
// }

// func ProcessLogListener(logConfig config.LogConfig) {
// 	channel := make(syslog.LogPartsChannel)
// 	handler := syslog.NewChannelHandler(channel)

// 	server := syslog.NewServer()
// 	server.SetFormat(syslog.Automatic)
// 	server.SetHandler(handler)

// 	var err error
// 	switch logConfig.Protocol {
// 	case "UDP":
// 		err = server.ListenUDP(logConfig.Address)
// 	case "TCP":
// 		err = server.ListenTCP(logConfig.Address)
// 	default:
// 		logging.Logger.Error("Protocol not supported")
// 		return
// 	}

// 	if err != nil {
// 		logging.Logger.Errorf("Error listening: %v", err)
// 		return
// 	}

// 	if err := server.Boot(); err != nil {
// 		logging.Logger.Errorf("Error booting: %v", err)
// 		return
// 	}

// 	queue := make(chan interface{}, logConfig.Queue)
// 	var workerWG sync.WaitGroup

// 	for i := 0; i < logConfig.WorkerProcesses; i++ {
// 		workerWG.Add(1)
// 		go worker(logConfig, queue, &workerWG, i)
// 	}

// 	go func() {
// 		for logParts := range channel {
// 			if logLine, ok := logParts["content"]; ok {
// 				queue <- logLine
// 			}
// 		}
// 		close(queue)
// 	}()

// 	server.Wait()
// 	workerWG.Wait()
// }

// func worker(logConfig config.LogConfig, queue chan interface{}, workerWG *sync.WaitGroup, workerID int) {
// 	defer workerWG.Done()

// 	logging.Logger.Info("Worker %d started", workerID)
// 	for line := range queue {
// 		slog.Debugf("Worker %d processing log content: %v", workerID, line)

// 		// TODO: Process log line
// 		if logConfig.IsJSON {
// 			if isValidJSON(line.(string)) {
// 				var tableName string
// 				slog.Debug("Valid JSON")
// 				if logConfig.AppendDate {
// 					tableName = fmt.Sprintf("%s_%s", logConfig.Name, time.Now().Format("2006_01_02"))
// 					slog.Info("Creating table: ", tableName)
// 				} else {
// 					tableName = logConfig.Name
// 				}

// 				manticoreMutex.Lock()
// 				manticore.IsManticore = false

// 				if err := manticore.EnsureManticoreConnection(); err != nil {
// 					logging.Logger.Error(err)
// 					manticoreMutex.Unlock()
// 					return
// 				}

// 				var tableCreate string
// 				if _, ok := manticore.CurrentIndex[tableName]; !ok {
// 					if logConfig.CreateTable != "" {
// 						tableCreate = strings.Replace(logConfig.CreateTable, "{table_name}", tableName, 1)
// 					} else {
// 						tableCreate = "create table " + tableName + "(id bigint,data json)"
// 					}

// 					if _, err := manticore.ExecManticore(tableCreate); err != nil {
// 						logging.Logger.Error(err)
// 					}
// 				}

// 				var s strings.Builder
// 				currentDate := ""
// 				columnString := "data"
// 				logDate, err := timeStamp(line.(string))
// 				if err != nil {
// 					logging.Logger.Error(err)
// 				}

// 				if currentDate == "" {
// 					currentDate = logDate
// 					s.WriteString("INSERT INTO ")
// 					s.WriteString(tableName)
// 					s.WriteString("(" + columnString + ") VALUES ('" + line.(string) + "')")
// 				}

// 				slog.Print(s.String())

// 				if _, err := manticore.ExecManticore(s.String()); err != nil {
// 					logging.Logger.Error(err)
// 					manticoreMutex.Unlock()
// 					return
// 				}
// 				manticoreMutex.Unlock()
// 			}
// 		}
// 	}
// 	logging.Logger.Info("Worker %d stopped", workerID)
// }

// func isValidJSON(str string) bool {
// 	var js map[string]interface{}
// 	return json.Unmarshal([]byte(str), &js) == nil
// }

// func timeStamp(jsonStr string) (string, error) {

// 	// TODO: Get timestamp
// 	var js map[string]interface{}
// 	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
// 		return "", err
// 	}

// 	if timestamp, exists := js["timestamp"]; exists {
// 		// Assuming the timestamp is a string
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
