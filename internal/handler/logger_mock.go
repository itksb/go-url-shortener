// Package handler creates server routines for the routes
package handler

// loggerMock mock the logger for testing purposes
type loggerMock struct {
}

// Info log with Info level
func (l *loggerMock) Info(message string, args ...interface{}) {
}

// Error log with error level
func (l *loggerMock) Error(message interface{}, args ...interface{}) {
}
