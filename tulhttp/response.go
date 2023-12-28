package tulhttp

import (
	"encoding/json"
	"net/http"
)

type TulResponse struct {
	StatusCode int    `json:"status"`
	Data       []byte `json:"data_byte"`
	Error      error  `json:"err"`
}

func (r TulResponse) DataAsMap() map[string]interface{} {
	var result map[string]interface{}

	err := json.Unmarshal(r.Data, &result)

	if err != nil {
		logger.Debug("Error unmarshalling data as map: %s", err.Error())
		return nil
	}

	return result
}

func (r TulResponse) DataAsSlice() []interface{} {
	var result []interface{}

	err := json.Unmarshal(r.Data, &result)

	if err != nil {
		logger.Debug("Error unmarshalling data as map: %s", err.Error())
		return nil
	}

	return result
}

func (r TulResponse) Is404() bool {
	return r.StatusCode == http.StatusNotFound
}

func (r TulResponse) IsOK() bool {
	if r.StatusCode == 0 {
		r.StatusCode = http.StatusInternalServerError
	}

	return (r.StatusCode >= 200 && r.StatusCode < 300) && r.Error == nil
}
