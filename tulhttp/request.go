package tulhttp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TulRequest struct {
	Url                string            `json:"url"`
	Body               interface{}       `json:"body"`
	Method             string            `json:"method"`
	Headers            map[string]string `json:"headers"`
	InsecureSkipVerify bool              `json:"insecureSkipVerify"`
}

func (r *TulRequest) Fetch() TulResponse {
	url := r.Url

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: r.InsecureSkipVerify,
		},
		IdleConnTimeout: 5 * time.Minute,
	}

	client := &http.Client{
		Transport: transport,
	}

	var reqBody []byte
	var err error
	if r.Body != nil {
		reqBody, err = json.Marshal(r.Body)
		if err != nil {
			return TulResponse{Error: err}
		}
	}

	req, err := http.NewRequest(r.Method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return TulResponse{Error: err}
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
		logger.Error("Exception fetching:", url, err)
		return TulResponse{Error: err}
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("Error closing response body:", err)
		}
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response:", err)
		return TulResponse{Error: err}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TulResponse{
			StatusCode: resp.StatusCode,
			Data:       bodyBytes,
			Error:      fmt.Errorf("bad status: %s", resp.Status),
		}
	}

	return TulResponse{
		StatusCode: resp.StatusCode,
		Data:       bodyBytes,
	}
}

func (r *TulRequest) FetchSSE(messageChan chan []byte) {
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
