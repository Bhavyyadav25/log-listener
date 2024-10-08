// pkg/logging/logging.go
package logging

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogType int

const (
	AuditLog LogType = iota
	ErrorLog
	InfoLog
)

var (
	AuditLogger *zap.SugaredLogger
	ErrorLogger *zap.SugaredLogger
	InfoLogger  *zap.SugaredLogger
)

func InitLogger(logFilePath map[LogType]string) {
	AuditLogger = createLogger(logFilePath[AuditLog], zapcore.InfoLevel)
	ErrorLogger = createLogger(logFilePath[ErrorLog], zapcore.ErrorLevel)
	InfoLogger = createLogger(logFilePath[InfoLog], zapcore.InfoLevel)
}

func createLogger(logFilePath string, level zapcore.Level) *zap.SugaredLogger {
	rotator := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(rotator),
		level,
	)

	logger := zap.New(core, zap.AddCaller())
	return logger.Sugar()
}

func SyncLogger() {
	if AuditLogger != nil {
		_ = AuditLogger.Sync()
	}
	if ErrorLogger != nil {
		_ = ErrorLogger.Sync()
	}
	if InfoLogger != nil {
		_ = InfoLogger.Sync()
	}
}
