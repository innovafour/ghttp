package tulhttp

import (
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

func (r TulRequest) Fetch() TulResponse {
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

	var data interface{}

	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			logger.Error("Error unmarshalling response:", string(bodyBytes), err, url)
			return TulResponse{Error: err}
		}
	} else {
		data = make(map[string]interface{})
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Error fetching: ", " URL: ", url, " STATUS CODE: ", resp.StatusCode, " DATA: ", data)
		return TulResponse{
			StatusCode: resp.StatusCode,
			Data:       data,
			Error:      fmt.Errorf("bad status: %s", resp.Status),
		}
	}

	return TulResponse{
		StatusCode: resp.StatusCode,
		Data:       data,
	}
}
