package user

import (
	"strings"
	"sync"
	"time"

	"anacove.com/backend/common"
	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Service godoc
// defines all the user related crud operations
type Service struct {
}

// ServiceInstance Service instance
var ServiceInstance *Service

// ServiceMu mutex for user service
var ServiceMu sync.Mutex

// GetService returns the singleton instance of the Service
func GetService() *Service {
	ServiceMu.Lock()
	defer ServiceMu.Unlock()

	if ServiceInstance == nil {
		ServiceInstance = &Service{}
	}

	return ServiceInstance
}

// CreateUser godoc
// perform operation to create a user
func (Service *Service) CreateUser(model CreateUserModel) error {
	// preparing database connectivity
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)
	clientCollection := session.DB("").C(common.ClientCollection)
	siteCollection := session.DB("").C(common.SiteCollection)

	log.Infof("Getting user by email %s", model.Email)
	user := common.User{}
	err := c.Find(bson.M{"email": model.Email}).One(&user)
	if err == nil || err != mgo.ErrNotFound {
		log.Errorf("Error occured during getting user by email %s, error: %v", model.Email, err)
		return errors.CreateError(400, "duplicate_user")
	}

	// perform notification preference validation
	if model.NotificationPreference == common.NotificationPhone && len(model.Phone) == 0 {
		log.Infof("Invalid notification preference %s", model.NotificationPreference)
		return errors.CreateError(400, "Invalid notification preference")
	}

	//validate client id
	if len(model.ClientID) == 0 || !bson.IsObjectIdHex(model.ClientID) {
		log.Infof("Invalid clientid %s", model.ClientID)
		return errors.CreateError(400, "invalid_client")
	}

	objID := bson.ObjectIdHex(model.ClientID)
	client := common.Client{}
	err = clientCollection.Find(bson.M{"_id": objID, "status": bson.M{"$ne": common.Archive}}).One(&client)
	if err != nil {
		log.Errorf("Error occured while checking for client, error: %v", err)
		return errors.CreateError(400, "invalid_client")
	}

	// preparing data
	id := bson.NewObjectId()
	user = model.ToUser()
	user.ID = id
	user.Status = common.Active
	user.CreatedAt = time.Now().UTC()
	user.ActivationCode = uuid.New().String()

	bypassEmail := false

	// Calculate role and permissions
	if len(model.AdminUserType) != 0 {
		if model.AdminUserType == "GA" {
			sites := []struct {
				ID bson.ObjectId `json:"id" bson:"_id,omitempty"`
			}{}
			_ = siteCollection.Find(bson.M{"siteGroupName": model.SiteGroupName}).All(&sites)

			siteIds := []string{}
			for _, site := range sites {
				siteIds = append(siteIds, site.ID.Hex())
			}

			user.Permission = []common.Permission{common.Permission{
				Role: "GA",
				Scopes: []common.Scope{
					common.Scope{
						Resource: []string{"site"},
						Ids:      siteIds,
					},
				},
			}}
		} else {
			// checking number of admin
			adminCount, _ := c.Find(bson.M{"clientId": model.ClientID, "permissions.role": "CSA"}).Count()

			if adminCount == 3 {
				log.Errorf("CSA user limit for client %s reached", model.ClientID)
				return errors.CreateError(400, "admin user limit reached")
			}

			// update client admin list
			client.AdminUsers = append(client.AdminUsers, user.ID.Hex())

			user.Permission = []common.Permission{common.Permission{
				Role: "CSA",
				Scopes: []common.Scope{
					common.Scope{
						Resource: []string{"client"},
						Ids:      []string{model.ClientID},
					},
				},
			}}
		}

		log.Infof("Updating clients")
		client.NumberOfUsers++
		client.UpdatedOn = time.Now().UTC()

	} else if len(model.SiteUserType) != 0 {
		if len(model.SiteID) == 0 || !bson.IsObjectIdHex(model.SiteID) {
			log.Infof("Invalid siteid %s", model.SiteID)
			return errors.CreateError(400, "invalid_request_data")
		}

		objID := bson.ObjectIdHex(model.SiteID)
		_, err := siteCollection.Find(bson.M{"_id": objID}).Count()
		if err != nil {
			log.Errorf("Error occured while checking for site, error: %v", err)
			return errors.CreateError(400, "invalid_request_data")
		}

		user.Permission = []common.Permission{common.Permission{
			Role: model.SiteUserType,
			Scopes: []common.Scope{
				common.Scope{
					Resource: []string{"site"},
					Ids:      []string{model.SiteID},
				},
			},
		}}

		log.Infof("Updating number of site")
		err = siteCollection.Update(bson.M{"_id": objID}, bson.M{"$inc": bson.M{"numberOfUsers": 1}})
		if err != nil {
			log.Errorf("Error occured while update, error: %v", err)
			return errors.CreateError(500, "update_site_error")
		}
	} else {
		// checking number of admin
		contactCount, err := c.Find(bson.M{"clientId": model.ClientID, "permissions.role": "CC"}).Count()
		if err != nil {
			log.Errorf("Error occured while checking for CC user count, error: %v", err)
			return errors.CreateError(500, "server_error")
		}

		if contactCount == 10 {
			log.Errorf("CC user limit for client %s reached", model.ClientID)
			return errors.CreateError(400, "customer user limit reached")
		}

		// update client contact list
		client.Contacts = append(client.Contacts, user.ID.Hex())
		client.NumberOfUsers++
		client.UpdatedOn = time.Now().UTC()

		bypassEmail = true

		user.ActivationCode = ""
		user.Permission = []common.Permission{common.Permission{
			Role: "CC",
		}}
	}

	err = clientCollection.Update(bson.M{"_id": objID}, client)
	if err != nil {
		log.Errorf("Error occured while update, error: %v", err)
		return errors.CreateError(500, "update_client_error")
	}

	err = c.Insert(&user)
	if err != nil {
		log.Errorf("Error occured while insert, error: %v", err)
		return errors.CreateError(500, "create_user_error")
	}

	if bypassEmail {
		log.Infof("Sending ativation email")
		utils.SendMailViaSES(user.Email, user.ActivationCode)
	}

	return nil
}

