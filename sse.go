package ghttp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"

	"github.com/innovafour/logger"
)

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

	var insecure bool

	if r.InsecureSkipVerify != nil {
		insecure = *r.InsecureSkipVerify
	} else {
		insecure = true
	}

	client := &http.Client{
		Timeout: 0,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
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
			req.Header.Set(key, value[0])
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
