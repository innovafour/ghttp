package ghttp

import (
	"bytes"
	"compress/gzip"
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
	Url                  string                   `json:"url"`
	Body                 interface{}              `json:"body"`
	Method               string                   `json:"method"`
	Headers              map[string][]string      `json:"headers"`
	InsecureSkipVerify   *bool                    `json:"insecureSkipVerify"`
	RenegotiationSupport tls.RenegotiationSupport `json:"renegotiationSupport"`
	MaxRetries           int                      `json:"maxRetries"`
	RetryDelay           time.Duration            `json:"retryDelay"`
}

func (r *Request) Fetch() Response {
	uri := r.Url

	var insecure bool

	if r.InsecureSkipVerify != nil {
		insecure = *r.InsecureSkipVerify
	} else {
		insecure = true
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
			Renegotiation:      r.RenegotiationSupport,
		},
		IdleConnTimeout: 5 * time.Minute,
	}

	client := &http.Client{
		Transport: transport,
	}

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

	var resp *http.Response

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		req, err := http.NewRequest(r.Method, uri, bytes.NewBuffer(reqBody))
		if err != nil {
			return Response{Error: err}
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

		resp, err = client.Do(req)
		if err != nil {
			logger.Error("Exception fetching:", uri, err)
			return Response{Error: err}
		}

		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			if attempt < r.MaxRetries {
				fmt.Printf("Request failed with status %d, retrying... (attempt %d of %d)\n", resp.StatusCode, attempt+1, r.MaxRetries)
				time.Sleep(r.RetryDelay)
				continue
			}
		}

		defer resp.Body.Close()

		var bodyBytes []byte

		if resp.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return Response{Error: err}
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
				return Response{Error: err}
			}
		} else {
			bodyBytes, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Error reading body: ", err)
				return Response{Error: err}
			}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return Response{
				StatusCode: resp.StatusCode,
				Data:       bodyBytes,
				Error:      fmt.Errorf("bad status: %s", resp.Status),
			}
		}

		return Response{
			StatusCode: resp.StatusCode,
			Data:       bodyBytes,
			Headers:    resp.Header,
		}
	}

	return Response{
		Error: fmt.Errorf("request failed after %d retries", r.MaxRetries),
	}
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
