package zabbix

import "github.com/AlekSi/reflector"

// History object
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/history/definitions
type History struct {
	ItemId      string `json:"itemid,omitempty"`
	Clock       int    `json:"clock"`
	Value       string `json:"value"`
	Nanoseconds int    `json:"ns,omitempty"`
	Id          string `json:"id,omitempty"`
	LogEventId  int    `json:"logeventid,omitempty"`
	Severity    int    `json:"severity,omitempty"`
	Source      string `json:"source,omitempty"`
	Timestamp   int    `json:"timestamp,omitempty"`
}

// HistoriesGet is a wrapper for 'item.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/history/get
func (api *API) HistoriesGet(params Params) (res []History, err error) {
	if _, ok := params["output"]; !ok {
		params["output"] = "extend"
	}
	response, err := api.CallWithError("history.get", params)
	if err != nil {
		return
	}

	reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}
