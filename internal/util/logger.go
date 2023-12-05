package util

import "github.com/sirupsen/logrus"

const LogLevelInfo = "info"
const LogLevelDebug = "debug"
const LogLevelTrace = "trace"

func NewLogrusEntry(fields logrus.Fields, logger *logrus.Logger) *logrus.Entry {
	l := logrus.New()
	l.Formatter = logger.Formatter
	loggerEntry := logger.WithFields(fields)
	loggerEntry.Level = logger.Level

	return loggerEntry
}

func StringToLogrusLogLevel(log_level string) logrus.Level {
	switch log_level {
		case LogLevelDebug:
			return logrus.DebugLevel
		case LogLevelTrace:
			return logrus.TraceLevel
		default:
			return logrus.InfoLevel
	}
}