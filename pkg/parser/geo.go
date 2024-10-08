package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"logListener/pkg/config"
	"logListener/pkg/logging"
	"logListener/pkg/manticore"
	"net"
	"sync"

	"github.com/gookit/slog"
	"github.com/oschwald/maxminddb-golang"
)

var (
	maxdb     *maxminddb.Reader
	geoIPInit sync.Once
)

func ProcessLogWithGeo(logStr string, logConfig config.LogConfig) (string, error) {
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logStr), &logData); err != nil {
		return "", fmt.Errorf("error unmarshalling log: %v", err)
	}

	if logConfig.Geo == nil {
		// No geo configuration present, return the original log
		return logStr, nil
	}

	for _, geoField := range logConfig.Geo {
		var countryRecord struct {
			Country struct {
				IsoCode string `maxminddb:"iso_code"`
				Name    string `maxminddb:"name"`
				Names   struct {
					En string `maxminddb:"en"`
				} `maxminddb:"names"`
			} `maxminddb:"country"`
		}

		ipStr, ok := logData[geoField.Field].(string)
		if !ok {
			slog.Warnf("Field %s not found or not a string in log", geoField.Field)
			continue
		}

		slog.Debug(fmt.Sprintf("Looking up country for IP %s", ipStr))

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return logStr, errors.New("invalid IP address, ip: " + ipStr)
		}

		err := maxdb.Lookup(ip, &countryRecord)
		if err != nil {
			logging.ErrorLogger.Errorf("Error looking up country for IP %s: %v", ip, err)
			continue
		}

		// Append country code to the log with the specified tag
		logData[geoField.Tag] = countryRecord.Country.IsoCode
		slog.Debug(fmt.Sprintf("Country code for IP %s: %s", ipStr, countryRecord))
	}

	modifiedLog, err := json.Marshal(logData)
	if err != nil {
		return "", fmt.Errorf("error marshalling modified log: %v", err)
	}

	return string(modifiedLog), nil
}

func EnsureGeoIpInitialized(path string, logConfig config.Config) {
	manticore.ManticoreProtocol = logConfig.MantiProtocol
	manticore.ManticoreIP = logConfig.MantiIP
	manticore.ManticorePort = logConfig.MantiPort

	geoIPInit.Do(func() {
		var err error
		maxdb, err = maxminddb.Open(path)
		if err != nil {
			logging.ErrorLogger.Errorf("Failed to open GeoLite2-Country database: %v", err)
		}
	})
}
