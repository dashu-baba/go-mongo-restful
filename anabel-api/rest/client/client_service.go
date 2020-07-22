package client

import (
	"sync"
	"time"

	"anacove.com/backend/common"
	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	log "github.com/sirupsen/logrus"
)

// Service godoc
// @summary define Service type
type Service struct {
}

// ClientServiceInstance clientservice instance
var ClientServiceInstance *Service

// ClientServiceMu mutex for client service
var ClientServiceMu sync.Mutex

// GetClientService returns the singleton instance of the ClientService
func GetClientService() *Service {
	ClientServiceMu.Lock()
	defer ClientServiceMu.Unlock()

	if ClientServiceInstance == nil {
		ClientServiceInstance = &Service{}
	}

	return ClientServiceInstance
}

// CreateClient godoc
// @summary create client
func (ClientService *Service) CreateClient(model CreateClientModel) (*common.Client, error) {
	//Create session and connect to db
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.ClientCollection)

	//Convert request model to db model
	client := common.Client{}
	id := bson.NewObjectId()
	client = model.toClient()
	client.ID = id
	client.CreatedOn = time.Now().UTC()
	client.UID = time.Now().Unix()
	client.Status = common.Active

	err := c.Insert(&client)
	if err != nil {
		log.Errorf("error occured during insert to client info: error: %v\n", err)
		return nil, errors.CreateError(500, "create_client_error")
	}
	log.Infof("client created")

	return &client, nil
}

// SearchClients godoc
// @summary try to create client
func (ClientService *Service) SearchClients(query *Query, permissions []common.Permission) (*common.PagedList, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.ClientCollection)

	//Main part of the query check not archive clients
	mainPart := bson.M{
		"status": bson.M{
			"$ne": common.Archive,
		},
	}

	// building permission based query
	clientIds := []bson.ObjectId{}
	isSuperAdmin := false
	for _, p := range permissions {
		if p.Role == "SA" {
			isSuperAdmin = true
			break
		}
		for _, scope := range p.Scopes {
			if utils.Contains(scope.Resource, "client") {
				for _, id := range scope.Ids {
					objID := bson.ObjectIdHex(id)
					clientIds = append(clientIds, objID)
				}
			}
		}
	}

	if !isSuperAdmin {
		//check clients within scopes
		mainPart["$and"] = []bson.M{bson.M{"_id": bson.M{"$in": clientIds}}}
	}

	//Building search part based on provided data
	queryAndPart := []bson.M{}
	queryOrPart := []bson.M{}
	if len(query.Status) > 0 {
		queryAndPart = append(queryOrPart, bson.M{"status": query.Status})
	}

	if len(query.Keyword) > 0 {
		queryOrPart = append(queryOrPart,
			bson.M{"uid": bson.M{"$regex": bson.RegEx{Pattern: query.Keyword, Options: "im"}}},
			bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: query.Keyword, Options: "im"}}},
			bson.M{"fullAddress": bson.M{"$regex": bson.RegEx{Pattern: query.Keyword, Options: "im"}}})
	}

	//Create the main query
	dbQuery := mainPart
	queryAndPart = append(queryAndPart, mainPart)
	if len(queryAndPart) > 0 {
		dbQuery = bson.M{"$and": queryAndPart}
	}
	if len(queryOrPart) > 0 {
		queryAndPart = append(queryAndPart, bson.M{"$or": queryOrPart})
		dbQuery = bson.M{"$and": queryAndPart}
	}

	log.Infof("query build complete")

	//Calculate total number of items
	count, err := c.Find(dbQuery).Count()
	if count == 0 && err != nil {
		log.Errorf("error occured during getting count: error: %v\n", err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "query_execute_error")
	}

	//Preparing sorting part in query
	sortQuery := bson.M{"$sort": bson.M{"uid": query.SortOrder}}
	switch query.SortBy {
	case SortByName:
		{
			sortQuery = bson.M{"$sort": bson.M{"name": query.SortOrder}}
		}
	case SortByNumberAlerts:
		{
			sortQuery = bson.M{"$sort": bson.M{"numberOfAlerts": query.SortOrder}}
		}
	case SortByNumberOfSites:
		{
			sortQuery = bson.M{"$sort": bson.M{"numberOfSites": query.SortOrder}}
		}
	case SortByNumberOfUsers:
		{
			sortQuery = bson.M{"$sort": bson.M{"numberOfUsers": query.SortOrder}}
		}
	}

	clients := []common.Client{}
	err = c.Pipe([]bson.M{bson.M{"$match": dbQuery}, bson.M{"$limit": query.PageSize * query.PageNumber}, sortQuery, bson.M{"$skip": query.PageSize * (query.PageNumber - 1)}}).All(&clients)
	if err != nil {
		log.Errorf("error occured during perform search: error: %v\n", err)
		if err != mgo.ErrNotFound {
			return nil, errors.CreateError(500, "search_error")
		}
	}

	response := common.PagedList{
		Items: clients,
		Page:  query.PageNumber,
		Size:  query.PageSize,
		Total: count,
	}

	return &response, nil
}

