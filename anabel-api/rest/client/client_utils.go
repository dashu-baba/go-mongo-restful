package client

import (
	"encoding/json"
	"fmt"
	"strconv"

	"anacove.com/backend/common"
	"anacove.com/backend/utils"

	"anacove.com/backend/errors"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// ToUpdateResponseModel convert common.Client to UpdateResponseModel from common.Client
func ToUpdateResponseModel(model UpdateResponseModel, client common.Client) UpdateResponseModel {
	bytes, err := json.Marshal(&client)
	if err != nil {
		log.Errorf("error occurred during marshalling: error: %v\n", err)
		return model
	}
	json.Unmarshal(bytes, &model)

	return model
}

// ToClient convert UpdateRequestModel to client
func (model *UpdateRequestModel) ToClient(client common.Client, permissions []common.Permission) (common.Client, error) {
	bytes, err := json.Marshal(&model)
	if err != nil {
		log.Errorf("error occurred during marshalling: error: %v\n", err)
		return client, nil
	}

	address := Address{}
	json.Unmarshal(bytes, &address)
	if (Address{}) != address {
		//Check CSA, AM persons readonly policy on client name
		if len(permissions) == 1 && permissions[0].Role != "SA" && len(model.Name) > 0 {
			return client, errors.CreateError(403, "forbidden_name_change")
		}

		err = utils.GetValidator().Struct(address)
		if err != nil {
			log.Errorf("Address validation error: error: %v\n", err)
			return client, errors.CreateError(400, "invalid_data")
		}
		client = address.ToClient(client)
	}

	configuration := Configuration{}
	json.Unmarshal(bytes, &configuration)
	if !configuration.IsEmpty() {
		client = configuration.ToClient(client)
	}

	gd := GroupDefinition{}
	json.Unmarshal(bytes, &gd)
	if !gd.IsEmpty() {
		client = gd.ToClient(client)
	}

	return client, nil
}

// ToClient converts Address to common.Client from address
func (model *Address) ToClient(client common.Client) common.Client {
	bytes, err := json.Marshal(&model)
	if err != nil {
		log.Errorf("error occurred during marshalling: error: %v\n", err)
		return client
	}

	json.Unmarshal(bytes, &client)
	addressLine := model.Address.Line1
	if len(model.Address.Line2) > 0 {
		addressLine += " " + model.Address.Line2
	}
	client.FullAddress = fmt.Sprintf("%s, %s, %s %s", addressLine, model.Address.City, model.Address.State, model.Address.Zip)

	return client
}

// ToClient convert GroupDefinition to common.Client from group definition
func (model *GroupDefinition) ToClient(client common.Client) common.Client {
	bytes, err := json.Marshal(&model)
	if err != nil {
		log.Errorf("error occurred during marshalling: error: %v\n", err)
		return client
	}
	json.Unmarshal(bytes, &client)

	return client
}

// IsEmpty check the group definition entity empty
func (model *GroupDefinition) IsEmpty() bool {
	return len(model.Groups) == 0
}

// IsEmpty check the configuration model empty
func (model *Configuration) IsEmpty() bool {
	if len(model.Configuration.FS.FutureModules) > 0 {
		return false
	}

	if (common.Item{}) != model.Configuration.FS.AnalyticsLevel1 {
		return false
	}

	if (common.Item{}) != model.Configuration.FS.TVTheftPrevention {
		return false
	}

	if len(model.Configuration.RequiredDevice) > 0 {
		return false
	}

	if len(model.Configuration.TFS.FutureModules) > 0 {
		return false
	}

	if (common.Item{}) != model.Configuration.TFS.AnalyticsLevel1 {
		return false
	}

	if (common.Item{}) != model.Configuration.TFS.AnalyticsLevel1 {
		return false
	}

	if (common.Item{}) != model.Configuration.TFS.TVTheftPrevention {
		return false
	}

	if model.Configuration.FS.SuspendClientAccess {
		return false
	}

	if model.Configuration.FS.StaffAlert {
		return false
	}

	return true
}

// ToClient Convert to common.Client from configuration
func (model *Configuration) ToClient(client common.Client) common.Client {
	bytes, err := json.Marshal(&model)
	if err != nil {
		log.Errorf("error occurred during marshalling: error: %v\n", err)
		return client
	}

	json.Unmarshal(bytes, &client)

	return client
}

// Convert to common.Client domain model
func (model *CreateClientModel) toClient() common.Client {
	bytes, err := json.Marshal(&model)
	if err != nil {
		log.Errorf("error occurred during marshalling: error: %v\n", err)
		return common.Client{}
	}

	client := common.Client{}
	json.Unmarshal(bytes, &client)

	return client
}

//PrepareClientSearchQuery Get client search query
func PrepareClientSearchQuery(req *restful.Request) (*Query, error) {
	query := Query{
		PageNumber: 1,
		PageSize:   20,
		SortOrder:  -1,
		Status:     common.Active,
	}

	val := req.QueryParameter("pageNumber")
	if val != "" {
		i, err := strconv.Atoi(val)
		if err != nil {
			log.Errorf("error occurred during conversion: error: %v\n", err)
			return nil, errors.CreateError(400, "invalid_data")
		}

		query.PageNumber = i
	}

	val = req.QueryParameter("pageSize")
	if val != "" {
		i, err := strconv.Atoi(val)
		if err != nil {
			log.Errorf("error occurred during conversion: error: %v\n", err)
			return nil, errors.CreateError(400, "invalid_data")
		}

		query.PageSize = i
	}

	query.SortBy = req.QueryParameter("sortBy")
	query.Status = req.QueryParameter("status")
	query.Keyword = req.QueryParameter("keyword")
	val = req.QueryParameter("sortOrder")
	if val != "" {
		if val == "asc" {
			query.SortOrder = 1
		}
	}

	return &query, nil
}
