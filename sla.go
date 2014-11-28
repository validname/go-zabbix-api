package zabbix

//import "github.com/AlekSi/reflector"

// https://www.zabbix.com/documentation/2.0/manual/appendix/api/history/definitions
type SLA struct {
	From      float64 `json:"from,omitempty"`
	To       float64    `json:"to"`
	Sla       float64 `json:"sla"`
	OkTime float64    `json:"okTime,omitempty"`
	ProblemTime float64    `json:"problemTime,omitempty"`
	DowntimeTime float64    `json:"downtimeTime,omitempty"`
	HostId	int64 `json:"hostid"`
}


// Wrapper for item.get https://www.zabbix.com/documentation/2.0/manual/appendix/api/item/get
func (api *API) SlaGet(params Params) (res *SLA, err error) {
	if _, present := params["output"]; !present {
		params["output"] = "extend"
	}
	response, err := api.CallWithError("service.getsla", params)
	if err != nil {
		return
	}

	ids:=params["serviceids"].(string)
	k:=response.Result.(map[string]interface{})
	k2:=k[ids].(map[string]interface{})["sla"]
	k3:=k2.([]interface{})
	if len(k3)==0 {
		return nil,nil
	}
	k4:=k3[0].(map[string]interface{})
	k5:=k4
	
	res=&SLA{}

	res.From=k5["from"].(float64)
	res.To=k5["to"].(float64)
	res.Sla=k5["sla"].(float64)
	res.OkTime=k5["okTime"].(float64)
	res.ProblemTime=k5["problemTime"].(float64)
	res.DowntimeTime=k5["downtimeTime"].(float64)

	//reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}
