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

	TriggerEnabled  TriggerStatus = 0
	TriggerDisabled TriggerStatus = 1

	Nomultiple            TriggerType = 0
	GenerateMulipleEvents TriggerType = 1

	TriggerOk        TriggerValue = 0
	TriggerProblem   TriggerValue = 1
	TriggerUnknown18 TriggerValue = 2

	TriggerUpToDate TriggerValueFlags = 0
	TriggerUnknown  TriggerValueFlags = 1
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
			if trigger.Value == TriggerUnknown18 {
				trigger.ValueFlags = TriggerUnknown
				trigger.Value = TriggerOk
			}
		}
	}
	return
}

// Get trigger extended information by Id only if there is exactly 1 matching trigger
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

// Return triggers on hosts which was inherited from template trigger
//
func (api *API) TriggersGetInheritedFromId(id string,
	OptionalFilters ...map[string]string) (result Triggers, err error) {

	params := map[string]interface{}{
		"output":            "extend",
		"expandExpression":  "extend",
		"expandDescription": "flag",
		"expandData":        "extend",
		"inherited":         1,
	}

	filter := make(map[string]string)
	filter["templateid"] = id

	for _, optionalFilter := range OptionalFilters {
		for property, value := range optionalFilter {
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
