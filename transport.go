package ghttp

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

// DefaultTimeout bounds one attempt of Fetch (connect, send, read the whole
// body) when Request.Timeout is zero.
const DefaultTimeout = 30 * time.Second

type transportKey struct {
	insecure      bool
	renegotiation tls.RenegotiationSupport
}

// transports holds one *http.Transport per TLS setting, shared by every
// request. A transport per request never reused a connection, and each one
// left its idle connection (and two goroutines) open until IdleConnTimeout.
var transports sync.Map

func transportFor(insecure bool, renegotiation tls.RenegotiationSupport) *http.Transport {
	key := transportKey{insecure: insecure, renegotiation: renegotiation}
	if transport, ok := transports.Load(key); ok {
		return transport.(*http.Transport)
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
			Renegotiation:      renegotiation,
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   16,
		IdleConnTimeout:       90 * time.Second,
	}

	actual, _ := transports.LoadOrStore(key, transport)

	return actual.(*http.Transport)
}

// insecure reports whether certificate verification is off: only when the
// caller sets InsecureSkipVerify to true.
func (r *Request) insecure() bool {
	return r.InsecureSkipVerify != nil && *r.InsecureSkipVerify
}
