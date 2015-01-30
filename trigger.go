package zabbix

import (
	"encoding/json"
)

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
	FunctionId string `json:"functionid"`
	ItemId     string `json:"itemid"`
	Function   string `json:"function"`
	Parameter  string `json:"parameter"`
}

// Trigger object
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/definitions
type Trigger struct {
	TriggerId   string            `json:"triggerid"`
	Description string            `json:"description"`
	Functions   []Function        `json:"functions"`
	Expression  string            `json:"expression"`
	Comments    string            `json:"comments"`
	Error       string            `json:"error"`
	Flags       TriggerFlags      `json:"flags,string"`
	LastChange  int64             `json:"lastchange,string"`
	Priority    TriggerPriority   `json:"priority,string"`
	Status      TriggerStatus     `json:"status,string"`
	TemplateId  string            `json:"templateid,omitempty"`
	Type        TriggerType       `json:"type,string"`
	Url         string            `json:"url"`
	Value       TriggerValue      `json:"value,string"`
	ValueFlags  TriggerValueFlags `json:"value_flags,string"`
	// only when expandData flag is set
	//HostId string `json:"hostid,string"`
	HostId string `json:"hostid"`
	Host   string `json:"host"`
}

type Triggers []Trigger

// TriggersGet is a wrapper for 'trigger.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/trigger/get
func (api *API) TriggersGet(params Params) (result Triggers, err error) {
	defaults := Params{
		"output":            "extend",
		"expandExpression":  "true",
		"expandDescription": "true",
		"expandData":        "true",
		"selectFunctions":   "extend",
	}
	for key, defaultValue := range defaults {
		if _, ok := params[key]; !ok {
			params[key] = defaultValue
		}
	}

	if !api.isVersionBigger(2, 0, 0) {
		// Transform parameters for Zabbix 1.8
		if _, ok := params["expandExpression"]; ok {
			delete(params, "expandExpression")
		}
		if value, ok := params["expandDescription"]; ok {
			if value == "true" {
				params["expandDescription"] = "extend"
			}
		}
		if value, ok := params["expandData"]; ok {
			if value == "true" {
				params["expandData"] = "extend"
			}
		}
		if value, ok := params["selectFunctions"]; ok {
			params["select_functions"] = value
			delete(params, "selectFunctions")
		}
	}

	/* Warning! Reflector by AlekSi (from github.com/AlekSi/reflector)
	 * which used in original parts of that API implementation
	 * has some error which caused empty slices, e.g. Trigger.Functions
	 * So we do manual unmarshalling. */
	var response ResponseWithJson
	b, err := api.callBytes("trigger.get", params)
	if err == nil {
		err = json.Unmarshal(b, &response)
	}
	if err == nil && response.Error != nil {
		err = response.Error
	}
	if err != nil {
		return
	}

	result = make(Triggers, 0)
	err = json.Unmarshal(response.Result, &result)

	// transform Zabbix 1.8 status values to a newer ones
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
	params := Params{
		"output":            "extend",
		"expandExpression":  "true",
		"expandDescription": "true",
		"expandData":        "true",
		"selectFunctions":   "extend",
		"triggerids":        []string{id},
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
// Use nil for empty additional parameters or filter
func (api *API) TriggersGetInheritedFromId(id string, params Params, filter map[string]string) (result Triggers, err error) {
	if params == nil {
		params = make(Params)
	}
	params["output"] = "extend"
	params["expandExpression"] = "true"
	params["expandDescription"] = "true"
	params["expandData"] = "true"
	params["inherited"] = 1

	if filter != nil {
		filter["templateid"] = id
	} else {
		filter = map[string]string{"templateid": id}
	}
	params["filter"] = filter
	return api.TriggersGet(params)
}

// TriggersGetByTemplateId gets triggers from template by it's Id
// Use nil for empty additional parameters or filter
func (api *API) TriggersGetByTemplateId(id string, params Params, filter map[string]string) (result Triggers, err error) {
	if params == nil {
		params = make(Params)
	}
	params["output"] = "extend"
	params["expandExpression"] = "true"
	params["expandDescription"] = "true"
	params["expandData"] = "true"
	params["templated"] = 1
	params["templateids"] = []string{id}

	if filter != nil {
		params["filter"] = filter
	}
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
