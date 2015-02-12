package zabbix

import (
	"encoding/json"
)

type (
	AvailableType int
	StatusType    int
)

const (
	Available   AvailableType = 1
	Unavailable AvailableType = 2

	Monitored   StatusType = 0
	Unmonitored StatusType = 1
)

// Host object
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/host/definitions
type Host struct {
	HostId    string        `json:"hostid,omitempty"`
	Host      string        `json:"host"`
	Available AvailableType `json:"available,string"`
	Error     string        `json:"error"`
	Name      string        `json:"name"`
	Status    StatusType    `json:"status,string"`
	ProxyId   string        `json:"proxy_hostid,omitempty"`

	// Fields below used only when creating hosts
	//GroupIds   HostGroupIds   `json:"groups,omitempty"`
	GroupIds   HostGroupIds   `json:"groups"`
	Interfaces HostInterfaces `json:"interfaces,omitempty"`

	// Only for backward compatibility with Zabbix 1.8
	DNS         string        `json:"dns,omitempty"`
	IP          string        `json:"ip,omitempty"`
	UseIP       int           `json:"useip,string,omitempty"`
	Port        string        `json:"port,omitempty"`
}

type Hosts []Host

type HostId struct {
	HostId string `json:"hostid"`
}

type HostIds []HostId

// HostsGet is a wrapper for 'host.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/host/get
func (api *API) HostsGet(params Params) (result Hosts, err error) {
	defaults := Params{
		"output":           "extend",
		"selectGroups":     "extend",
		"selectInterfaces": "extend",
	}

	for key, defaultValue := range defaults {
		if _, ok := params[key]; !ok {
			params[key] = defaultValue
		}
	}

	if !api.IsVersionBigger(2, 0, 0) {
		// Transform parameters for Zabbix 1.8
		if _, ok := params["selectInterfaces"]; ok {
			delete(params, "selectInterfaces")
		}
		if value, ok := params["selectGroups"]; ok {
			params["select_groups"] = value
			delete(params, "selectGroups")
		}
	}

	var response ResponseWithJson
	b, err := api.callBytes("host.get", params)
	if err == nil {
		err = json.Unmarshal(b, &response)
	}
	if err == nil && response.Error != nil {
		err = response.Error
	}
	if err != nil {
		return
	}

	result = make(Hosts, 0)
	err = json.Unmarshal(response.Result, &result)
	return
}

// HostsGetByHostGroupIds gets hosts by host group Ids.
func (api *API) HostsGetByHostGroupIds(ids []string) (res Hosts, err error) {
	return api.HostsGet(Params{"groupids": ids})
}

// HostsGetByHostGroups gets hosts by host groups.
func (api *API) HostsGetByHostGroups(hostGroups HostGroups) (res Hosts, err error) {
	ids := make([]string, len(hostGroups))
	for i, id := range hostGroups {
		ids[i] = id.GroupId
	}
	return api.HostsGetByHostGroupIds(ids)
}

// HostsGetByTemplateIds gets hosts by linked template Ids.
func (api *API) HostsGetByTemplateIds(ids []string) (res Hosts, err error) {
	return api.HostsGet(Params{"templateids": ids})
}

// HostGetById gets host by Id only if there is exactly 1 matching host.
func (api *API) HostGetById(id string) (res *Host, err error) {
	hosts, err := api.HostsGet(Params{"hostids": id})
	if err != nil {
		return
	}

	if len(hosts) == 1 {
		res = &hosts[0]
	} else {
		e := ExpectedOneResult(len(hosts))
		err = &e
	}
	return
}

// HostGetByHost gets host by Host only if there is exactly 1 matching host.
func (api *API) HostGetByHost(host string) (res *Host, err error) {
	hosts, err := api.HostsGet(Params{"filter": map[string]string{"host": host}})
	if err != nil {
		return
	}

	if len(hosts) == 1 {
		res = &hosts[0]
	} else {
		e := ExpectedOneResult(len(hosts))
		err = &e
	}
	return
}

// HostsCreate is a wrapper for 'host.create'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/host/create
func (api *API) HostsCreate(hosts Hosts) (err error) {
	response, err := api.CallWithError("host.create", hosts)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	hostids := result["hostids"].([]interface{})
	for i, id := range hostids {
		hosts[i].HostId = id.(string)
	}
	return
}

// HostsDelete is a wrapper for 'host.delete'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/host/delete
// Deletes by hosts list
// Cleans HostId in all hosts elements if call succeed.
func (api *API) HostsDelete(hosts Hosts) (err error) {
	ids := make([]string, len(hosts))
	for i, host := range hosts {
		ids[i] = host.HostId
	}

	if api.IsVersionBigger(2, 4, 0) {
		err = api.HostsDeleteByIdsNew(ids)
	} else {
		err = api.HostsDeleteByIds(ids)
	}
	if err == nil {
		for i := range hosts {
			hosts[i].HostId = ""
		}
	}
	return
}

// HostsDeleteByIds is a wrapper for 'host.delete'
// see https://www.zabbix.com/documentation/1.8/manual/appendix/api/host/delete
// Deletes by hosts Id list
// Only for Zabbix version up to 1.8
func (api *API) HostsDeleteByIds(ids []string) (err error) {
	hostIds := make([]map[string]string, len(ids))
	for i, id := range ids {
		hostIds[i] = map[string]string{"hostid": id}
	}

	response, err := api.CallWithError("host.delete", hostIds)

	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	hostids := result["hostids"].([]interface{})
	if len(ids) != len(hostids) {
		err = &ExpectedMore{len(ids), len(hostids)}
	}
	return
}

// HostsDeleteByIdsNew is a wrapper for 'host.delete'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/host/delete
// Deletes by hosts Id list
// Only for Zabbix version more than 1.8
func (api *API) HostsDeleteByIdsNew(ids []string) (err error) {
	response, err := api.CallWithError("host.delete", ids)

	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	hostids := result["hostids"].([]interface{})
	if len(ids) != len(hostids) {
		err = &ExpectedMore{len(ids), len(hostids)}
	}
	return
}
