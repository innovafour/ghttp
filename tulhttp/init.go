package tulhttp

import "net/http"

var logger iLogger
var headers http.Header

func Init(h http.Header, l iLogger) {
	headers = h
	logger = l

	sessionLocal = TulSessionFromHeaders(headers)
	clientUuid = GetKeyFromHeaders(headers, clientUuidHeadersKey)
}
