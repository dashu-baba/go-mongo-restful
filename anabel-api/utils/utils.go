package utils

import (
	"encoding/hex"
	"math/rand"
	"regexp"
	"time"

	"anacove.com/backend/common"
	"anacove.com/backend/errors"
	"github.com/emicklei/go-restful"
	"github.com/globalsign/mgo/bson"
	"github.com/go-playground/validator"
)

// WriteError writes the given error to response
func WriteError(resp *restful.Response, err error) {
	status := 500
	switch err := err.(type) {
	case nil:
		// call succeeded, nothing to do
	case *errors.HttpError:
		status = err.StatusCode
	default:
		// unknown error
	}
	resp.WriteHeaderAndEntity(status, err)
}

// IsValidBsonID checks if the given string is a valid BsonObjectId
func IsValidBsonID(id string) bool {
	d, err := hex.DecodeString(id)
	if err != nil || len(d) != 12 {
		return false
	}
	return true
}

// IsValidEmail checks if the given email string is a valid email address
func IsValidEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

// GetUserID gets the current user id from request
func GetUserID(req *restful.Request) string {
	userIDInterface := req.Attribute(common.CurrentUserID)
	userID, ok := userIDInterface.(string)
	if !ok {
		return ""
	}
	return userID
}

// RangeIn generates a random number with the given range
func RangeIn(low, hi int) int {
	rand.Seed(time.Now().UnixNano())
	return low + rand.Intn(hi-low)
}

// DocumentsExistByIDs checkes whether the documents with the given ids all exist in the specified collection
// If any of the document doesn't exist in the DB false will be returned
func DocumentsExistByIDs(collectionName string, documentIDs []bson.ObjectId) bool {
	session := NewDBSession()
	defer session.Close()
	c := session.DB("").C(collectionName)
	count, err := c.Find(bson.M{
		"_id": bson.M{"$in": documentIDs},
	}).Count()

	if err != nil || count != len(documentIDs) {
		return false
	}

	return true
}

var validate *validator.Validate

// GetValidator returns single validator instance to validate request models
func GetValidator() *validator.Validate {
	if validate == nil {
		validate = validator.New()
	}

	return validate
}

// Contains check weather a string exists into an array of string
func Contains(arr []string, val string) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}
