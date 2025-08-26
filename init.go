package ghttp

import (
	"net/http"

	"github.com/innovafour/logger"
)

var headers http.Header

func Init(h http.Header, logLevel ...string) {
	headers = h

	if len(logLevel) > 0 {
		logger.New(logger.LevelFromString(logLevel[0]))
		return
	}

	clientUuid = GetKeyFromHeaders(headers, clientUuidHeadersKey)
}
