package user

import (
	"strings"

	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// Controller godoc
// Define the user controller that is responsible for all user related rest operations
type Controller struct {
}

// AddRouters allows the endpoints defined in this controller to be added to router
func (controller Controller) AddRouters(ws *restful.WebService) *restful.WebService {
	// Registering the routes
	ws.Route(ws.POST("/users").Filter(utils.BearerAuth).To(createUsers))
	ws.Route(ws.GET("/users").Filter(utils.BearerAuth).To(searchUsers))
	ws.Route(ws.GET("/me").Filter(utils.BearerAuth).To(getMe))
	ws.Route(ws.GET("/users/{id}").Filter(utils.BearerAuth).To(getUserByID))
	ws.Route(ws.PUT("/users/{id}").Filter(utils.BearerAuth).To(updateUsers))
	ws.Route(ws.DELETE("/users/{id}").Filter(utils.BearerAuth).To(deleteUser))
	return ws
}

// createUsers creates user
func createUsers(req *restful.Request, resp *restful.Response) {
	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA", "GA", "SM") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	// Reading the request model
	request := CreateUserModel{}
	err := req.ReadEntity(&request)
	if err != nil {
		log.Errorf("Error occured while trying to read request model from request, error: %v", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_request_data"))
		return
	}

	// perform model validations
	err = utils.GetValidator().Struct(request)
	if err != nil {
		log.Errorf("Failed validation, error: %v", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_request_data"))
		return
	}

	claims := utils.GetClaims(req)
	// check siteId is required for GA,SM
	if len(strings.TrimSpace(request.SiteID)) == 0 {
		for _, p := range claims.Permissions {
			if p.Role == "GA" || p.Role == "SM" {
				utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
				return
			}
		}
	}

	//Checking permission
	if len(request.AdminUserType) != 0 {
		if len(request.SiteGroupName) != 0 {
			if !utils.HasRole(req, "SA", "AM") {
				log.Infof("User not authorized")
				utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
				return
			}
		} else {
			if !utils.HasRole(req, "SA", "AM", "CSA") {
				log.Infof("User not authorized")
				utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
				return
			}
		}

		if !utils.CanAccessResource(req, "client", request.ClientID) {
			log.Infof("User access forbidden for client id %s", request.ClientID)
			utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
			return
		}

	} else if len(request.SiteUserType) != 0 {
		if !utils.HasRole(req, "SA", "AM", "CSA", "GA", "SM") {
			log.Infof("User not authorized")
			utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
			return
		}

		if !utils.CanAccessResource(req, "site", request.SiteID) {
			log.Infof("User access forbidden for site id %s", request.SiteID)
			utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
			return
		}

	} else {
		if !utils.HasRole(req, "SA", "AM", "CSA") {
			log.Infof("User not authorized")
			utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
			return
		}
	}

	log.Infof("Performing create user")
	// perform operations
	err = GetService().CreateUser(request)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(204, nil)

}

// searchUsers search users by their permission level using query parameter
// and returns list of users if succeeds
func searchUsers(req *restful.Request, resp *restful.Response) {
	// Prepare query model
	query, err := PrepareUserSearchQuery(req)
	if err != nil {
		log.Errorf("Error occured during query preparation, error: %v", err)
		utils.WriteError(resp, err)
		return
	}

	log.Infof("Performing searching")
	var claims = utils.GetClaims(req)
	users, err := GetService().SearchUsers(query, claims.Permissions, utils.GetUserID(req))

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, users)

}

// updateUsers find user by id an update the properties
// and returns updated user if succeeds
func updateUsers(req *restful.Request, resp *restful.Response) {
	// get path value
	id := req.PathParameter("id")
	if id == "" {
		log.Infof("Error occured during getting path value from request")
		utils.WriteError(resp, errors.CreateError(400, "invalid_path_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA", "GA", "SM") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//Check weather user has permission to the resource
	if !utils.CanAccessResource(req, "user", id) {
		log.Infof("User access forbidden for user id %s", id)
		utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
		return
	}

	var claims = utils.GetClaims(req)

	request := UpdateUserModel{}
	err := req.ReadEntity(&request)
	if err != nil {
		log.Errorf("Error occured during getting request data, error: %v", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_request_data"))
		return
	}

	log.Infof("Performing update user")
	user, err := GetService().UpdateUser(id, request, claims.Permissions, claims.ID)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, user)
}

// getUserByID find user by id
// and returns user if succeeds
func getUserByID(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("id")
	if id == "" {
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	var claims = utils.GetClaims(req)
	user, err := GetService().GetUser(id, claims.Permissions, utils.GetUserID(req))

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, user)
}

// getMe find the current login user
// and returns user if succeeds
func getMe(req *restful.Request, resp *restful.Response) {
	// Get id from token
	id := utils.GetUserID(req)
	if id == "" {
		log.Infof("Error occured during getting id from token")
		utils.WriteError(resp, errors.CreateError(400, "invalid_path_data"))
		return
	}

	var claims = utils.GetClaims(req)
	user, err := GetService().GetUser(id, claims.Permissions, utils.GetUserID(req))

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(200, user)
}

// deleteUser find a user by id and delete it
// and returns nothing if succeeds
func deleteUser(req *restful.Request, resp *restful.Response) {
	// get path value
	id := req.PathParameter("id")
	if id == "" {
		log.Infof("Error occured during getting path value from request")
		utils.WriteError(resp, errors.CreateError(400, "invalid_path_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA", "GA", "SM") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	//Check weather user has permission to the resource
	if !utils.CanAccessResource(req, "user", id) {
		log.Infof("User access forbidden for user id %s", id)
		utils.WriteError(resp, errors.CreateError(403, "Forbidden"))
		return
	}

	if id == utils.GetUserID(req) {
		log.Infof("Error occured during getting path value from request")
		utils.WriteError(resp, errors.CreateError(400, "cannot delete own account"))
		return
	}

	err := GetService().DeleteUser(id)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeaderAndEntity(204, nil)
}
