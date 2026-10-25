package ghttp

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func init() {
	Init(nil, "ERROR")
}

// countingServer counts the TCP connections clients open to it.
func countingServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *int32) {
	t.Helper()

	var opened int32
	server := httptest.NewUnstartedServer(handler)
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddInt32(&opened, 1)
		}
	}
	server.Start()
	t.Cleanup(server.Close)

	return server, &opened
}

// Scenario: twenty sequential requests to one host.
// Expected: they share one connection (v1.2.0 opened one per request).
func TestFetchReusesConnections(t *testing.T) {
	server, opened := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	for i := 0; i < 20; i++ {
		req := Request{Method: http.MethodGet, Url: server.URL}
		if resp := req.Fetch(); !resp.IsOK() {
			t.Fatalf("request %d: %v", i, resp.Error)
		}
	}

	if got := atomic.LoadInt32(opened); got != 1 {
		t.Fatalf("want 1 connection, got %d", got)
	}
}

// Scenario: a 5xx answered twice, then a 200.
// Expected: retried to success over the same connection, so the failed
// bodies were read and closed.
func TestFetchRetriesOn5xxWithoutLeakingConnections(t *testing.T) {
	var calls int32
	server, opened := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) <= 2 {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})

	req := Request{Method: http.MethodPost, Url: server.URL, Body: map[string]string{"a": "b"}, RetryDelay: time.Millisecond}
	if resp := req.Fetch(); !resp.IsOK() {
		t.Fatalf("want success after retries, got %d %v", resp.StatusCode, resp.Error)
	}

	if atomic.LoadInt32(&calls) != 3 || atomic.LoadInt32(opened) != 1 {
		t.Fatalf("want 3 calls on 1 connection, got %d calls on %d", calls, *opened)
	}
}

// Scenario: retries exhausted on a 5xx.
// Expected: the last status and an error come back.
func TestFetchReturnsLastStatusAfterRetries(t *testing.T) {
	server, _ := countingServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})

	req := Request{Method: http.MethodGet, Url: server.URL, MaxRetries: 1, RetryDelay: time.Millisecond}
	resp := req.Fetch()
	if resp.StatusCode != http.StatusBadGateway || resp.Error == nil {
		t.Fatalf("want 502 with error, got %d %v", resp.StatusCode, resp.Error)
	}
}

// Scenario: a server that never answers.
// Expected: the attempt gives up at Timeout instead of hanging forever.
func TestFetchTimesOut(t *testing.T) {
	release := make(chan struct{})
	server, _ := countingServer(t, func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})
	defer close(release)

	start := time.Now()
	req := Request{Method: http.MethodGet, Url: server.URL, Timeout: 100 * time.Millisecond}
	if resp := req.Fetch(); resp.Error == nil {
		t.Fatal("want a timeout error")
	}

	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("took %s", elapsed)
	}
}

// Scenario: a self-signed certificate, with and without the opt-out.
// Expected: rejected by default, accepted only with InsecureSkipVerify true.
func TestFetchVerifiesTLSByDefault(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	server.Config.ErrorLog = nil

	req := Request{Method: http.MethodGet, Url: server.URL}
	if resp := req.Fetch(); resp.Error == nil {
		t.Fatal("a self-signed certificate must be rejected by default")
	}

	insecure := true
	req = Request{Method: http.MethodGet, Url: server.URL, InsecureSkipVerify: &insecure}
	if resp := req.Fetch(); !resp.IsOK() {
		t.Fatalf("explicit opt-out must connect: %v", resp.Error)
	}
}
