package zabbix

import (
	"github.com/AlekSi/reflector"
)

// https://www.zabbix.com/documentation/2.0/manual/appendix/api/proxy/definitions
type Proxy struct {
	ProxyId string `json:"proxyid,omitempty"`
	Host    string `json:"host"`
	Error   string `json:"error"`
}

type Proxies []Proxy

// Wrapper for host.get: https://www.zabbix.com/documentation/2.0/manual/appendix/api/proxy/get
func (api *API) ProxyGet(params Params) (res Proxies, err error) {
	if _, present := params["output"]; !present {
		params["output"] = "extend"
	}
	response, err := api.CallWithError("proxy.get", params)
	if err != nil {
		return
	}

	reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}
