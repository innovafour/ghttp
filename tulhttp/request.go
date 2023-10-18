package tulhttp

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TulRequest struct {
	Method             string      `json:"method"`
	Body               interface{} `json:"body"`
	Url                string      `json:"url"`
	InsecureSkipVerify bool        `json:"insecureSkipVerify"`
}

func (r TulRequest) Fetch() TulResponse {
	url := r.Url

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: r.InsecureSkipVerify,
		},
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

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Exception fetching:", url, err)
		return TulResponse{Error: err}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response:", err)
		return TulResponse{Error: err}
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
		if err != nil {
			logger.Error("Error creating gzip reader:", err)
			return TulResponse{Error: err}
		}
		defer reader.Close()

		bodyBytes, err = io.ReadAll(reader)
		if err != nil {
			logger.Error("Error reading decompressed response:", err)
			return TulResponse{Error: err}
		}
	}

	var data interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		logger.Error("Error unmarshalling response:", string(bodyBytes), err, url)
		return TulResponse{Error: err}
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Error fetching:", url, data)
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
