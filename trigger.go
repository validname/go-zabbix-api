package zabbix

import "encoding/json"

type (
	TriggerFlags      int
	TriggerPriority   int
	TriggerStatus     int
	TriggerType       int
	TriggerValue      int
	TriggerValueFlags int
)

const (
	PlainTrigger      TriggerFlags = 0
	DiscoveredTrigger TriggerFlags = 4

	NoClssified TriggerPriority = 0
	Information TriggerPriority = 1
	Warning     TriggerPriority = 2
	Average     TriggerPriority = 3
	High        TriggerPriority = 4
	Disaster    TriggerPriority = 5

	Enabled  TriggerStatus = 0
	Disabled TriggerStatus = 1

	Nomultiple            TriggerType = 0
	GenerateMulipleEvents TriggerType = 1

	Ok      TriggerValue = 0
	Problem TriggerValue = 1

	UpToDate TriggerValueFlags = 0
	Unknown  TriggerValueFlags = 1
)

type ResponseTrigger struct {
	Jsonrpc string    `json:"jsonrpc"`
	Error   *Error    `json:"error"`
	Result  []Trigger `json:"result"`
	Id      int32     `json:"id"`
}

type Function struct {
	FunctionId int64  `json:"functionid,string"`
	ItemId     int64  `json:"itemid,string"`
	Function   string `json:"function"`
	Parameter  string `json:"parameter"`
}

type Trigger struct {
	TriggerId   int64             `json:"triggerid,string"`
	Description string            `json:"description"`
	Functions   []Function        `json:"functions"`
	Expression  string            `json:"expression"`
	Comments    string            `json:"comments"`
	Error       string            `json:"error"`
	Flags       TriggerFlags      `json:"flags,string"`
	LastChange  int64             `json:"lastchange,string"`
	Priority    TriggerPriority   `json:"priority,string"`
	Status      TriggerStatus     `json:"status,string"`
	TemplateId  string            `json:"templateid"`
	Type        TriggerType       `json:"type,string"`
	Url         string            `json:"url"`
	Value       TriggerValue      `json:"value,string"`
	ValueFlags  TriggerValueFlags `json:"value_flags,string"`
}

//wraper para get.trigger
func (api *API) GetTrigger(params Params) ([]Trigger, error) {
	if _, present := params["output"]; !present {
		params["output"] = "extend"
	}
	if _, present2 := params["expandExpression"]; !present2 {
		params["expandExpression"] = "extend"
	}
	if _, present3 := params["expandDescription"]; !present3 {
		params["expandDescription"] = "flag"
	}
	if _, present4 := params["selectFunctions"]; !present4 {
		params["selectFunctions"] = "extend"
	}
	response, err := api.callBytes("trigger.get", params)
	if err != nil {
		return nil, err
	}
	r := ResponseTrigger{}
	err = json.Unmarshal(response, &r)

	return r.Result, err

}
