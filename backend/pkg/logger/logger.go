package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

// NewLogger initializes and returns a new logger instance with the specified log level and output settings.
func NewLogger(level, format, output string) *logrus.Logger {
	log := logrus.New()

	// Set log level based on the configuration
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set log format (JSON or text)
	switch format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true, // Show full timestamps in text format
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{}) // Default to JSON
	}

	// Set output destination (stdout or file)
	switch output {
	case "stdout":
		log.SetOutput(os.Stdout)
	default:
		// Attempt to open the file for logging
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.SetOutput(os.Stdout) // Fallback to stdout if file can't be opened
			log.Warn("Failed to log to file, using default stdout")
		} else {
			log.SetOutput(file)
		}
	}

	return log
}
