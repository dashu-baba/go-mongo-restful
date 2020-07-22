package utils

import (
	"strings"
	"time"

	"anacove.com/backend/common"
	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// HasRole checks weather user has role in claims.
func HasRole(req *restful.Request, roles ...string) bool {
	claims := GetClaims(req)

	for _, p := range claims.Permissions {
		for _, role := range roles {
			if strings.ToLower(p.Role) == strings.ToLower(role) {
				return true
			}
		}
	}

	return false
}

// CanAccessResource checks weather user has permission on specific resource.
func CanAccessResource(req *restful.Request, resource string, resourceID string) bool {
	claims := GetClaims(req)
	service := GetCommonService()

	for _, p := range claims.Permissions {
		return service.HasPermissions(p.Role, p.Scopes, resource, resourceID)
	}

	return false
}

// BearerAuth is used by all other endpoints to performan bearer token authorization
func BearerAuth(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	tokenHeader := req.HeaderParameter("Authorization")
	if len(tokenHeader) == 0 {
		log.Infof("Authorization token header not found")
		resp.WriteErrorString(401, "Not Authorized")
		return
	}

	splitted := strings.Split(tokenHeader, " ")
	if len(splitted) != 2 {
		log.Infof("Authorization token header not found")
		resp.WriteErrorString(401, "Not Authorized")
		return
	}

	// TODO: verify the token and authorize here
	token := splitted[1]
	account, err := GetCommonService().GetUserByToken(token)

	if err != nil {
		log.Errorf("Error occured during account retrival via token, error: %v", err)
		resp.WriteErrorString(401, "Not Authorized")
		return
	}

	if account.Status != common.Active {
		log.Infof("User is not active")
		resp.WriteErrorString(401, "Account is not active")
		return
	}

	claims := &common.Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return GenerateKey(account.ID.Hex()), nil
	})

	if err != nil {
		log.Errorf("Parsing jwt, error: %v", err)
		resp.WriteErrorString(401, "Not Authorized")
		return
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		log.Infof("Token expired")
		resp.WriteErrorString(401, "Token expired")
		return
	}

	// Set user id and claims in request attribute to access the whole lifetime of request
	req.SetAttribute(common.CurrentUserID, account.ID.Hex())
	req.SetAttribute(common.ClaimsKey, claims)
	chain.ProcessFilter(req, resp)
}

// GenerateKey create sart key array from salt string
func GenerateKey(salt string) []byte {
	return []byte(salt)
}

// parseBearerToken parses the bearer token from request header, this should only be called after an API is authorized.
func parseBearerToken(req *restful.Request) string {
	return strings.Split(req.HeaderParameter("Authorization"), " ")[1]
}

//GetClaims returns the claim object from token
func GetClaims(req *restful.Request) common.Claims {
	claimsObj := req.Attribute(common.ClaimsKey)
	claims, ok := claimsObj.(*common.Claims)

	//Try to convert to claims
	if !ok {
		return common.Claims{}
	}

	return *claims
}
