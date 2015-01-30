package zabbix

import (
	"encoding/json"
	"fmt"
)

type (
	ItemType  int
	ValueType int
	DataType  int
	DeltaType int
)

const (
	ZabbixAgent       ItemType = 0
	SNMPv1Agent       ItemType = 1
	ZabbixTrapper     ItemType = 2
	SimpleCheck       ItemType = 3
	SNMPv2Agent       ItemType = 4
	ZabbixInternal    ItemType = 5
	SNMPv3Agent       ItemType = 6
	ZabbixAgentActive ItemType = 7
	ZabbixAggregate   ItemType = 8
	WebItem           ItemType = 9
	ExternalCheck     ItemType = 10
	DatabaseMonitor   ItemType = 11
	IPMIAgent         ItemType = 12
	SSHAgent          ItemType = 13
	TELNETAgent       ItemType = 14
	Calculated        ItemType = 15
	JMXAgent          ItemType = 16

	Float     ValueType = 0
	Character ValueType = 1
	Log       ValueType = 2
	Unsigned  ValueType = 3
	Text      ValueType = 4

	Decimal     DataType = 0
	Octal       DataType = 1
	Hexadecimal DataType = 2
	Boolean     DataType = 3

	AsIs  DeltaType = 0
	Speed DeltaType = 1
	Delta DeltaType = 2
)

type AppInfo struct {
	HostList      HostIds `json:"hosts"`
	ApplicationId string   `json:"applicationid"`
	Name          string   `json:"name"`
	TemplateId    string   `json:"templateid"`
}

// Item object
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/item/definitions
type Item struct {
	ItemId      string    `json:"itemid,omitempty"`
	Delay       int       `json:"delay,string"`
	HostId      string    `json:"hostid"`
	InterfaceId string    `json:"interfaceid,omitempty"`
	Key         string    `json:"key_"`
	LastValue   string    `json:"lastvalue"`
	LastClock   string    `json:"lastclock"`
	Units       string    `json:"units"`
	Name        string    `json:"name"`
	Type        ItemType  `json:"type,string"`
	ValueType   ValueType `json:"value_type,string"`
	DataType    DataType  `json:"data_type,string"`
	Delta       DeltaType `json:"delta,string"`
	Description string    `json:"description"`
	Error       string    `json:"error"`
	History     int       `json:"history,omitempty,string"`
	Trends      int       `json:"trends,omitempty,string"`

	// Field below used for receiving from Zabbix server and
	// used only when 'selectApplications' parameter is set
	Applications []AppInfo `json:"applications,omitempty"`

	// Field below used only when creating applications
	// It used once when invoking ItemsCreate() for backward compatibility with AlekSi old code
	ApplicationIds []string
}

type Items []Item

// Used only for for marshalling JSON in the ItemsCreate() function
type ItemWrite struct {
	Item
	ApplicationIds []string `json:"applications"`
}

// Converts slice to map by key. Panics if there are duplicate keys.
func (items Items) ByKey() (res map[string]Item) {
	res = make(map[string]Item, len(items))
	for _, i := range items {
		_, ok := res[i.Key]
		if ok {
			panic(fmt.Errorf("Duplicate key %s", i.Key))
		}
		res[i.Key] = i
	}
	return
}

// ItemsGet is a wrapper for 'item.get'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/item/get
func (api *API) ItemsGet(params Params) (result Items, err error) {
	if _, ok := params["output"]; !ok {
		params["output"] = "extend"
	}
	if _, ok := params["selectApplications"]; !ok {
		params["selectApplications"] = "extend"
	}

	if !api.isVersionBigger(2, 0, 0) {
		// Transform parameters for Zabbix 1.8
		if _, ok := params["selectApplications"]; ok {
			params["select_applications"] = "extend"
			// it's a hidden option from PHP sources, one must use it to enable 'select_*' options
			params["extendoutput"] = 1
			delete(params, "selectApplications")
		}
	}

	/* Warning! Reflector by AlekSi (from github.com/AlekSi/reflector)
	 * which used in original parts of that API implementation
	 * has some error which caused empty slices, e.g. Item.Applications
	 * So we do manual unmarshalling. */
	var response ResponseWithJson
	b, err := api.callBytes("item.get", params)
	if err == nil {
		err = json.Unmarshal(b, &response)
	}
	if err == nil && response.Error != nil {
		err = response.Error
	}
	if err != nil {
		return
	}

	result = make(Items, 0)
	err = json.Unmarshal(response.Result, &result)

	if err == nil {
		if !api.isVersionBigger(2, 0, 0) {
			// Transform results from Zabbix 1.8
			for idx, _ := range result {
				result[idx].Name = result[idx].Description
				result[idx].Description = ""
			}
		}
	}

	return
}

