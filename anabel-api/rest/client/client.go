package client

import (
	"anacove.com/backend/common"

	"github.com/globalsign/mgo/bson"
)

// Contact godoc
// provides definition for user contact
type Contact struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	FirstName  string        `json:"firstName" bson:"firstName"`
	FamilyName string        `json:"familyName" bson:"familyName"`
	ProfileURL string        `json:"profileUrl" bson:"profileUrl"`
	Position   string        `json:"position" bson:"position"`
	Phone      string        `json:"phone" bson:"phone"`
}

// CreateClientModel godoc
// define the request for create client
type CreateClientModel struct {
	Name    string `json:"name"`
	LogoURL string `json:"logoUrl"`
}

// UpdateResponseModel godoc
// define the response model for update client
type UpdateResponseModel struct {
	common.Client
	DetailContacts   []Contact `json:"detailContacts"`
	AdminDetailUsers []Contact `json:"adminDetailUsers"`
}

// UpdateRequestModel godoc
// define te reuest model for update client
type UpdateRequestModel struct {
	Address
	GroupDefinition
	Configuration
}

// Address godoc
// defines address part of client request model
type Address struct {
	LogoURL        string         `json:"logoUrl" bson:"logoUrl"`
	Name           string         `json:"name" bson:"name"`
	Address        common.Address `json:"address" bson:"address"`
	BillingAddress common.Address `json:"billingAddress" bson:"billingAddress"`
}

// GroupDefinition godoc
// defines group definition part of client request model
type GroupDefinition struct {
	Groups []struct {
		ID            string `json:"id" bson:"id"`
		Name          string `json:"name" bson:"name"`
		Enable        bool   `json:"enable" bson:"enable"`
		StaffAlert    bool   `json:"staffAlert" bson:"staffAlert"`
		Notifications bool   `json:"notifications" bson:"notifications"`
		SystemAlert   bool   `json:"systemAlert" bson:"systemAlert"`
	} `json:"groups" bson:"groups"`
}

// Configuration godoc
// define the configuration part of client model
type Configuration struct {
	Configuration struct {
		FS             common.FS        `json:"FS" bson:"FS"`
		TFS            common.FS        `json:"TFS" bson:"TFS"`
		RequiredDevice []RequiredDevice `json:"requiredDevice" bson:"requiredDevice"`
	} `json:"configuration" bson:"configuration"`
}

// RequiredDevice godoc
// defines the device informations
type RequiredDevice struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Amount      int    `json:"amount" bson:"amount"`
}

//Query godoc
// @Summary The Query entity.
type Query struct {
	PageNumber int
	PageSize   int
	SortBy     string
	SortOrder  int
	Status     string
	Keyword    string
}

const (
	// SortByUID godoc
	SortByUID = "uid"
	// SortByName godoc
	SortByName = "name"
	// SortByNumberOfSites godoc
	SortByNumberOfSites = "numberOfSites"
	// SortByNumberOfUsers godoc
	SortByNumberOfUsers = "numberOfUsers"
	// SortByNumberAlerts godoc
	SortByNumberAlerts = "numberAlerts"
)
