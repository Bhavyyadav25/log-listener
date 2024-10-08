package config

import "github.com/spf13/viper"

// LogConfig holds the configuration settings for a log listener.
// Fields:
//
//	Name            - Name of the log configuration.
//	Address         - Address to listen for log data.
//	Protocol        - Protocol to use (UDP or TCP).
//	Queue           - Size of the log queue.
//	WorkerProcesses - Number of worker processes for handling logs.
//	CreateTable     - SQL query to create the log table.
//	IsJSON          - Indicates if logs are in JSON format.
//	AppendDate      - Indicates if date should be appended to the table name.
type LogConfig struct {
	Name            string `yaml:"name"`
	Address         string `yaml:"address"`
	Protocol        string `yaml:"protocol"`
	Queue           int    `yaml:"queue"`
	WorkerProcesses int    `yaml:"workerProcesses"`
	CreateTable     string `yaml:"createTable"`
	IsJSON          bool   `yaml:"isJson"`
	AppendDate      bool   `yaml:"appendDate"`
	Rules           []Rule `yaml:"rules"`
	Geo             []Geo  `yaml:"geo"`
}

type Geo struct {
	Field string `yaml:"field"`
	Tag   string `yaml:"tag"`
}

type Rule struct {
	Condition Condition `yaml:"condition,omitempty"`
}

type Condition struct {
	Field    string `yaml:"field"`
	IsEmpty  bool   `yaml:"isEmpty,omitempty"`
	Contains string `yaml:"contains,omitempty"`
	Name     string `yaml:"name"`
}

// Config holds the overall configuration for the application,
// including multiple log configurations.
type Config struct {
	Logs          []LogConfig `yaml:"logs"`
	GeoDirectory  string      `yaml:"geoDirectory"`
	MantiRotation string      `yaml:"mantiRotation"`
	MantiProtocol string      `yaml:"mantiProtocol"`
	MantiIP       string      `yaml:"mantiIP"`
	MantiPort     string      `yaml:"mantiPort"`
}

var Cfg Config

// InitConfig initializes the configuration by reading the specified
// YAML file. It uses Viper to read and unmarshal the configuration
// into the Cfg variable. Returns an error if the configuration cannot
// be read or parsed.
func InitConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		return err
	}

	return nil
}
