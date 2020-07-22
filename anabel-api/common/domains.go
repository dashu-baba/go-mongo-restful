package common

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

//User godoc
// @Summary The User entity.
type User struct {
	ID                     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Email                  string        `json:"email" bson:"email"`
	Token                  string        `json:"-" bson:"token"`
	ActivationCode         string        `json:"-" bson:"activationCode"`
	Status                 string        `json:"status" bson:"status"`
	FirstName              string        `json:"firstName" bson:"firstName"`
	FamilyName             string        `json:"familyName" bson:"familyName"`
	ProfileURL             string        `json:"profileUrl" bson:"profileUrl"`
	Position               string        `json:"position" bson:"position"`
	Phone                  string        `json:"phone" bson:"phone"`
	NotificationPreference string        `json:"notificationPreference" bson:"notificationPreference"`
	SiteID                 string        `json:"siteId" bson:"siteId"`
	SiteTagID              int           `json:"siteTagId" bson:"siteTagId"`
	SiteUserType           string        `json:"siteUserType" bson:"siteUserType"`
	AdminUserType          string        `json:"adminUserType" bson:"adminUserType"`
	SiteGroupName          string        `json:"siteUserGroup" bson:"siteUserGroup"`
	ClientID               string        `json:"clientId" bson:"clientId"`
	Password               string        `json:"-" bson:"password"`
	CreatedAt              time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt              time.Time     `json:"updatedAt" bson:"updatedAt"`
	LastLoginAt            time.Time     `json:"lastLoginAt" bson:"lastLoginAt"`
	Permission             []Permission  `json:"permissions" bson:"permissions"`
}

//Client godoc
// @Summary The Client entity.
type Client struct {
	ID             bson.ObjectId `json:"id" bson:"_id,omitempty"`
	UID            int64         `json:"uid" bson:"uid"`
	LogoURL        string        `json:"logoUrl" bson:"logoUrl"`
	Name           string        `json:"name" bson:"name"`
	Address        Address       `json:"address" bson:"address"`
	FullAddress    string        `json:"fullAddress" bson:"fullAddress"`
	BillingAddress Address       `json:"billingAddress" bson:"billingAddress"`
	NumberOfAlerts int           `json:"numberOfAlerts" bson:"numberOfAlerts"`
	NumberOfUsers  int           `json:"numberOfUsers" bson:"numberOfUsers"`
	NumberOfSites  int           `json:"numberOfSites" bson:"numberOfSites"`
	Status         string        `json:"status" bson:"status"`
	CreatedOn      time.Time     `json:"createdOdn" bson:"createdOdn"`
	UpdatedOn      time.Time     `json:"updatedOn" bson:"updatedOn"`
	Contacts       []string      `json:"contacts" bson:"contacts"`
	AdminUsers     []string      `json:"adminUsers" bson:"adminUsers"`
	Configuration  struct {
		FS             FS `json:"FS" bson:"FS"`
		TFS            FS `json:"TFS" bson:"TFS"`
		RequiredDevice []struct {
			Name        string `json:"name" bson:"name"`
			Description string `json:"description" bson:"description"`
			Amount      int    `json:"amount" bson:"amount"`
		} `json:"requiredDevice" bson:"requiredDevice"`
	} `json:"configuration" bson:"configuration"`
	Groups []struct {
		ID            string `json:"id" bson:"id"`
		Name          string `json:"name" bson:"name"`
		Enable        bool   `json:"enable" bson:"enable"`
		StaffAlert    bool   `json:"staffAlert" bson:"staffAlert"`
		Notifications bool   `json:"notifications" bson:"notifications"`
		SystemAlert   bool   `json:"systemAlert" bson:"systemAlert"`
	} `json:"groups" bson:"groups"`
}

//Address godoc
// @Summary The Address entity.
type Address struct {
	Line1  string `validate:"required" json:"line1" bson:"line1"`
	Line2  string `json:"line2" bson:"line2"`
	Zip    string `validate:"required" json:"zip" bson:"zip"`
	City   string `validate:"required" json:"city" bson:"city"`
	State  string `validate:"required" json:"state" bson:"state"`
	Phone1 string `validate:"required" json:"phone1" bson:"phone1"`
	Phone2 string `json:"phone2" bson:"phone2"`
}

//FS godoc
// @FS The Address entity.
type FS struct {
	SuspendClientAccess bool     `json:"suspendClientAccess" bson:"suspendClientAccess"`
	StaffAlert          bool     `json:"staffAlert" bson:"staffAlert"`
	TVTheftPrevention   Item     `json:"TVTheftPrevention" bson:"TVTheftPrevention"`
	AnalyticsLevel1     Item     `json:"analyticsLevel1" bson:"analyticsLevel1"`
	FutureModules       []string `json:"futureModules" bson:"futureModules"`
}

//Item godoc
// @FS The FS items.
type Item struct {
	Label string `json:"label" bson:"label"`
	Start string `json:"start" bson:"start"`
	End   string `json:"end" bson:"end"`
}

//Permission godoc
// @Summary The Permission entity.
type Permission struct {
	Role   string  `bson:"role"`
	Scopes []Scope `bson:"scopes"`
}

//Scope godoc
// @Summary The Scope entity.
type Scope struct {
	Resource []string `bson:"resources"`
	Ids      []string `bson:"ids"`
}
