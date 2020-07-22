package user

import (
	"encoding/json"
	"strconv"
	"time"

	"anacove.com/backend/common"
	"anacove.com/backend/errors"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// ToUser will convert to User domain model from UpdateUserModel
func (model *UpdateUserModel) ToUser(user *common.User) {
	if len(model.NotificationPreference) > 0 {
		user.NotificationPreference = model.NotificationPreference
	}

	if len(model.Phone) > 0 {
		user.Phone = model.Phone
	}

	if len(model.Position) > 0 {
		user.Position = model.Position
	}

	if len(model.SiteGroupName) > 0 {
		user.SiteGroupName = model.SiteGroupName
	}

	user.UpdatedAt = time.Now().UTC()
}

// ToUser will convert CreateUserModel to User domain model
func (model *CreateUserModel) ToUser() common.User {
	bytes, err := json.Marshal(&model)
	if err != nil {
		//TODO log error
		return common.User{}
	}

	user := common.User{}
	json.Unmarshal(bytes, &user)

	return user
}

//PrepareUserSearchQuery will prepare the query model
func PrepareUserSearchQuery(req *restful.Request) (*Query, error) {
	query := Query{
		PageNumber: 1,
		PageSize:   20,
		SortOrder:  -1,
	}

	// get query params and try parse and update model
	val := req.QueryParameter("pageNumber")
	if len(val) > 0 {
		i, err := strconv.Atoi(val)
		if err != nil {
			log.Errorf("Error occured during type convertion, error: %v", err)
			return nil, errors.CreateError(400, "invalid_data")
		}

		query.PageNumber = i
	}

	val = req.QueryParameter("pageSize")
	if len(val) > 0 {
		i, err := strconv.Atoi(val)
		if err != nil {
			log.Errorf("Error occured during type convertion, error: %v", err)
			return nil, errors.CreateError(400, "invalid_data")
		}

		query.PageSize = i
	}

	query.Role = req.QueryParameter("role")
	query.SortBy = req.QueryParameter("sortBy")
	query.Status = req.QueryParameter("status")
	query.Keyword = req.QueryParameter("keyword")

	val = req.QueryParameter("sortOrder")
	if len(val) > 0 {
		if val == "asc" {
			query.SortOrder = 1
		}
	}

	return &query, nil
}
