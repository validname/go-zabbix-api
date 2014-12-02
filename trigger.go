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

	NotClassified TriggerPriority = 0
	Information   TriggerPriority = 1
	Warning       TriggerPriority = 2
	Average       TriggerPriority = 3
	High          TriggerPriority = 4
	Disaster      TriggerPriority = 5

	TriggerEnabled  TriggerStatus = 0
	TriggerDisabled TriggerStatus = 1

	Nomultiple            TriggerType = 0
	GenerateMulipleEvents TriggerType = 1

	TriggerValueOk      TriggerValue = 0
	TriggerValueProblem TriggerValue = 1
	TriggerValueUnknown TriggerValue = 2

	TriggerValueFlagsUpToDate TriggerValueFlags = 0
	TriggerValueFlagsUnknown  TriggerValueFlags = 1
)

type Function struct {
	FunctionId int64  `json:"functionid,string"`
	ItemId     int64  `json:"itemid,string"`
	Function   string `json:"function"`
	Parameter  string `json:"parameter"`
}

// Trigger object
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/definitions
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

// TriggersGet is a wrapper for 'trigger.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/get
func (api *API) TriggersGet(params Params) (result Triggers, err error) {
	if _, ok := params["output"]; !ok {
		params["output"] = "extend"
	}
	if _, ok := params["expandExpression"]; !ok {
		params["expandExpression"] = "extend"
	}
	if _, ok := params["expandDescription"]; !ok {
		params["expandDescription"] = "flag"
	}
	if _, ok := params["expandData"]; !ok {
		params["expandData"] = "extend"
	}
	if _, ok := params["selectFunctions"]; !ok {
		params["selectFunctions"] = "extend"
	}

	response, err := api.CallWithError("trigger.get", params)
	if err != nil {
		return
	}
	reflector.MapsToStructs2(response.Result.([]interface{}), &result, reflector.Strconv, "json")

	// mimic Zabbix 1.8 status values to a newer ones
	if !api.isVersionBigger(2, 0, 0) {
		for _, trigger := range result {
			if trigger.Value == TriggerValueUnknown {
				trigger.ValueFlags = TriggerValueFlagsUnknown
				trigger.Value = TriggerValueOk
			}
		}
	}
	return
}

// TriggerGetById gets trigger extended information by Id only if there is exactly 1 matching trigger
func (api *API) TriggerGetById(id string) (result *Trigger, err error) {
	params := map[string]interface{}{
		"output":            "extend",
		"expandExpression":  "extend",
		"expandDescription": "flag",
		"expandData":        "extend",
		"selectFunctions":   "extend",
	}

	triggers, err := api.TriggersGet(params)
	if err != nil {
		return
	}
	if len(triggers) == 1 {
		result = &triggers[0]
	} else {
		e := ExpectedOneResult(len(triggers))
		err = &e
	}
	return
}

// TriggersGetInheritedFromId gets triggers on hosts which was inherited from template trigger
// Use nil for empty filter
func (api *API) TriggersGetInheritedFromId(id string, Filter map[string]string) (result Triggers, err error) {
	params := map[string]interface{}{
		"output":            "extend",
		"expandExpression":  "extend",
		"expandDescription": "flag",
		"expandData":        "extend",
		"inherited":         1,
	}

	filter := make(map[string]string)
	filter["templateid"] = id

	for property, value := range Filter {
		filter[property] = value
	}
	params["filter"] = filter
	return api.TriggersGet(params)
}

// TriggersCreate is a wrapper for 'trigger.create'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/create
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
