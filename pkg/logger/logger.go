package logger

import (
	"log"
	"os"
)

// Interface -.
type Interface interface {
	Info(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
}

// Logger -.
type Logger struct {
	logger *log.Logger
}

// NewLogger - Constructor
func NewLogger() (*Logger, error) {
	var l = log.New(os.Stderr, "", log.LstdFlags)
	return &Logger{
		logger: l,
	}, nil
}

// Info -.
func (l *Logger) Info(message string, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Println("Info: ", message, args)
		return
	}
	l.logger.Println("Info: ", message)
}

// Error -.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	if len(args) > 0 {
		l.logger.Println("Error: ", message, args)
		return
	}
	l.logger.Println("Error: ", message)
}
