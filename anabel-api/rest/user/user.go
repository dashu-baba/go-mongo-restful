package user

// CreateUserModel godoc
// This is the user create request model definition
type CreateUserModel struct {
	Email                  string   `validate:"required" json:"email"`
	FirstName              string   `json:"firstName"`
	FamilyName             string   `json:"familyName"`
	ProfileURL             string   `json:"profileUrl"`
	Position               string   `json:"position"`
	Phone                  string   `json:"phone"`
	NotificationPreference string   `json:"notificationPreference"`
	UserGroups             []string `json:"userGroups"`
	SiteID                 string   `json:"siteId"`
	SiteTagID              int      `json:"siteTagId"`
	SiteUserType           string   `json:"siteUserType"`
	AdminUserType          string   `json:"adminUserType"`
	SiteGroupName          string   `json:"siteGroupName"`
	ClientID               string   `json:"clientId"`
}

// UpdateUserModel godoc
// This is the user update request model definition
type UpdateUserModel struct {
	Email                  string `json:"email"`
	FirstName              string `json:"firstName"`
	FamilyName             string `json:"familyName"`
	Position               string `json:"position"`
	Phone                  string `json:"phone"`
	NotificationPreference string `json:"notificationPreference"`
	SiteGroupName          string `json:"siteGroupName"`
}

// Query godoc
// This is the query request model definition
type Query struct {
	PageNumber int
	PageSize   int
	SortBy     string
	SortOrder  int
	Role       string
	Status     string
	Keyword    string
}

const (
	// SortByStatus godoc
	SortByStatus = "status"
	// SortByUsername godoc
	SortByUsername = "username"
	// SortByEmail godoc
	SortByEmail = "email"
	// SortByRole godoc
	SortByRole = "role"
	// SortOrderAsc godoc
	SortOrderAsc = "asc"
	// SortOrderDesc godoc
	SortOrderDesc = "desc"
)