// SearchUsers godoc
// search user by query and return list if succeeds
func (Service *Service) SearchUsers(query *Query, permissions []common.Permission, currentUserID string) (*common.PagedList, error) {
	//preparing db connection
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)

	//Main part of the query
	mainPart := bson.M{}

	//build query by user permission
	isSuperAdmin := false
	isGroupAdmin := false
	clientIds := []string{}
	siteIds := []string{}
	for _, p := range permissions {
		if p.Role == "SA" {
			isSuperAdmin = true
			break
		}
		if p.Role == "GA" {
			isGroupAdmin = true
			break
		}
		for _, scope := range p.Scopes {
			if utils.Contains(scope.Resource, "client") {
				clientIds = append(clientIds, scope.Ids...)
			} else if utils.Contains(scope.Resource, "site") {
				siteIds = append(siteIds, scope.Ids...)
			}
		}
	}

	if !isSuperAdmin {
		log.Infof("Building query for non super admin role")
		orParts := []bson.M{bson.M{"clientId": bson.M{"$in": clientIds}}, bson.M{"siteId": bson.M{"$in": siteIds}}}
		if isGroupAdmin {
			user := common.User{}
			_ = c.Find(bson.M{"_id": bson.ObjectIdHex(currentUserID)}).One(&user)

			log.Infof("Building query for non super admin role")
			orParts = append(orParts, bson.M{"clientId": user.ClientID, "permissions.role": "CC"})
		}

		mainPart["$or"] = orParts
	}

	//Building search part based on provided query data
	queryAndPart := []bson.M{}
	queryOrPart := []bson.M{}
	if len(query.Status) > 0 {
		queryAndPart = append(queryAndPart, bson.M{"status": query.Status})
	}

	if len(query.Role) > 0 {
		queryAndPart = append(queryAndPart, bson.M{"permissions.role": query.Role})
	}

	// applying fuzzy search
	if len(query.Keyword) > 0 {
		words := strings.Fields(query.Keyword)
		words = append(words, query.Keyword)

		for _, word := range words {
			part := bson.M{"email": bson.M{"$regex": bson.RegEx{Pattern: word, Options: "im"}}}
			queryOrPart = append(queryOrPart, part)
		}
	}

	//Create the main query
	log.Infof("Building main query body")
	dbQuery := mainPart
	queryAndPart = append(queryAndPart, mainPart)
	if len(queryAndPart) > 0 {
		dbQuery = bson.M{"$and": queryAndPart}
	}
	if len(queryOrPart) > 0 {
		queryAndPart = append(queryAndPart, bson.M{"$or": queryOrPart})
		dbQuery = bson.M{"$and": queryAndPart}
	}

	//Calculate total number of items
	count, err := c.Find(dbQuery).Count()

	if count == 0 && err != nil {
		log.Errorf("Error occured getting count, error: %v", err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "search_error")
	}

	//Preparing sorting part in query
	sortQuery := bson.M{"$sort": bson.M{"status": query.SortOrder}}
	switch query.SortBy {
	case SortByEmail:
		{
			sortQuery = bson.M{"$sort": bson.M{"email": query.SortOrder}}
		}
	}

	// executing query
	log.Infof("executing query")
	users := []common.User{}
	err = c.Pipe([]bson.M{bson.M{"$match": dbQuery}, bson.M{"$limit": query.PageSize * query.PageNumber}, sortQuery, bson.M{"$skip": query.PageSize * (query.PageNumber - 1)}}).All(&users)
	if err != nil {
		log.Errorf("Error occured executing search query, error: %v", err)
		if err != mgo.ErrNotFound {
			return nil, errors.CreateError(500, "search_error")
		}
	}

	response := common.PagedList{
		Items: users,
		Page:  query.PageNumber,
		Size:  query.PageSize,
		Total: count,
	}

	return &response, nil
}

