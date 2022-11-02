package handler

type loggerMock struct {
}

func (l *loggerMock) Info(message string, args ...interface{}) {
}
func (l *loggerMock) Error(message interface{}, args ...interface{}) {
}
