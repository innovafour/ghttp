package ghttp

type iLogger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Error(v ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Debug(v ...interface{}) {
}

func (l *defaultLogger) Info(v ...interface{}) {
}

func (l *defaultLogger) Error(v ...interface{}) {
}
