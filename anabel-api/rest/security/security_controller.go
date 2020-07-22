package security

import (
	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// SecurityController type
type SecurityController struct {
}

// AddRouters allows the endpoints defined in this controller to be added to router
func (controller SecurityController) AddRouters(ws *restful.WebService) *restful.WebService {

	ws.Route(ws.POST("/login").To(login))
	ws.Route(ws.POST("/logout").Filter(utils.BearerAuth).To(logout))
	ws.Route(ws.POST("/initiate-forgot-password").To(forgotPassword))
	ws.Route(ws.POST("/change-password").Filter(utils.BearerAuth).To(changePassword))
	ws.Route(ws.PUT("/user-confirmation").To(confirmUser))
	return ws
}

// login uses the provided username and password to authenticate toward the auth system
// and returns a valid token with user data if login succeeds
func login(req *restful.Request, resp *restful.Response) {
	logingRequest := struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	}{}

	//Parasing request model from request
	err := req.ReadEntity(&logingRequest)
	if err != nil {
		log.Errorf("error read entity from request: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Call service method to get data
	response, err := GetService().Login(logingRequest.Account, logingRequest.Password)
	if err != nil {
		log.Errorf("error calling service method: error %v\n", err)
		utils.WriteError(resp, err)
		return
	}

	resp.WriteEntity(response)
}

// logout uses the token and remove it from database to invalidate that for the next requests
func logout(req *restful.Request, resp *restful.Response) {
	err := GetService().Logout(utils.GetUserID(req))

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeader(204)
}

// forgotPassword check account by email and if exists send a activation mail to email with code
func forgotPassword(req *restful.Request, resp *restful.Response) {
	request := struct {
		Email string `validate:"required" json:"email"`
	}{}
	err := req.ReadEntity(&request)
	if err != nil {
		log.Errorf("error read entity from request, error: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	err = utils.GetValidator().Struct(request)
	if err != nil {
		log.Errorf("error validate entity, error: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	err = GetService().ForgotPassword(request.Email)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeader(204)
}

// changePassword requests to update new password using old one.
func changePassword(req *restful.Request, resp *restful.Response) {
	request := struct {
		OldPassword string `validate:"required" json:"oldPassword"`
		NewPassword string `validate:"required" json:"newPassword"`
	}{}

	err := req.ReadEntity(&request)
	if err != nil {
		log.Errorf("error read entity from request, error: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	err = utils.GetValidator().Struct(request)
	if err != nil {
		log.Errorf("error validate entity, error: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	//Check weather user has permission to perform this operation
	if !utils.HasRole(req, "SA", "AM", "CSA", "GA", "SM", "SU") {
		log.Infof("User not authorized")
		utils.WriteError(resp, errors.CreateError(401, "Not Authorized"))
		return
	}

	err = GetService().ChangePassword(request.OldPassword, request.NewPassword, utils.GetUserID(req))

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeader(204)
}

// confirmUser checks user by activation token and if found success update the data
func confirmUser(req *restful.Request, resp *restful.Response) {
	request := UserConfirmationModel{}
	err := req.ReadEntity(&request)
	if err != nil {
		log.Errorf("error read entity from request, error: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	err = utils.GetValidator().Struct(request)
	if err != nil {
		log.Errorf("error validate entity, error: %v\n", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_data"))
		return
	}

	err = GetService().ConfirmUser(request)

	if err != nil {
		utils.WriteError(resp, err)
		return
	}

	resp.WriteHeader(204)
}
