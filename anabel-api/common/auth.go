package common

import "github.com/dgrijalva/jwt-go"

//AuthPermissionScheme godoc
// Struct that defines fields of role, resource and identity
type AuthPermissionScheme struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Resource string `json:"resource"`
}

// Claims godoc
// Struct that will be encoded to a JWT
type Claims struct {
	ID          string       `json:"id"`
	Permissions []Permission `json:"permissions"`
	ClientID    string       `json:"clientId"`
	SiteID      string       `json:"siteId"`
	Email       string       `json:"email"`
	jwt.StandardClaims
}