// GetUser godoc
// Find user by id and return it is succeeds
func (Service *Service) GetUser(id string, permissions []common.Permission, currentUserID string) (*common.User, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)

	objID := bson.ObjectIdHex(id)
	user := common.User{}
	err := c.Find(bson.M{"_id": objID}).One(&user)

	if err != nil {
		log.Errorf("cannot find the user with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "get_by_id_error")
	}

	if id == currentUserID {
		return &user, nil
	}

	// Applying role based search
	isSuperAdmin := false
	isGroupAdmin := false
	clientIds := []string{}
	siteIds := []string{}
	for _, p := range permissions {
		if p.Role == "SA" {
			isSuperAdmin = true
			break
		}
		if p.Role == "GA" {
			isGroupAdmin = true
			break
		}
		for _, scope := range p.Scopes {
			if utils.Contains(scope.Resource, "client") {
				clientIds = append(clientIds, scope.Ids...)
			} else if utils.Contains(scope.Resource, "site") {
				siteIds = append(siteIds, scope.Ids...)
			}
		}
	}

	if !isSuperAdmin {
		log.Infof("Getiing possible data to check user")
		projection := []struct {
			ID bson.ObjectId `json:"id" bson:"_id,omitempty"`
		}{}

		orParts := []bson.M{bson.M{"clientId": bson.M{"$in": clientIds}},
			bson.M{"siteId": bson.M{"$in": siteIds}}}

		if isGroupAdmin {
			user := common.User{}
			_ = c.Find(bson.M{"_id": bson.ObjectIdHex(currentUserID)}).One(&user)

			log.Infof("Building query for non super admin role")
			orParts = append(orParts, bson.M{"clientId": user.ClientID, "permissions.role": "CC"})
		}

		err := c.Find(bson.M{"$or": orParts}).All(&projection)

		found := false
		for _, p := range projection {
			if p.ID == objID {
				found = true
				break
			}
		}

		if err != nil || !found {
			log.Errorf("cannot find the user with id: %s, error: %v\n", id, err)
			return nil, errors.CreateError(403, "forbidden")
		}

	}

	return &user, nil
}

