package ghttp

import "net/http"

var logger iLogger
var headers http.Header

func Init(h http.Header, l iLogger) {
	headers = h
	if l == nil {
		l = &defaultLogger{}
	}

	logger = l
}
