package tulhttp

type iLogger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Error(v ...interface{})
}
