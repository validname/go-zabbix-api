package zabbix

import "github.com/AlekSi/reflector"

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

type Function struct {
	FunctionId int64  `json:"functionid,string"`
	ItemId     int64  `json:"itemid,string"`
	Function   string `json:"function"`
	Parameter  string `json:"parameter"`
}

// https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/definitions
type Trigger struct {
	TriggerId   string            `json:"triggerid,string"`
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
	// only when expandData flag is set
	HostId string `json:"hostid,string"`
	Host   string `json:"host,string"`
}

type Triggers []Trigger

// Wrapper for trigger.get: https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/get
func (api *API) TriggersGet(params Params) (res Triggers, err error) {
	if _, present := params["output"]; !present {
		params["output"] = "extend"
	}
	if _, present := params["expandExpression"]; !present {
		params["expandExpression"] = "extend"
	}
	if _, present := params["expandDescription"]; !present {
		params["expandDescription"] = "flag"
	}
	if _, present := params["expandData"]; !present {
		params["expandData"] = "extend"
	}
	if _, present := params["selectFunctions"]; !present {
		params["selectFunctions"] = "extend"
	}

	response, err := api.CallWithError("trigger.get", params)
	if err != nil {
		return
	}
	reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}

// Get trigger extended information by Id only if there is exactly 1 matching trigger
func (api *API) TriggerGetById(id string) (res *Trigger, err error) {
	params := make( map [string]interface{} )
	params["output"] = "extend"
	params["expandExpression"] = "extend"
	params["expandDescription"] = "flag"
	params["expandData"] = "extend"
	params["selectFunctions"] = "extend"

	triggers, err := api.TriggersGet(params)
	if err != nil {
		return
	}
	if len(triggers) == 1 {
		res = &triggers[0]
	} else {
		e := ExpectedOneResult(len(triggers))
		err = &e
	}
	return
}

// Return triggers on hosts which was inherited from template trigger
// 
func (api *API) TriggersGetInheritedFromId(id string, optional_filters ...map[string]string) (res Triggers, err error) {
	params := make( map [string]interface{} )
	params["output"] = "extend"
	params["expandExpression"] = "extend"
	params["expandDescription"] = "flag"
	params["expandData"] = "extend"
	params["inherited"] = 1

	filter := make( map [string]string )
	filter["templateid"] = id

	for _, optional_filter := range optional_filters {
		for property, value := range optional_filter {
			filter[property] = value
		}
	}
	params["filter"] = filter
	return api.TriggersGet(params)
}

// Wrapper for trigger.create: https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/create
func (api *API) TriggersCreate(triggers Triggers) (err error) {
	response, err := api.CallWithError("trigger.create", triggers)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	triggerids := result["triggerids"].([]interface{})
	for i, id := range triggerids {
		triggers[i].TriggerId = id.(string)
	}
	return
}

