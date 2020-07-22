package security

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// User godoc
// This is the summary of user model definition
type User struct {
	ID          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Email       string        `json:"email" bson:"email"`
	Password    string        `json:"password,omitempty" bson:"password"`
	CreatedAt   time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt" bson:"updatedAt"`
	LastLoginAt time.Time     `json:"lastLoginAt" bson:"lastLoginAt"`
	Status      string        `json:"status" bson:"status"`
}

// UserConfirmationModel godoc
// This is the user confirmation request model definition
type UserConfirmationModel struct {
	Token                  string `validate:"required" json:"token" bson:"token"`
	SiteGroupName          string `json:"siteGroupName" bson:"siteGroupName"`
	Position               string `json:"position" bson:"position"`
	Password               string `validate:"required" json:"password,omitempty" bson:"password"`
	Phone                  string `json:"phone" bson:"phone"`
	NotificationPreference string `json:"notificationPreference" bson:"notificationPreference"`
}