// DeleteUser godoc
// @summary Delete user by id
func (Service *Service) DeleteUser(id string) error {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)
	clientCollection := session.DB("").C(common.ClientCollection)
	siteCollection := session.DB("").C(common.SiteCollection)

	user := common.User{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID}).One(&user)

	if err != nil {
		log.Errorf("cannot find the user with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return errors.CreateError(404, "not_found")
		}
		return errors.CreateError(500, "get_by_id_error")
	}

	_ = c.Remove(bson.M{"_id": objID})

	if len(user.AdminUserType) != 0 {
		if len(user.ClientID) == 0 || !bson.IsObjectIdHex(user.ClientID) {
			log.Infof("Invalid client data")
			return errors.CreateError(400, "invalid_data")
		}

		objID := bson.ObjectIdHex(user.ClientID)
		_, err := c.Find(bson.M{"_id": objID, "status": bson.M{"$ne": common.Archive}}).Count()
		if err != nil {
			log.Errorf("Error occurred during getting user, error: %v\n", err)
			return errors.CreateError(400, "invalid_data")
		}

		err = clientCollection.Update(bson.M{"_id": objID}, bson.M{"$inc": bson.M{"numberOfUsers": -1}})
		if err != nil {
			log.Errorf("Error occurred during update client, error: %v\n", err)
			return errors.CreateError(500, "update_client_error")
		}

		for _, p := range user.Permission {
			if p.Role == "CSA" {
				_ = clientCollection.Update(bson.M{"_id": objID}, bson.M{"$pull": bson.M{"adminUsers": user.ID.Hex()}})
			}
		}

	} else if len(user.SiteUserType) != 0 {
		if len(user.SiteID) == 0 || !bson.IsObjectIdHex(user.SiteID) {
			log.Infof("Invalid dite data %s", user.SiteID)
			return errors.CreateError(400, "invalid_data")
		}

		objID := bson.ObjectIdHex(user.SiteID)
		_, err := siteCollection.Find(bson.M{"_id": objID}).Count()
		if err != nil {
			log.Errorf("Error occurred during getting site, error: %v\n", err)
			return errors.CreateError(400, "invalid_data")
		}

		err = siteCollection.Update(bson.M{"_id": objID}, bson.M{"$inc": bson.M{"numberOfUsers": -1}})
		if err != nil {
			log.Errorf("Error occurred during update site, error: %v\n", err)
			return errors.CreateError(500, "update_site_error")
		}
	} else {
		objID := bson.ObjectIdHex(user.ClientID)
		for _, p := range user.Permission {
			if p.Role == "CC" {
				_ = clientCollection.Update(bson.M{"_id": objID}, bson.M{"$inc": bson.M{"numberOfUsers": -1}})
				_ = clientCollection.Update(bson.M{"_id": objID}, bson.M{"$pull": bson.M{"contacts": user.ID.Hex()}})
			}
		}
	}

	return nil
}

// UpdateUser godoc
// Update user by id retunrs it if succeeds
func (Service *Service) UpdateUser(id string, model UpdateUserModel, permissions []common.Permission, currentUserID string) (*common.User, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)

	if model.NotificationPreference == common.NotificationPhone && len(model.Phone) == 0 {
		return nil, errors.CreateError(400, "invalid_data")
	}

	user := common.User{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID}).One(&user)

	for _, p := range permissions {
		if p.Role == "GA" && id != currentUserID && len(model.SiteGroupName) > 0 {
			return nil, errors.CreateError(403, "forbidden_property_Access_error")
		}
		if !(p.Role == "SA" || p.Role == "AM") && (len(model.Email) > 0 || len(model.FirstName) > 0 || len(model.FamilyName) > 0) {
			return nil, errors.CreateError(403, "forbidden_property_Access_error")
		}
	}

	if err != nil {
		log.Errorf("cannot find the user with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "get_by_id_error")
	}

	model.ToUser(&user)
	err = c.Update(bson.M{"_id": objID}, user)

	if err != nil {
		log.Errorf("Error occurred during update, error: %v\n", err)
		return nil, errors.CreateError(500, "update_error")
	}

	return &user, nil
}
