package ghttp

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	Url                  string                   `json:"url"`
	Body                 interface{}              `json:"body"`
	Method               string                   `json:"method"`
	Headers              map[string]string        `json:"headers"`
	InsecureSkipVerify   bool                     `json:"insecureSkipVerify"`
	RenegotiationSupport tls.RenegotiationSupport `json:"renegotiationSupport"`
}

func (r *Request) Fetch() Response {
	uri := r.Url

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: r.InsecureSkipVerify,
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
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Exception fetching:", uri, err)
		return Response{Error: err}
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
	}
}

func (r *Request) getBody() (reqBody []byte, err error) {
	contentType, hasContentType := r.Headers["Content-Type"]

	if hasContentType && contentType == "application/x-www-form-urlencoded" {
		switch typedBody := r.Body.(type) {
		case map[string]string:
			formValues := url.Values{}
			for key, value := range typedBody {
				formValues.Set(key, value)
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

func (r *Request) FetchSSE(messageChan chan []byte) {
	defer close(messageChan)

	if r.Url == "" {
		logger.Error("URL is empty")
		return
	}

	reqBody := []byte{}
	var err error
	if r.Body != nil {
		reqBody, err = json.Marshal(r.Body)
		if err != nil {
			logger.Error("Error marshalling request body: ", err)
			return
		}
	}

	client := &http.Client{
		Timeout: 0,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: r.InsecureSkipVerify,
			},
		},
	}

	req, err := http.NewRequest(r.Method, r.Url, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("Error creating request: ", err)
		return
	}

	for key, value := range r.Headers {
		if len(value) > 0 {
			req.Header.Set(key, value)
		}
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error fetching: ", "URL: ", r.Url, " Error: ", err)
		return
	}

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			logger.Error("Error reading response: ", err)
			break
		}

		messageChan <- line
	}
}
