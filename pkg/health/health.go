package health

import (
	"encoding/json"
	"fmt"
	"log"
	"logListener/pkg/manticore"
	"logListener/pkg/processor"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

// Struct to hold all the information for JSON output
type SystemInfo struct {
	Timestamp       int64        `json:"timestamp"`
	CPUUsage        float64      `json:"cpu_usage"`
	MemoryUsage     float64      `json:"memory_usage"`
	DiskUsage       float64      `json:"disk_usage"`
	ServersStatuses []ServerInfo `json:"server_statuses"`
}

// Struct to hold individual server status information
type ServerInfo struct {
	ServerIP   string `json:"server_ip"`
	IsUp       bool   `json:"is_up"`
	IsInConfig bool   `json:"is_in_config"`
}

// Get CPU usage with a short delay for accurate results
func getCPUUsage() float64 {
	percentages, err := cpu.Percent(200*time.Millisecond, false) // Adding a delay of 200ms for better accuracy
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}
	if len(percentages) > 0 {
		return percentages[0]
	}
	return 0
}

// Get memory usage
func getMemoryUsage() float64 {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalf("Error getting memory usage: %v", err)
	}
	return v.UsedPercent
}

// Get disk usage with manual calculation
func getDiskUsage() float64 {
	d, err := disk.Usage("/")
	if err != nil {
		slog.Fatalf("Error getting disk usage: %v", err)
	}

	// Manually calculate the percentage of used space
	usedPercent := (100 - (float64(d.Free)/float64(d.Total))*100)
	return usedPercent
}

// Parse Nginx reverse proxy config to extract server IPs
func parseReverseProxyConfig(path string) ([]string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	content := string(file)
	lines := strings.Split(content, "\n")
	var servers []string

	for _, line := range lines {
		if strings.Contains(line, "upstream") || strings.Contains(line, "server") {
			if strings.Contains(line, "fail_timeout") {
				// Extract IP:port
				parts := strings.Fields(line)
				if len(parts) > 1 {
					serverIP := strings.Split(parts[1], ":")[0]
					servers = append(servers, serverIP)
				}
			}
		}
	}
	return servers, nil
}

// Check if server is up
func isServerUp(server string) bool {
	conn, err := net.Dial("tcp", server+":443")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Health generates and prints system health in JSON format
func Health() {
	// Get system stats
	cpuUsage := getCPUUsage()
	memUsage := getMemoryUsage()
	diskUsage := getDiskUsage()

	// Parse reverse proxy config
	servers, err := parseReverseProxyConfig("/etc/nginx/sites-enabled/reverse-proxy.conf")
	if err != nil {
		slog.Fatalf("Error parsing reverse proxy config: %v", err)
	}

	// Check the status of each server
	var serverStatuses []ServerInfo
	for _, server := range servers {
		serverStatus := ServerInfo{
			ServerIP:   server,
			IsInConfig: true,
			IsUp:       isServerUp(server),
		}
		serverStatuses = append(serverStatuses, serverStatus)
	}

	// Compile all data into a struct
	systemInfo := SystemInfo{
		Timestamp:       time.Now().Unix(),
		CPUUsage:        cpuUsage,
		MemoryUsage:     memUsage,
		DiskUsage:       diskUsage,
		ServersStatuses: serverStatuses,
	}

	// Convert systemInfo to JSON
	jsonData, err := json.MarshalIndent(systemInfo, "", "  ")
	if err != nil {
		slog.Fatalf("Error converting to JSON: %v", err)
	}

	// Print the JSON
	slog.Println(string(jsonData))

	// Ensure Manticore connection
	if err := manticore.EnsureConnection(); err != nil {
		slog.Error(err)
		return
	}

	// Table name for the log, based on the current date
	tableName := "health_log_" + time.Now().Format("2006_01_02")

	slog.Info(manticore.CurrentIndex[tableName])
	// Check if the table already exists in the Manticore index
	if _, exists := manticore.CurrentIndex[tableName]; !exists {
		slog.Info("Table does not exist, creating it... : ", tableName)
		// Create table only if it doesn't exist
		tableCreate := fmt.Sprintf("CREATE TABLE %s (id bigint, data json) rt_mem_limit='10M'", tableName)
		if _, err := manticore.ExecuteQuery(tableCreate); err != nil {
			slog.Error("Error creating table:", err)
			return
		}
	}

	// Insert the log data only once
	query := fmt.Sprintf("INSERT INTO %s(data) VALUES ('%s')", tableName, processor.EscapeSingleQuotes(string(jsonData)))
	if _, err := manticore.ExecuteQuery(query); err != nil {
		slog.Error("Error inserting data into table:", err)
		return
	}

	slog.Info("Health check completed successfully")
}
