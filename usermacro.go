package zabbix

import (
	"encoding/json"
)

// Macro object
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/usermacro/definitions
type UserMacro struct {
	Id            string `json:"id,omitempty"`
	GlobalMacroId string `json:"globalmacroid,omitempty"`
	HostMacroId   string `json:"hostmacroid,omitempty"`
	HostId        string `json:"hostid,omitempty"`
	Macro         string `json:"macro"`
	Value         string `json:"value"`
}

type UserMacros []UserMacro

// UserMacrosGet is a wrapper for 'usermacro.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/usermacro/get

func (api *API) UserMacrosGet(params Params) (result UserMacros, err error) {
	defaults := Params{
		"output": "extend",
	}
	for key, defaultValue := range defaults {
		if _, ok := params[key]; !ok {
			params[key] = defaultValue
		}
	}

	var response ResponseWithJson
	b, err := api.callBytes("usermacro.get", params)
	if err == nil {
		err = json.Unmarshal(b, &response)
	}
	if err == nil && response.Error != nil {
		err = response.Error
	}
	if err != nil {
		return
	}

	result = make(UserMacros, 0)
	err = json.Unmarshal(response.Result, &result)
	return
}

// UserMacroGetByMacro gets user macro by it's name only if there is exactly 1 matching host group.
// Use nil for empty additional parameters or filter
func (api *API) UserMacroGetGlobalByMacro(macro string) (result *UserMacro, err error) {
	macros, err := api.UserMacrosGet(Params{
		"globalmacro": "1",
		"filter": Params{
			"macro": macro,
		},
	})
	if err != nil {
		return
	}

	if len(macros) == 1 {
		result = &macros[0]
	} else {
		e := ExpectedOneResult(len(macros))
		err = &e
	}
	return
}