// ItemGetById gets items by Id only if there is exactly 1 matching item.
func (api *API) ItemGetById(id string) (result *Item, err error) {
	items, err := api.ItemsGet(Params{"itemids": []string{id}})
	if err != nil {
		return
	}
	if len(items) == 1 {
		result = &items[0]
	} else {
		e := ExpectedOneResult(len(items))
		err = &e
	}
	return
}

// ItemsGetByApplicationId gets items by application Id.
func (api *API) ItemsGetByApplicationId(id string) (res Items, err error) {
	return api.ItemsGet(Params{"applicationids": id})
}

// ItemsCreate is a wrapper for 'item.create'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/item/create
func (api *API) ItemsCreate(items Items) (err error) {
	// fill ItemWrite structure
	itemsToWrite := make([]ItemWrite, len(items))
	for idx, item := range items {
		// just assign pointer, no need to copy full structure
		itemsToWrite[idx].Item = item
		itemsToWrite[idx].ApplicationIds = make([]string,len(itemsToWrite[idx].Item.ApplicationIds))
		// force empty Applications field to prevent it's marshalling
		itemsToWrite[idx].Item.Applications = nil
		copy(itemsToWrite[idx].ApplicationIds, itemsToWrite[idx].Item.ApplicationIds)
		itemsToWrite[idx].Item.ApplicationIds = nil

		if !api.isVersionBigger(2, 0, 0) {
			// Transform parameters for Zabbix 1.8
			itemsToWrite[idx].Item.Description = itemsToWrite[idx].Item.Name
			itemsToWrite[idx].Item.Name = ""
		}
	}

	response, err := api.CallWithError("item.create", itemsToWrite)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	itemids := result["itemids"].([]interface{})
	for i, id := range itemids {
		items[i].ItemId = id.(string)
	}
	return
}

// ItemsUpdate is a wrapper for 'item.update'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/item/update
func (api *API) ItemsUpdate(items Items) (err error) {
	// fill ItemWrite structure
	itemsToWrite := make([]ItemWrite, len(items))
	for idx, item := range items {
		// just assign pointer, no need to copy full structure
		itemsToWrite[idx].Item = item
		itemsToWrite[idx].ApplicationIds = make([]string,len(itemsToWrite[idx].Item.ApplicationIds))
		// force empty Applications field to prevent it's marshalling
		itemsToWrite[idx].Item.Applications = nil
		copy(itemsToWrite[idx].ApplicationIds, itemsToWrite[idx].Item.ApplicationIds)
		itemsToWrite[idx].Item.ApplicationIds = nil

		if !api.isVersionBigger(2, 0, 0) {
			// Transform parameters for Zabbix 1.8
			itemsToWrite[idx].Item.Description = itemsToWrite[idx].Item.Name
			itemsToWrite[idx].Item.Name = ""
		}

	}
	response, err := api.CallWithError("item.update", itemsToWrite)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	itemids := result["itemids"].([]interface{})
	for i, id := range itemids {
		items[i].ItemId = id.(string)
	}
	return
}

// ItemsDelete is a wrapper for 'item.delete'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/item/delete
// Delete by Items list
// Cleans ItemId in all items elements if call succeed.
func (api *API) ItemsDelete(items Items) (err error) {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ItemId
	}

	err = api.ItemsDeleteByIds(ids)
	if err == nil {
		for i := range items {
			items[i].ItemId = ""
		}
	}
	return
}

// ItemsDeleteByIds is a wrapper for 'item.delete'
// see https://www.zabbix.com/documentation/2.0/manual/appendix/api/item/delete
// Delete by Ids list
func (api *API) ItemsDeleteByIds(ids []string) (err error) {
	response, err := api.CallWithError("item.delete", ids)
	if err != nil {
		return
	}

	result := response.Result.(map[string]interface{})
	itemids1, ok := result["itemids"].([]interface{})
	l := len(itemids1)
	if !ok {
		// some versions actually return map there
		itemids2 := result["itemids"].(map[string]interface{})
		l = len(itemids2)
	}
	if len(ids) != l {
		err = &ExpectedMore{len(ids), l}
	}
	return
}
