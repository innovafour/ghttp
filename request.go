package ghttp

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/innovafour/logger"
)

type Request struct {
	Url     string              `json:"url"`
	Body    interface{}         `json:"body"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	// InsecureSkipVerify turns certificate verification off when true.
	// Unset verifies (before v1.3.0 unset skipped verification).
	InsecureSkipVerify   *bool                    `json:"insecureSkipVerify"`
	RenegotiationSupport tls.RenegotiationSupport `json:"renegotiationSupport"`
	MaxRetries           int                      `json:"maxRetries"`
	RetryDelay           time.Duration            `json:"retryDelay"`
	// Timeout bounds each attempt, body included. Zero is DefaultTimeout.
	Timeout time.Duration `json:"timeout"`
}

func (r *Request) Fetch() Response {
	reqBody, err := r.getBody()
	if err != nil {
		return Response{Error: err}
	}

	if r.MaxRetries == 0 {
		r.MaxRetries = 3
	}

	if r.RetryDelay == 0 {
		r.RetryDelay = 1 * time.Second
	}

	client := &http.Client{Transport: transportFor(r.insecure(), r.RenegotiationSupport)}

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		response, retry := r.fetchOnce(client, reqBody)
		if !retry || attempt == r.MaxRetries {
			return response
		}

		fmt.Printf("Request failed with status %d, retrying... (attempt %d of %d)\n", response.StatusCode, attempt+1, r.MaxRetries)
		time.Sleep(r.RetryDelay)
	}

	return Response{
		Error: fmt.Errorf("request failed after %d retries", r.MaxRetries),
	}
}

// fetchOnce sends one attempt and reads its whole body within Timeout. retry
// is true for a 5xx answer; the body is always read and closed, so the
// connection goes back to the shared pool.
func (r *Request) fetchOnce(client *http.Client, reqBody []byte) (response Response, retry bool) {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, r.Url, bytes.NewBuffer(reqBody))
	if err != nil {
		return Response{Error: err}, false
	}

	for key, value := range headers {
		if len(value) > 0 {
			req.Header.Set(key, value[0])
		}
	}

	for key, value := range r.Headers {
		if len(value) > 0 {
			req.Header.Set(key, value[0])
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Exception fetching:", r.Url, err)
		return Response{Error: err}, false
	}

	defer resp.Body.Close()

	var bodyBytes []byte

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return Response{Error: err}, false
		}

		defer func() {
			err := reader.Close()
			if err != nil {
				logger.Error("Error closing gzip reader:", err)
			}
		}()

		bodyBytes, err = io.ReadAll(reader)
		if err != nil {
			logger.Error("Error reading gzip body: ", err)
			return Response{Error: err}, false
		}
	} else {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Error reading body: ", err)
			return Response{Error: err}, false
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{
			StatusCode: resp.StatusCode,
			Data:       bodyBytes,
			Error:      fmt.Errorf("bad status: %s", resp.Status),
		}, resp.StatusCode >= 500 && resp.StatusCode < 600
	}

	return Response{
		StatusCode: resp.StatusCode,
		Data:       bodyBytes,
		Headers:    resp.Header,
	}, false
}

func (r *Request) getBody() (reqBody []byte, err error) {
	contentType := GetKeyFromHeaders(r.Headers, "Content-Type")

	if contentType == "application/x-www-form-urlencoded" {
		switch typedBody := r.Body.(type) {
		case map[string]string:
			formValues := url.Values{}
			for key, value := range typedBody {
				formValues.Set(key, value)
			}
			reqBody = []byte(formValues.Encode())
		case map[string]interface{}:
			formValues := url.Values{}
			for key, value := range typedBody {
				valueString := fmt.Sprintf("%v", value)
				formValues.Set(key, valueString)
			}
			reqBody = []byte(formValues.Encode())
		default:
			return nil, fmt.Errorf("invalid body type for Content-Type application/x-www-form-urlencoded")
		}
	} else {
		switch typedBody := r.Body.(type) {
		case string:
			reqBody = []byte(typedBody)
		case []byte:
			reqBody = typedBody
		case io.ReadCloser:
			body, err := io.ReadAll(typedBody)
			if err != nil {
				return nil, err
			}
			reqBody = body
		default:
			if r.Body != nil {
				reqBody, err = json.Marshal(r.Body)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return reqBody, nil
}
