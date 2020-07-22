package dummy

import (
	"encoding/json"
	"sync"
	"time"

	"anacove.com/backend/common"
	"anacove.com/backend/utils"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Service godoc
// defines all the dummy related crud operations
type Service struct {
}

// ServiceInstance Service instance
var ServiceInstance *Service

// ServiceMu mutex for dummy service
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
// perform operation to create a dummy
func (Service *Service) CreateUser(model User) error {

	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)

	bytes, _ := json.Marshal(&model)
	user := common.User{}
	json.Unmarshal(bytes, &user)

	// preparing data
	id := bson.NewObjectId()
	user.ID = id
	user.Status = common.Active
	user.CreatedAt = time.Now().UTC()
	user.ActivationCode = uuid.New().String()
	user.Password, _ = hashAndSalt(model.Password)

	_ = c.Insert(&user)

	return nil
}

// hashAndSalt takes a plain password and returns the hash of it
func hashAndSalt(pwd string) (string, error) {
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		log.Errorf("error hashing password: %v\n", err)
		return pwd, err
	}

	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}
