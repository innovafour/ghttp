package tulhttp

import (
	"encoding/json"
	"strings"
)

const (
	sessionTypeHeadersKey               string = "x-session-type"
	sessionAddressUuidHeadersKey        string = "x-session-address-uuid"
	sessionAddressTypeHeadersKey        string = "x-session-address-type"
	sessionMarketingZonesHeadersKey     string = "x-session-marketing-zones"
	sessionSegmentsHeadersKey           string = "x-session-segments"
	sessionLogisticZonesHeadersKey      string = "x-session-logistic-zones"
	sessionZoneUuidHeadersKey           string = "x-session-zone-uuid"
	sessionWarehousesHeadersKey         string = "x-session-warehouses"
	sessionUuidHeadersKey               string = "x-session-uuid"
	sessionWarehouseUuidHeadersKey      string = "x-session-warehouse-uuid"
	sessionActivityStatusHeadersKey     string = "x-session-activity-status"
	sessionClientTypeHeadersKey         string = "x-session-client-type"
	sessionClusterExperimentsHeadersKey string = "x-session-cluster-experiments"
	sessionDeviceUuidHeadersKey         string = "x-session-device-uuid"
	sessionClientSubtypeHeadersKey      string = "x-session-client-subtype"
)

type TulSession struct {
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

func TulSessionFromHeaders(headers map[string][]string) TulSession {
	return TulSession{
		Uuid:               GetKeyFromHeaders(headers, sessionUuidHeadersKey),
		Type:               GetKeyFromHeaders(headers, sessionTypeHeadersKey),
		AddressUuid:        GetKeyFromHeaders(headers, sessionAddressUuidHeadersKey),
		AddressType:        GetKeyFromHeaders(headers, sessionAddressTypeHeadersKey),
		ZoneUuid:           GetKeyFromHeaders(headers, sessionZoneUuidHeadersKey),
		WarehouseUuid:      GetKeyFromHeaders(headers, sessionWarehouseUuidHeadersKey),
		ActivityStatus:     GetKeyFromHeaders(headers, sessionActivityStatusHeadersKey),
		ClientType:         GetKeyFromHeaders(headers, sessionClientTypeHeadersKey),
		ClusterExperiments: GetKeyFromHeaders(headers, sessionClusterExperimentsHeadersKey),
		DeviceUuid:         GetKeyFromHeaders(headers, sessionDeviceUuidHeadersKey),
		ClientSubtype:      GetKeyFromHeaders(headers, sessionClientSubtypeHeadersKey),
		MarketingZones:     strings.Split(GetKeyFromHeaders(headers, sessionMarketingZonesHeadersKey), ","),
		Segments:           strings.Split(GetKeyFromHeaders(headers, sessionSegmentsHeadersKey), ","),
		LogisticZones:      strings.Split(GetKeyFromHeaders(headers, sessionLogisticZonesHeadersKey), ","),
		Warehouses:         strings.Split(GetKeyFromHeaders(headers, sessionWarehousesHeadersKey), ","),
	}
}

func (s TulSession) HasToValidateAvailability() bool {
	return s.AddressType == "warehouse"
}
