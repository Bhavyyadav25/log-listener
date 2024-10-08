package parser

import (
	"regexp"
)

type LogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Client    string `json:"client"`
	ID        string `json:"id"`
	Msg       string `json:"msg"`
	Tag       string `json:"tag"`
	Hostname  string `json:"hostname"`
	Request   string `json:"request"`
	Host      string `json:"host"`
}

// ParseErrorLog takes a raw log entry string and extracts the fields
// into a LogEntry struct. It uses regular expressions to identify and
// capture specific parts of the log entry.
//
// The log parameter is a string representing the raw log entry.
// The function returns a LogEntry struct containing the parsed fields
// and an error if any parsing issue occurs.
func ParseErrorLog(log string) (LogEntry, error) {
	var logEntry LogEntry

	clientPattern := regexp.MustCompile(`\[client\s(?P<client>[^\]]+)\]`)
	idPattern := regexp.MustCompile(`\[id\s"(?P<id>[^"]+)"\]`)
	msgPattern := regexp.MustCompile(`\[msg\s"(?P<msg>[^"]+)"\]`)
	tagPattern := regexp.MustCompile(`\[tag\s"(?P<tag>[^"]+)"\]`)
	hostnamePattern := regexp.MustCompile(`server:\s(?P<server>[^,]+),`)
	requestPattern := regexp.MustCompile(`request:\s"(?P<request>[^"]+)"`)
	hostPattern := regexp.MustCompile(`host:\s"(?P<host>[^"]+)"`)

	if match := clientPattern.FindStringSubmatch(log); match != nil {
		logEntry.Client = match[1]
	}
	if match := idPattern.FindStringSubmatch(log); match != nil {
		logEntry.ID = match[1]
	}
	if match := msgPattern.FindStringSubmatch(log); match != nil {
		logEntry.Msg = match[1]
	}
	if match := tagPattern.FindStringSubmatch(log); match != nil {
		logEntry.Tag = match[1]
	}
	if match := hostnamePattern.FindStringSubmatch(log); match != nil {
		logEntry.Hostname = match[1]
	}
	if match := requestPattern.FindStringSubmatch(log); match != nil {
		logEntry.Request = match[1]
	}
	if match := hostPattern.FindStringSubmatch(log); match != nil {
		logEntry.Host = match[1]
	}

	return logEntry, nil
}

//TODO : Handle recover
