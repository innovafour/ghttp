package tulhttp

import (
	"encoding/json"
	"fmt"
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

func (r *TulResponse) IsOK() bool {
	if r.StatusCode == 0 {
		r.StatusCode = http.StatusInternalServerError
	}

	jsonData, err := json.Marshal(r.Data)
	if err != nil {
		fmt.Println("Error al serializar el mapa:", err)
		return false
	}

	r.DataByte = jsonData

	return (r.StatusCode >= 200 && r.StatusCode < 300) && r.Error == nil
}
