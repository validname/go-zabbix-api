package zabbix

import "encoding/json"

type (
	SourceType int
	ObjectType int
)

const (
	FromTrigger  SourceType = 0
	FromDiscover SourceType = 1
	FromAutoReg  SourceType = 2

	ObjTrigger           ObjectType = 0
	ObjDiscoveredHost    ObjectType = 1
	ObjDiscoveredService ObjectType = 2
	ObjAutoRegHost       ObjectType = 3
)

type ResponseEvent struct {
	Jsonrpc string  `json:"jsonrpc"`
	Error   *Error  `json:"error"`
	Result  []Event `json:"result"`
	Id      int32   `json:"id"`
}

type AckType struct {
	AckId   int64  `json:"acknowledgeid,string"`
	Userid  int64  `json:"userid,string"`
	EventId int64  `json:"eventid,string"`
	Clock   int64  `json:"clock,string"`
	Message string `json:"message"`
	Alias   string `json:"alias"`
}

type Event struct {
	EventId      int64      `json:"eventid,string"`
	Source       SourceType `json:"source,string"`
	Object       ObjectType `json:"object,string"`
	ObjectId     int64      `json:"objectid,string"`
	Clock        int64      `json:"clock,string"`
	Value        int64      `json:"value,string"`
	AckNowLedge  AckType    `json:"acknowledges"`
	Ns           int64      `json:"ns,string"`
	ValueChanged int64      `json:"value_changed,string"`
}

//wraper para get.trigger
func (api *API) GetEvents(params Params) ([]Event, error) {
	if _, present := params["output"]; !present {
		params["output"] = "extend"
	}
	if _, present2 := params["select_acknowledges"]; !present2 {
		params["select_acknowledges"] = "extend"
	}

	response, err := api.callBytes("event.get", params)
	if err != nil {
		return nil, err
	}
	r := ResponseEvent{}
	err = json.Unmarshal(response, &r)

	return r.Result, err

}
