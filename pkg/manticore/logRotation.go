package manticore

import (
	"fmt"
	"logListener/pkg/logging"
	"strings"
	"time"

	"github.com/gookit/slog"
)

const dateFormat = "2006_01_02"

// CleanupOldTables checks for tables with a date suffix older than 30 days and drops them.
func CleanupOldTables() {
	if err := EnsureConnection(); err != nil {
		logging.ErrorLogger.Errorf("Error ensuring Manticore connection: %v", err)
		return
	}

	now := time.Now()
	thresholdDate := now.AddDate(0, 0, -30) // 30 days ago

	mutex.Lock()
	defer mutex.Unlock()

	slog.Info("Cleaning up old tables... : ", CurrentIndex)
	for tableName := range CurrentIndex {
		tableDate, err := extractDateFromTableName(tableName)
		if err != nil {
			slog.Warnf("Skipping table %s due to error: %v", tableName, err)
			continue
		}

		if tableDate.Before(thresholdDate) {
			dropTable(tableName)
		}
	}
}

func extractDateFromTableName(tableName string) (time.Time, error) {
	slog.Debug("Extracting date from table name: ", tableName)
	parts := strings.Split(tableName, "_")
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("table name %s does not contain a valid date", tableName)
	}

	dateStr := parts[len(parts)-3] + "_" + parts[len(parts)-2] + "_" + parts[len(parts)-1]
	return time.Parse(dateFormat, dateStr)
}

func dropTable(tableName string) {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	if _, err := ExecuteQuery(query); err != nil {
		logging.ErrorLogger.Errorf("Error dropping table %s: %v", tableName, err)
	} else {
		delete(CurrentIndex, tableName)
		logging.AuditLogger.Info("Dropped table %s", tableName)
	}
}