// GetClient godoc
// @summary Get client by id
func (ClientService *Service) GetClient(id string) (*common.Client, error) {
	// preparing db connectivity and session
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.ClientCollection)

	// getting client by id
	client := common.Client{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID, "status": bson.M{"$ne": common.Archive}}).One(&client)

	if err != nil {
		log.Errorf("cannot find the client with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "get_client_error")
	}

	return &client, nil
}

// DeleteClient godoc
// @summary Delete client by id
func (ClientService *Service) DeleteClient(id string) error {
	// preparing db connectivity and session
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.ClientCollection)
	userCollection := session.DB("").C(common.UserCollection)
	siteCollection := session.DB("").C(common.SiteCollection)

	// getting client by id
	client := common.Client{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID, "status": bson.M{"$ne": common.Archive}}).One(&client)

	if err != nil {
		log.Errorf("cannot find the client with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return errors.CreateError(404, "not_found")
		}
		return errors.CreateError(500, "get_client_error")
	}

	// removing all users by client
	_, err = userCollection.RemoveAll(bson.M{"clientId": objID.Hex()})
	if err != nil {
		log.Errorf("error occurred during remove user, error: %v\n", err)
		return errors.CreateError(500, "remove_user_error")
	}

	// removinf all site by client
	_, err = siteCollection.RemoveAll(bson.M{"clientId": objID.Hex()})
	if err != nil {
		log.Errorf("error occurred during remove site, error: %v\n", err)
		return errors.CreateError(500, "remove_site_error")
	}

	// removing client
	err = c.Remove(bson.M{"_id": objID})
	if err != nil {
		log.Errorf("error occurred during remove client, error: %v\n", err)
		return errors.CreateError(500, "remove_client_error")
	}

	return nil
}

// ArchiveClient godoc
// @summary Archive client info by id
func (ClientService *Service) ArchiveClient(id string) error {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.ClientCollection)
	userCollection := session.DB("").C(common.UserCollection)

	// getting client by id
	client := common.Client{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID, "status": bson.M{"$ne": common.Archive}}).One(&client)

	if err != nil {
		log.Errorf("cannot find the client with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return errors.CreateError(404, "not_found")
		}
		return errors.CreateError(500, "get_client_error")
	}

	// updating user by client with clear token
	err = userCollection.Update(bson.M{"clientId": objID},
		bson.M{"$set": bson.M{"status": common.Inactive, "token": "", "updatedAt": time.Now().UTC()}})
	if err != nil && err != mgo.ErrNotFound {
		log.Errorf("error occurred during update user, error: %v\n", err)
		return errors.CreateError(500, "update_user_error")
	}

	// archiving client
	err = c.Update(bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": common.Archive, "updatedOn": time.Now().UTC()}})
	if err != nil {
		log.Errorf("error occurred during update client, error: %v\n", err)
		return errors.CreateError(500, "update_client_error")
	}

	return nil
}

// GetSiteGroup godoc
// @summary Get site group by id
func (ClientService *Service) GetSiteGroup(id string) (interface{}, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.ClientCollection)
	userCollection := session.DB("").C(common.UserCollection)

	// getting client by id
	client := common.Client{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID, "status": bson.M{"$ne": common.Archive}}).One(&client)

	if err != nil {
		log.Errorf("cannot find the client with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "login_error")
	}

	// Getting GA people by client
	user := struct {
		ID            bson.ObjectId `json:"id" bson:"_id,omitempty"`
		SiteGroupName string        `json:"siteUserGroup" bson:"siteUserGroup"`
	}{}
	err = userCollection.Find(bson.M{"clientId": id, "permissions.role": "GA"}).One(&user)
	if err != nil && err != mgo.ErrNotFound {
		log.Errorf("error occurred during finding user, error: %v\n", err)
		return nil, errors.CreateError(500, "login_error")
	}

	return &user, nil
}

// UpdateClient godoc
// @summary Update client by id
func (ClientService *Service) UpdateClient(id string, model UpdateRequestModel, permissions []common.Permission) (*UpdateResponseModel, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.ClientCollection)
	userCollection := session.DB("").C(common.UserCollection)

	// Getting client by id
	client := common.Client{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID, "status": bson.M{"$ne": common.Archive}}).One(&client)

	if err != nil {
		log.Errorf("cannot find the client with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "internal_error")
	}

	client, err = model.ToClient(client, permissions)
	if err != nil {
		log.Errorf("error occurred during model conversion, error: %v\n", err)
		return nil, err
	}

	client.UpdatedOn = time.Now().UTC()
	err = c.Update(bson.M{"_id": objID}, client)
	if err != nil {
		log.Errorf("error occurred during update client, error: %v\n", err)
		return nil, err
	}

	// create response
	response := ToUpdateResponseModel(UpdateResponseModel{}, client)

	// getting contact users by clients
	objIds := []bson.ObjectId{}
	for _, id := range response.Contacts {
		objIds = append(objIds, bson.ObjectIdHex(id))
	}
	userCollection.Find(bson.M{"_id": bson.M{"$in": objIds}}).All(&response.DetailContacts)

	// getting admin users by clients
	objIds = []bson.ObjectId{}
	for _, id := range response.AdminUsers {
		objIds = append(objIds, bson.ObjectIdHex(id))
	}
	userCollection.Find(bson.M{"_id": bson.M{"$in": objIds}}).All(&response.AdminDetailUsers)

	return &response, nil
}
