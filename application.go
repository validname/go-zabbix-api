package zabbix

import (
	"github.com/AlekSi/reflector"
)

// Application object
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/application/definitions
type Application struct {
	ApplicationId string `json:"applicationid,omitempty"`
	HostId        string `json:"hostid"`
	Name          string `json:"name"`
	TemplateId    string `json:"templateid,omitempty"`
}

type Applications []Application

// ApplicationsGet is a wrapper for 'application.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/application/get
func (api *API) ApplicationsGet(params Params) (res Applications, err error) {
	if _, ok := params["output"]; !ok {
		params["output"] = "extend"
	}
	response, err := api.CallWithError("application.get", params)
	if err != nil {
		return
	}

	reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}

// ApplicationGetById gets application by Id only if there is exactly 1 matching application.
func (api *API) ApplicationGetById(id string) (res *Application, err error) {
	apps, err := api.ApplicationsGet(Params{"applicationids": id})
	if err != nil {
		return
	}

	if len(apps) == 1 {
		res = &apps[0]
	} else {
		e := ExpectedOneResult(len(apps))
		err = &e
	}
	return
}

// ApplicationGetByHostIdAndName gets application by host Id and name only if there is exactly 1 matching application.
func (api *API) ApplicationGetByHostIdAndName(hostId, name string) (res *Application, err error) {
	apps, err := api.ApplicationsGet(Params{"hostids": hostId, "filter": map[string]string{"name": name}})
	if err != nil {
		return
	}

	if len(apps) == 1 {
		res = &apps[0]
	} else {
		e := ExpectedOneResult(len(apps))
		err = &e
	}
	return
}

// ApplicationsCreate is a wrapper for 'application.create'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/application/create
func (api *API) ApplicationsCreate(apps Applications) (err error) {
	response, err := api.CallWithError("application.create", apps)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	applicationids := result["applicationids"].([]interface{})
	for i, id := range applicationids {
		apps[i].ApplicationId = id.(string)
	}
	return
}

// ApplicationsDelete is a wrapper for 'application.delete'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/application/delete
// Deletes applications by given application list
// Cleans ApplicationId in all apps elements if call succeed.
func (api *API) ApplicationsDelete(apps Applications) (err error) {
	ids := make([]string, len(apps))
	for i, app := range apps {
		ids[i] = app.ApplicationId
	}

	err = api.ApplicationsDeleteByIds(ids)
	if err == nil {
		for i := range apps {
			apps[i].ApplicationId = ""
		}
	}
	return
}

// ApplicationsDeleteByIds is a wrapper for 'application.delete'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/application/delete
// Deletes applications by given Id list
func (api *API) ApplicationsDeleteByIds(ids []string) (err error) {
	response, err := api.CallWithError("application.delete", ids)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	applicationids := result["applicationids"].([]interface{})
	if len(ids) != len(applicationids) {
		err = &ExpectedMore{len(ids), len(applicationids)}
	}
	return
}
