package client

import (
	"encoding/json"
	"strings"

	"github.com/globalsign/mgo/bson"

	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// Controller godoc
// @summary Controller type
type Controller struct {
}

// AddRouters allows the endpoints defined in this controller to be added to router
func (controller Controller) AddRouters(ws *restful.WebService) *restful.WebService {
	ws.Route(ws.POST("/clients").Filter(utils.BearerAuth).To(createClients))
	ws.Route(ws.GET("/clients").Filter(utils.BearerAuth).To(searchClients))
	ws.Route(ws.GET("/clients/{clientId}/site-groups").Filter(utils.BearerAuth).To(getSiteGroups))
	ws.Route(ws.PUT("/clients/{clientId}/archive").Filter(utils.BearerAuth).To(archiveClient))
	ws.Route(ws.GET("/clients/{clientId}").Filter(utils.BearerAuth).To(getClientByID))
	ws.Route(ws.PUT("/clients/{clientId}").Filter(utils.BearerAuth).To(updateClients))
	ws.Route(ws.DELETE("/clients/{clientId}").Filter(utils.BearerAuth).To(deleteClient))
	return ws
}

// createClients uses the provided model to create client in the system
// and returns no content if succeeds
func createClients(req *restful.Request, resp *restful.Response) {
	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//parsing data from request
	request := CreateClientModel{}
	err := req.ReadEntity(&request)
	if err != nil {
		log.Errorf("Request data is not valid: error %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Create client and get the details
	client, err := GetClientService().CreateClient(request)

	if err != nil {
		log.Errorf("Error happend at service: error %v\n", err)
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(204, client)

}

// searchClients search clients in the system by query parameter
// and returns list of clients if succeeds
func searchClients(req *restful.Request, resp *restful.Response) {
	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	query, err := PrepareClientSearchQuery(req)
	if err != nil {
		log.Errorf("error occurred during query model parsing: error: %v\n", err)
		utils.WriteError(resp, err)
		return
	}

	var claims = utils.GetClaims(req)
	clients, err := GetClientService().SearchClients(query, claims.Permissions)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, clients)

}

// updateClients find client by id an update the properties
// and returns updated client if succeeds
func updateClients(req *restful.Request, resp *restful.Response) {
	//Get id from path and check validation
	id := req.PathParameter("clientId")
	if len(strings.TrimSpace(id)) == 0 || !bson.IsObjectIdHex(id) {
		log.Infof("invalid id property: id: %s\n", id)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//Check weather user has permission to the resource
	if !utils.CanAccessResource(req, "client", id) {
		log.Infof("access forbidden for client id %s", id)
		utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
		return
	}

	model := UpdateRequestModel{}
	err := req.ReadEntity(&model)
	if err != nil {
		log.Errorf("error occurred during reading entity from request: error: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	var claims = utils.GetClaims(req)
	//Check permission to edit configuration
	configuration := Configuration{}
	bytes, _ := json.Marshal(&model)
	json.Unmarshal(bytes, &configuration)
	if !configuration.IsEmpty() {
		for _, p := range claims.Permissions {
			if p.Role == "CSA" {
				log.Infof("Configuration access forbidden for client id %s", id)
				utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
				return
			}
		}
	}

	client, err := GetClientService().UpdateClient(id, model, claims.Permissions)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, client)
}

// getClientByID find client by id
// and returns client if succeeds
func getClientByID(req *restful.Request, resp *restful.Response) {
	//Get id from path and check validation
	id := req.PathParameter("clientId")
	if len(strings.TrimSpace(id)) == 0 || !bson.IsObjectIdHex(id) {
		log.Infof("invalid property id %s", id)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//Check weather user has permission to the resource
	if !utils.CanAccessResource(req, "client", id) {
		log.Infof("User access forbidden for client id %s", id)
		utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
		return
	}

	client, err := GetClientService().GetClient(id)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, client)
}

// deleteClient find a client by id and delete it
func deleteClient(req *restful.Request, resp *restful.Response) {
	//Get id from path and check validation
	id := req.PathParameter("clientId")
	if len(strings.TrimSpace(id)) == 0 || !bson.IsObjectIdHex(id) {
		log.Infof("invalid property id %s", id)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//Check weather user has permission to the resource
	if !utils.CanAccessResource(req, "client", id) {
		log.Infof("User access forbidden for client id %s", id)
		utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
		return
	}

	err := GetClientService().DeleteClient(id)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(204, nil)
}

// archiveClient find a client by id and archive it
// and returns nothing if succeeds
func archiveClient(req *restful.Request, resp *restful.Response) {
	//Get id from path and check validation
	id := req.PathParameter("clientId")
	if len(strings.TrimSpace(id)) == 0 || !bson.IsObjectIdHex(id) {
		log.Infof("invalid property id %s", id)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//Check weather user has permission to the resource
	if !utils.CanAccessResource(req, "client", id) {
		log.Infof("User access forbidden for client id %s", id)
		utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
		return
	}

	err := GetClientService().ArchiveClient(id)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(204, nil)
}

// getSiteGroups find site groups by client id
// and returns id and siteGroupName if succeeds
func getSiteGroups(req *restful.Request, resp *restful.Response) {
	//Get id from path and check validation
	id := req.PathParameter("clientId")
	if len(strings.TrimSpace(id)) == 0 || !bson.IsObjectIdHex(id) {
		log.Infof("invalid property id %s", id)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//Check weather user has permission to the resource
	if !utils.CanAccessResource(req, "client", id) {
		log.Infof("User access forbidden for client id %s", id)
		utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
		return
	}

	res, err := GetClientService().GetSiteGroup(id)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, res)
}
