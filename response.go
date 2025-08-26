package ghttp

import (
	"encoding/json"
	"net/http"

	"github.com/innovafour/logger"
)

type Response struct {
	StatusCode int    `json:"status"`
	Data       []byte `json:"data_byte"`
	Error      error  `json:"err"`
	Headers    http.Header
}

func (r *Response) DataAsMap() map[string]interface{} {
	var result map[string]interface{}

	err := json.Unmarshal(r.Data, &result)

	if err != nil {
		logger.Debug("Error unmarshalling data as map: %s", err.Error())
		return nil
	}

	return result
}

func (r *Response) DataAsSlice() []interface{} {
	var result []interface{}

	err := json.Unmarshal(r.Data, &result)

	if err != nil {
		logger.Debug("Error unmarshalling data as map: %s", err.Error())
		return nil
	}

	return result
}

func (r *Response) Is404() bool {
	return r.StatusCode == http.StatusNotFound
}

func (r *Response) IsOK() bool {
	if r.StatusCode == 0 {
		r.StatusCode = http.StatusInternalServerError
	}

	return (r.StatusCode >= 200 && r.StatusCode < 300) && r.Error == nil
}

func (r *Response) Unmarshal(data any) error {
	return json.Unmarshal(r.Data, data)
}
