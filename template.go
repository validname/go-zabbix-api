package zabbix

import (
	"github.com/AlekSi/reflector"
)

// Template object
// see https://www.zabbix.com/documentation/1.8/api/template
type Template struct {
	TemplateId string `json:"templateid,omitempty"`
	Host       string `json:"host"`
	Name       string `json:"name"`

	// Fields below used only when creating templates
	GroupIds    HostGroupIds        `json:"groups,omitempty"`
	TemplateIds TemplateIds         `json:"templates,omitempty"`
	Macros      map[string]string `json:"macros,omitempty"`
	HostIds     HostIds             `json:"hosts,omitempty"`
}

type Templates []Template

type TemplateId struct {
	TemplateId string `json:"templateid"`
}

type TemplateIds []TemplateId

// TemplateGet is a wrapper for 'template.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/template/get
func (api *API) TemplatesGet(params Params) (res Templates, err error) {
	if _, ok := params["output"]; !ok {
		params["output"] = "extend"
	}
	response, err := api.CallWithError("template.get", params)
	if err != nil {
		return
	}

	reflector.MapsToStructs2(response.Result.([]interface{}), &res, reflector.Strconv, "json")
	return
}

// TemplateGetById gets template by Id only if there is exactly 1 matching template.
func (api *API) TemplateGetById(id string) (res *Template, err error) {
	templates, err := api.TemplatesGet(Params{"templateids": id})
	if err != nil {
		return
	}

	if len(templates) == 1 {
		res = &templates[0]
	} else {
		e := ExpectedOneResult(len(templates))
		err = &e
	}
	return
}

// TemplateGetByTemplate gets template by Template only if there is exactly 1 matching template.
func (api *API) TemplateGetByHost(host string) (res *Template, err error) {
	templates, err := api.TemplatesGet(Params{"filter": map[string]string{"host": host}})
	if err != nil {
		return
	}

	if len(templates) == 1 {
		res = &templates[0]
	} else {
		e := ExpectedOneResult(len(templates))
		err = &e
	}
	return
}

// TemplateGetByTemplate gets template by Template only if there is exactly 1 matching template.
func (api *API) TemplateGetByHostIds(ids []string) (res Templates, err error) {
	return api.TemplatesGet(Params{"hostids": ids})
}

// TemplatesCreate is a wrapper for 'template.create'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/template/create
func (api *API) TemplatesCreate(templates Templates) (err error) {
	response, err := api.CallWithError("template.create", templates)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	templateids := result["templateids"].([]interface{})
	for i, id := range templateids {
		templates[i].TemplateId = id.(string)
	}
	return
}

