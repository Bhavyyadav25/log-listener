package manticore

import (
	"fmt"
	"logListener/pkg/logging"
	"net/http"
	"sync"
	"time"

	"github.com/gookit/slog"
	mcore "github.com/manticoresoftware/manticoresearch-go"
	"golang.org/x/net/context"
)

var (
	apiClient         *mcore.APIClient
	clientInitOnce    sync.Once
	mutex             sync.Mutex
	CurrentIndex      map[string]bool
	ManticoreProtocol string
	ManticoreIP       string
	ManticorePort     string
)

// Init initializes the Manticore search client and updates the current index.
func Init(protocol, ip, port string) error {
	clientInitOnce.Do(func() {
		CurrentIndex = make(map[string]bool)
		config := mcore.NewConfiguration()
		config.Servers = mcore.ServerConfigurations{
			{
				URL: fmt.Sprintf("%s://%s:%s", protocol, ip, port),
			},
		}
		apiClient = mcore.NewAPIClient(config)
	})
	return updateCurrentIndex()
}

// EnsureConnection retries establishing a connection to Manticore and updates the current index.
func EnsureConnection() error {
	mutex.Lock()
	defer mutex.Unlock()

	const maxRetries = 5
	const retryInterval = 5 * time.Second
	const longBreakInterval = 30 * time.Second

	for {
		for retryCount := 0; retryCount < maxRetries; retryCount++ {
			if err := initClient(); err == nil {
				// Successfully connected, update the current index
				if err := updateCurrentIndex(); err != nil {
					return err
				}
				return nil
			}
			logging.ErrorLogger.Errorf("Connection attempt %d failed", retryCount+1)
			time.Sleep(retryInterval)
		}

		logging.ErrorLogger.Errorf("All %d connection attempts failed, taking a long break", maxRetries)
		time.Sleep(longBreakInterval)
	}
}

// initClient initializes the Manticore client.
func initClient() error {
	config := mcore.NewConfiguration()
	config.Servers = mcore.ServerConfigurations{
		{
			URL: fmt.Sprintf("%s://%s:%s", ManticoreProtocol, ManticoreIP, ManticorePort),
		},
	}
	apiClient = mcore.NewAPIClient(config)
	return nil
}

// ExecuteQuery executes a SQL query against the Manticore search engine.
func ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if err := EnsureConnection(); err != nil {
		slog.Error(err)
		return nil, err
	}

	slog.Debugf("Executing query: %s", query)
	mutex.Lock()
	defer mutex.Unlock()

	resp, r, err := performSQLQuery(query)
	if err != nil {
		logging.ErrorLogger.Errorf("Error executing query: %v, Full HTTP response: %v", err, r)
		return nil, err
	}

	// Refresh the current index after executing any query to ensure it is up-to-date
	if err := updateCurrentIndex(); err != nil {
		logging.ErrorLogger.Errorf("Error updating current index: %v", err)
	}

	return parseResponse(resp), nil
}

// updateCurrentIndex retrieves and updates the list of current indexes in Manticore.
func updateCurrentIndex() error {
	resp, r, err := performSQLQuery("SHOW TABLES")
	if err != nil {
		logging.ErrorLogger.Errorf("Error extracting tables from Manticore: %v, Full HTTP response: %v", err, r)
		return err
	}

	return parseCurrentIndexes(resp)
}

// parseCurrentIndexes processes the response from the "SHOW TABLES" query and updates the current indexes.
func parseCurrentIndexes(resp []map[string]interface{}) error {
	if !isValidResponse(resp) {
		return fmt.Errorf("no tables found")
	}

	// Ensure CurrentIndex map is initialized
	if CurrentIndex == nil {
		CurrentIndex = make(map[string]bool)
	}

	// Clear the current index to refresh it
	for k := range CurrentIndex {
		delete(CurrentIndex, k)
	}

	for _, r := range resp {
		if tables, ok := r["data"].([]interface{}); ok {
			for _, item := range tables {
				if tableMap, ok := item.(map[string]interface{}); ok {
					if tableName, ok := tableMap["Index"].(string); ok && len(tableName) > 0 {
						CurrentIndex[tableName] = true
					}
				}
			}
		}
	}
	return nil
}

// performSQLQuery performs a SQL query against the Manticore API.
func performSQLQuery(body string) ([]map[string]interface{}, *http.Response, error) {
	if apiClient == nil {
		return nil, nil, fmt.Errorf("apiClient is not initialized")
	}
	rawResponse := true
	slog.Debugf("SQL query: %s", body)
	slog.Debug(apiClient.UtilsAPI.Sql(context.Background()).Body(body).RawResponse(rawResponse).Execute())
	return apiClient.UtilsAPI.Sql(context.Background()).Body(body).RawResponse(rawResponse).Execute()
}

// parseResponse processes the response from a SQL query.
func parseResponse(resp []map[string]interface{}) []map[string]interface{} {
	var data []map[string]interface{}
	for _, responseItem := range resp {
		if dataSlice, ok := responseItem["data"].([]interface{}); ok {
			for _, item := range dataSlice {
				if dataMap, ok := item.(map[string]interface{}); ok {
					data = append(data, dataMap)
				}
			}
		}
	}
	return data
}

// isValidResponse checks if the response from the SQL query is valid.
func isValidResponse(resp []map[string]interface{}) bool {
	return resp != nil
}
