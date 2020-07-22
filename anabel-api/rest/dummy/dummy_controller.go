package dummy

import (
	"anacove.com/backend/common"
	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// Controller godoc
// Define the dummy controller that is responsible for all dummy rest operations
type Controller struct {
}

// AddRouters allows the endpoints defined in this controller to be added to router
func (controller Controller) AddRouters(ws *restful.WebService) *restful.WebService {
	// Registering the routes
	ws.Route(ws.POST("/dummy").To(createSA))
	return ws
}

type User struct {
	Email                  string              `json:"email" bson:"email"`
	FirstName              string              `json:"firstName" bson:"firstName"`
	FamilyName             string              `json:"familyName" bson:"familyName"`
	ProfileURL             string              `json:"profileUrl" bson:"profileUrl"`
	Position               string              `json:"position" bson:"position"`
	Phone                  string              `json:"phone" bson:"phone"`
	NotificationPreference string              `json:"notificationPreference" bson:"notificationPreference"`
	Password               string              `json:"password" bson:"password"`
	Permission             []common.Permission `json:"permissions" bson:"permissions"`
}

// createSA creates SA user
func createSA(req *restful.Request, resp *restful.Response) {

	user := User{}

	err := req.ReadEntity(&user)
	if err != nil {
		log.Errorf("Error occured during getting request data, error: %v", err)
		utils.WriteError(resp, errors.CreateError(400, "invalid_request_data"))
		return
	}

	GetService().CreateUser(user)

	resp.WriteHeaderAndEntity(204, nil)
}
