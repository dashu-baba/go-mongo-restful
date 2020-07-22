package utils

import (
	"sync"

	"anacove.com/backend/common"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	log "github.com/sirupsen/logrus"
)

// CommonService godoc
// @summary CommonService defines all the common db operations
type CommonService struct {
}

// CommonServiceInstance CommonService instance
var CommonServiceInstance *CommonService

// CommonServiceMu mutex for common service
var CommonServiceMu sync.Mutex

// GetCommonService returns the singleton instance of the CommonService
func GetCommonService() *CommonService {
	CommonServiceMu.Lock()
	defer CommonServiceMu.Unlock()

	if CommonServiceInstance == nil {
		CommonServiceInstance = &CommonService{}
	}

	return CommonServiceInstance
}

// HasPermissions godoc
// @summary check user has permission to a resource or resource group
func (CommonService *CommonService) HasPermissions(role string, scopes []common.Scope, resource string, resourceID string) bool {
	session := NewDBSession()
	defer session.Close()
	userCollection := session.DB("").C(common.UserCollection)
	siteCollection := session.DB("").C(common.SiteCollection)
	clientCollection := session.DB("").C(common.ClientCollection)

	//Return for super admin user
	if role == "SA" {
		return true
	}

	log.Infof("Checking permission for non super admin user")
	//Check scope for other user role on multiple resource
	switch resource {
	case "user":
		{
			if isUserExistsInScope(scopes, resourceID, userCollection) {
				return true
			}
		}
	case "site":
		{
			if isSiteExistsInScope(scopes, resourceID, siteCollection) {
				return true
			}
		}
	case "client":
		{
			if isClientExistsInScope(scopes, resourceID, clientCollection) {
				return true
			}
		}
	}

	log.Infof("Permission not found")
	return false
}

// GetUserByToken return user by auth token
func (CommonService *CommonService) GetUserByToken(token string) (*common.User, error) {
	session := NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)
	user := common.User{}
	err := c.Find(bson.M{"token": token}).One(&user)

	if err != nil {
		log.Errorf("Failed to get user by token, error: %v", err)
		return nil, err
	}

	return &user, nil
}

// isUserExistsInScope checks user has permission to resource user
func isUserExistsInScope(scopes []common.Scope, userID string, collections *mgo.Collection) bool {
	objUserID := bson.ObjectIdHex(userID)
	for _, scope := range scopes {
		for _, resource := range scope.Resource {
			propName := ""
			if resource == "site" {
				propName = "siteId"
			} else {
				propName = "clientId"
			}

			count, err := collections.Find(bson.M{"_id": objUserID, propName: bson.M{"$in": scope.Ids}}).Count()
			if err == nil && count > 0 {
				log.Errorf("Failed to get permission, error: %v", err)
				return true
			}
		}
	}

	return false
}

// isSiteExistsInScope checks user has permission to resource site
func isSiteExistsInScope(scopes []common.Scope, siteID string, collections *mgo.Collection) bool {
	objSiteID := bson.ObjectIdHex(siteID)
	for _, scope := range scopes {
		for _, resource := range scope.Resource {
			if resource == "site" {
				objIds := []bson.ObjectId{}
				for _, id := range scope.Ids {
					objIds = append(objIds, bson.ObjectIdHex(id))
				}
				count, err := collections.Find(bson.M{"_id": objSiteID, "$and": []bson.M{bson.M{"_id": bson.M{"$in": objIds}}}}).Count()
				if err == nil && count > 0 {
					log.Errorf("Failed to get permission, error: %v", err)
					return true
				}
			} else {
				count, err := collections.Find(bson.M{"_id": objSiteID, "clientId": bson.M{"$in": scope.Ids}}).Count()
				if err == nil && count > 0 {
					log.Errorf("Failed to get permission, error: %v", err)
					return true
				}
			}
		}
	}

	return false
}

// isClientExistsInScope checks user has permission to resource client
func isClientExistsInScope(scopes []common.Scope, clientID string, collections *mgo.Collection) bool {
	objClientID := bson.ObjectIdHex(clientID)
	for _, scope := range scopes {
		for _, resource := range scope.Resource {
			if resource == "client" {
				objIds := []bson.ObjectId{}
				for _, id := range scope.Ids {
					objIds = append(objIds, bson.ObjectIdHex(id))
				}

				count, err := collections.Find(bson.M{"_id": objClientID, "$and": []bson.M{bson.M{"_id": bson.M{"$in": objIds}}}}).Count()
				if err == nil && count > 0 {
					log.Errorf("Failed to get permission, error: %v", err)
					return true
				}
			}
		}
	}

	return false
}
