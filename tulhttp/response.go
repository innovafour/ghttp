package tulhttp

import (
	"net/http"
)

type TulResponse struct {
	StatusCode int         `json:"status"`
	Data       interface{} `json:"data"`
	DataByte   []byte      `json:"data_byte"`
	Error      error       `json:"err"`
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
