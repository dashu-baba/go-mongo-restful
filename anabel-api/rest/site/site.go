package site

import (
	"time"

	"anacove.com/backend/models/common"
	"github.com/globalsign/mgo/bson"
)

//Site godoc
// @Summary The Site entity.
type Site struct {
	ID           string          `json:"id" bson:"_id,omitempty"`
	UID          int             `json:"uid" bson:"uid"`
	ClientID     string          `json:"clientId" bson:"clientId"`
	GroupAdminID int             `json:"groupAdminId" bson:"groupAdminId"`
	Name         string          `json:"name" bson:"name"`
	LogoURL      string          `json:"logoUrl" bson:"logoUrl"`
	Address      common.Address `json:"address" bson:"address"`
	FullAddress  string          `json:"fullAddress" bson:"fullAddress"`
	Details      struct {
		Building         int    `json:"building" bson:"building"`
		Room             int    `json:"room" bson:"room"`
		ManagementSystem string `json:"managementSystem" bson:"managementSystem"`
		Wifi             []struct {
			Ssid     string `json:"ssid" bson:"ssid"`
			Password string `json:"password" bson:"password"`
		} `json:"wifi" bson:"wifi"`
		FloorPlan []struct {
			URL  string `json:"url" bson:"url"`
			Name string `json:"name" bson:"name"`
		} `json:"floorPlan" bson:"floorPlan"`
	} `json:"details" bson:"details"`
	Rooms []struct {
		Room        int    `json:"room" bson:"room"`
		PhomeNumber string `json:"phomeNumber" bson:"phomeNumber"`
		Floor       string `json:"floor" bson:"floor"`
		Building    string `json:"building" bson:"building"`
	} `json:"rooms" bson:"rooms"`
	Team              []string `json:"team" bson:"team"`
	NotificationSetup []struct {
		ID            string `json:"id" bson:"id"`
		Name          string `json:"name" bson:"name"`
		StaffAlert    bool   `json:"staffAlert" bson:"staffAlert"`
		Notifications bool   `json:"notifications" bson:"notifications"`
		SystemAlert   bool   `json:"systemAlert" bson:"systemAlert"`
	} `json:"notificationSetup" bson:"notificationSetup"`
	Configuration struct {
		ID             string     `json:"id" bson:"id"`
		FS             common.FS `json:"FS" bson:"FS"`
		TFS            common.FS `json:"TFS" bson:"TFS"`
		RequiredDevice []struct {
			Name        string `json:"name" bson:"name"`
			Description string `json:"description" bson:"description"`
			Amount      int    `json:"amount" bson:"amount"`
		} `json:"requiredDevice" bson:"requiredDevice"`
	} `json:"configuration" bson:"configuration"`
	Options struct {
		Items []struct {
			Label  string `json:"label" bson:"label"`
			Value  int    `json:"value" bson:"value"`
			Unit   string `json:"unit" bson:"unit"`
			Enable bool   `json:"enable" bson:"enable"`
		} `json:"items" bson:"items"`
		TVTheftPreventionMessage struct {
			Value       string `json:"value" bson:"value"`
			Description string `json:"description" bson:"description"`
			AudioURL    string `json:"audioUrl" bson:"audioUrl"`
		} `json:"TVTheftPreventionMessage" bson:"TVTheftPreventionMessage"`
	} `json:"options" bson:"options"`
	CreatedOn  time.Time `json:"createdOdn" bson:"createdOdn"`
	UpdatedOn  time.Time `json:"updatedOn" bson:"updatedOn"`
	DetailTeam []struct {
		SiteUserGroup []string      `json:"siteUserGroup" bson:"siteUserGroup"`
		ID            bson.ObjectId `json:"id" bson:"id"`

		FirstName              string                         `json:"firstName" bson:"firstName"`
		FamilyName             string                         `json:"familyName" bson:"familyName"`
		ProfileURL             string                         `json:"profileUrl" bson:"profileUrl"`
		Position               string                         `json:"position" bson:"position"`
		Phone                  string                         `json:"phone" bson:"phone"`
		NotificationPreference common.NotificationPreference `json:"notificationPreference" bson:"notificationPreference"`
		SiteTagID              int                            `json:"siteTagId" bson:"siteTagId"`
		SiteUserType           string                         `json:"siteUserType" bson:"siteUserType"`
	} `json:"detailTeam" bson:"detailTeam"`
}
