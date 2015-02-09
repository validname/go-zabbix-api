package zabbix

import (
	"encoding/json"
	"errors"
)

type (
	InterfaceType int
)

const (
	Agent InterfaceType = 1
	SNMP  InterfaceType = 2
	IPMI  InterfaceType = 3
	JMX   InterfaceType = 4
)

// HostInterface object
// see https://www.zabbix.com/documentation/2.4/manual/api/reference/hostinterface/object
type HostInterface struct {
	InterfaceId string        `json:"interfaceid,omitempty"`
	DNS         string        `json:"dns"`
	IP          string        `json:"ip"`
	Main        int           `json:"main,string"`
	Port        string        `json:"port"`
	Type        InterfaceType `json:"type,string"`
	UseIP       int           `json:"useip,string"`
	UseBulkSNMP int           `json:"bulk,string,omitempty"`
}

type HostInterfaces []HostInterface

// HostInterfacesGet is a wrapper for 'hostinterface.get'
// see https://www.zabbix.com/documentation/2.4/manual/api/reference/hostinterface/get

func (api *API) HostInterfacesGet(params Params) (result HostInterfaces, err error) {
	if !api.isVersionBigger(2, 0, 0) {
		// there was no such object in Zabbix 1.8
		err = errors.New("Too old Zabbix version to use this function.")
		return
	}

	defaults := Params{
		"output": "extend",
	}
	for key, defaultValue := range defaults {
		if _, ok := params[key]; !ok {
			params[key] = defaultValue
		}
	}

	var response ResponseWithJson
	b, err := api.callBytes("hostinterface.get", params)
	if err == nil {
		err = json.Unmarshal(b, &response)
	}
	if err == nil && response.Error != nil {
		err = response.Error
	}
	if err != nil {
		return
	}

	result = make(HostInterfaces, 0)
	err = json.Unmarshal(response.Result, &result)
	return
}

// HostInterfacesGetByHostId gets host interfaces by host Id
// Use nil for empty additional parameters or filter
func (api *API) HostInterfacesGetByHostId(id string, params Params, filter map[string]string) (result HostInterfaces, err error) {
	if params == nil {
		params = make(Params)
	}
	params["hostids"] = []string{id}

	if filter != nil {
		params["filter"] = filter
	}
	return api.HostInterfacesGet(params)
}
