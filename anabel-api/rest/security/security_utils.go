package security

import (
	"anacove.com/backend/common"
)

//ToUser Convert UserConfirmationModel to User domain model
func (model *UserConfirmationModel) ToUser(user common.User) common.User {
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

	return user
}
