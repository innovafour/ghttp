package tulhttp

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"unicode"
)

type TulSession struct {
	Country            string   `json:"country"`
	Uuid               string   `json:"uuid"`
	Warehouses         []string `json:"warehouses"`
	Type               string   `json:"type"`
	AddressUuid        string   `json:"address_uuid"`
	AddressType        string   `json:"address_type"`
	MarketingZones     []string `json:"marketing_zones"`
	Segments           []string `json:"segments"`
	LogisticZones      []string `json:"logistic_zones"`
	ZoneUuid           string   `json:"zone_uuid"`
	WarehouseUuid      string   `json:"warehouse_uuid"`
	ActivityStatus     string   `json:"activity_status"`
	ClientType         string   `json:"client_type"`
	ClusterExperiments string   `json:"cluster_experiments"`
	DeviceUuid         string   `json:"device_uuid"`
	ClientSubtype      string   `json:"client_subtype"`
}

func (s TulSession) ToMap() (map[string]interface{}, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var resultMap map[string]interface{}
	err = json.Unmarshal(jsonData, &resultMap)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func (s TulSession) HasToValidateAvailability() bool {
	return s.AddressType == "warehouse"
}

func TulSessionFromHeaders(headers http.Header) TulSession {
	var session TulSession
	val := reflect.ValueOf(&session).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "-" || jsonTag == "" {
			continue
		}

		headerKey := "x-session-" + toSnakeCase(jsonTag)
		headerValue := GetKeyFromHeaders(headers, headerKey)

		if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String {
			sliceVal := strings.Split(headerValue, ",")
			field.Set(reflect.ValueOf(sliceVal))
		} else {
			field.SetString(headerValue)
		}
	}

	session.Country = GetKeyFromHeaders(headers, "x-country")

	return session
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}

	finalResult := strings.ReplaceAll(string(result), "_", "-")

	return finalResult
}
