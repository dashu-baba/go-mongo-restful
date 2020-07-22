package utils

import (
	"anacove.com/backend/config"
	"github.com/globalsign/mgo"
)

var gSession *mgo.Session = nil

// InitDB initializes the global database session
func InitDB() error {
	session, err := mgo.Dial(config.GetConfig().GetString("mongodb.url"))
	if err != nil {
		return err
	}

	session.SetMode(mgo.Monotonic, true)

	gSession = session

	return nil
}

// NewDBSession creates a new db session
func NewDBSession() *mgo.Session {
	return gSession.Copy()
}
