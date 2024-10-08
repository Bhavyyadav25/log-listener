package manticore

// import (
// 	"fmt"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"github.com/gookit/slog"
// 	manticore "github.com/manticoresoftware/manticoresearch-go"
// 	"golang.org/x/net/context"
// )

// var (
// 	apiClient         *manticore.APIClient
// 	ManticoreMutex    sync.Mutex
// 	CurrentIndex      map[string]bool = make(map[string]bool)
// 	mutexManti        sync.Mutex
// 	IsManticore       bool
// 	ManticoreProtocol string
// 	ManticoreIP       string
// 	ManticorePort     string
// )

// // InitManti initializes the connection to the Manticore search engine.
// func InitManti() {
// 	if IsManticore {
// 		return
// 	}
// 	initializeMantiClient()
// 	IsManticore = true
// 	CurrentIndex = make(map[string]bool)

// 	if err := UpdateCurrentIndex(); err != nil {
// 		slog.Println("Error updating current index: ", err)
// 		IsManticore = false
// 	}
// }

// func initializeMantiClient() {
// 	configuration := manticore.NewConfiguration()
// 	configuration.Servers[0].URL = fmt.Sprintf("%s://%s:%s", ManticoreProtocol, ManticoreIP, ManticorePort)
// 	apiClient = manticore.NewAPIClient(configuration)
// }

// // ensureManticoreConnection checks and establishes a Manticore connection if necessary.
// func EnsureManticoreConnection() error {
// 	if !IsManticore {
// 		InitManti()
// 	}

// 	err := CheckConnection()
// 	if err != nil {
// 		logging.Logger.Error("Re-initializing Manticore in 5 seconds due to connection error:", err)
// 		IsManticore = false
// 		time.Sleep(5 * time.Second)
// 		InitManti()
// 		if !IsManticore {
// 			return fmt.Errorf("failed to initialize Manticore after connection error")
// 		}
// 	}

// 	return nil
// }

// // ExecManticore executes a given SQL query against the Manticore search engine.
// func ExecManticore(query string) ([]map[string]interface{}, error) {
// 	if err := EnsureManticoreConnection(); err != nil {
// 		return nil, err
// 	}

// 	slog.Println(query)
// 	mutexManti.Lock()
// 	defer mutexManti.Unlock()

// 	resp, r, err := performSqlQuery(query)
// 	if err != nil {
// 		logging.Logger.Errorf("Error executing query: %v, Full HTTP response: %v", err, r)
// 		return nil, err
// 	}

// 	return parseManticoreResponse(resp), nil
// }

// func parseManticoreResponse(resp []map[string]interface{}) []map[string]interface{} {
// 	var data []map[string]interface{}
// 	for _, responseItem := range resp {
// 		if dataSlice, ok := responseItem["data"].([]interface{}); ok {
// 			for _, item := range dataSlice {
// 				if dataMap, ok := item.(map[string]interface{}); ok {
// 					data = append(data, dataMap)
// 				}
// 			}
// 		}
// 	}
// 	return data
// }

// // UpdateCurrentIndex retrieves and updates the list of current indexes in Manticore.
// func UpdateCurrentIndex() error {
// 	if err := EnsureManticoreConnection(); err != nil {
// 		return err
// 	}

// 	mutexManti.Lock()
// 	defer mutexManti.Unlock()

// 	resp, r, err := performSqlQuery("SHOW TABLES")
// 	if err != nil {
// 		logging.Logger.Errorf("Error extracting tables from Manticore: %v, Full HTTP response: %v", err, r)
// 		IsManticore = false
// 		return err
// 	}

// 	return updateCurrentIndexesFromResponse(resp)
// }

// // updateCurrentIndexesFromResponse processes the response from the "SHOW TABLES" query and updates the current indexes.
// func updateCurrentIndexesFromResponse(resp []map[string]interface{}) error {
// 	if !isResponseValid(resp) {
// 		slog.Println("Received no tables from Manticore")
// 		return fmt.Errorf("no tables found")
// 	}

// 	for _, r := range resp {
// 		if tables, ok := r["data"].([]interface{}); ok {
// 			processTables(tables)
// 		}
// 	}
// 	return nil
// }

// // isResponseValid checks if the response from the "SHOW TABLES" query is valid.
// func isResponseValid(resp []map[string]interface{}) bool {
// 	return resp != nil
// }

// // processTables processes the tables data to update the CurrentIndex map.
// func processTables(tables []interface{}) {
// 	for _, item := range tables {
// 		if tableMap, ok := item.(map[string]interface{}); ok {
// 			if tableName, ok := tableMap["Index"].(string); ok && len(tableName) > 0 {
// 				CurrentIndex[tableName] = true
// 			}
// 		}
// 	}
// }

// // performSqlQuery performs a SQL query against the Manticore API.
// func performSqlQuery(body string) ([]map[string]interface{}, *http.Response, error) {
// 	if apiClient == nil {
// 		InitManti()
// 	}
// 	rawResponse := true
// 	slog.Debug(body)
// 	return apiClient.UtilsAPI.Sql(context.Background()).Body(body).RawResponse(rawResponse).Execute()
// }

// // CheckConnection performs a SQL query to check the connection status against the Manticore API.
// func CheckConnection() error {
// 	body := "SHOW TABLES"
// 	_, _, err := performSqlQuery(body)
// 	return err
// }
