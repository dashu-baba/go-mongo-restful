package security

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"anacove.com/backend/common"
	"anacove.com/backend/config"
	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Service defines methods of security purpose
type Service struct {
}

// ServiceInstance defines the service instance
var ServiceInstance *Service

// ServiceMu provides the definition for service mutex
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

// comparePasswords compares the given password hash with the given plain password and
// tells the caller whether they match or not
func comparePasswords(hashedPwd string, plainPwd string) (bool, error) {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	return err == nil, err
}

// Login tries to perform login with the email and password
// returns user with token and expiry if succeeds
func (Service *Service) Login(email string, password string) (interface{}, error) {
	//validate input data
	if len(strings.TrimSpace(email)) == 0 || len(strings.TrimSpace(password)) == 0 {
		log.Infof("Invalid email and/or password")
		return nil, errors.CreateError(400, "invalid_credentials")
	}

	//Create session and connect to db
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)

	//Get user by email
	user := common.User{}
	err := c.Find(bson.M{"email": email}).One(&user)
	if err != nil {
		log.Errorf("cannot find the user with email: %s, error: %v\n", email, err)
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "not_found")
		}
		return nil, errors.CreateError(500, "user_find_error")
	}
	log.Infof("Get user by email success")

	//check password
	if ok, _ := comparePasswords(user.Password, password); !ok {
		log.Infof("password does not match")
		return nil, errors.CreateError(400, "invalid_password")
	}

	if user.Status != common.Active {
		log.Infof("inactive user")
		return nil, errors.CreateError(400, "account_not_active")
	}

	//Generating token
	user.LastLoginAt = time.Now().Truncate(time.Millisecond)
	expiary, token, err := generateToken(user)
	if err != nil {
		log.Errorf("error occurred during token generation: error: %v\n", err)
		return nil, errors.CreateErrorWithMsg(500, "token_generate_error", err.Error())
	}

	err = c.Update(bson.M{"email": email}, bson.M{"$set": bson.M{
		"lastLoginAt": user.LastLoginAt, "token": token, "updatedAt": user.LastLoginAt}})
	if err != nil {
		log.Errorf("error occurred during update: error: %v\n", err)
		return nil, errors.CreateErrorWithMsg(500, "update_user_error", err.Error())
	}

	log.Infof("Updated user token and lastLoginAt")

	// clear password in response
	user.Password = ""

	//Generate and sent response
	response := struct {
		User   common.User `json:"user"`
		Token  string      `json:"accessToken"`
		Expiry time.Time   `json:"accessTokenExpiredAt"`
	}{}
	response.User = user
	response.Token = *token
	response.Expiry = *expiary

	return response, nil
}

// Logout removes token from user tentity
func (Service *Service) Logout(id string) error {
	//Creating a db session and connection
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)

	//searching for user by id
	user := common.User{}
	objID := bson.ObjectIdHex(id)
	err := c.Find(bson.M{"_id": objID}).One(&user)

	if err != nil {
		log.Errorf("cannot find the user with id: %s, error: %v\n", id, err)
		if err == mgo.ErrNotFound {
			return errors.CreateError(404, "not_found")
		}
		return errors.CreateError(500, "user_find_error")
	}

	err = c.Update(bson.M{"_id": objID}, bson.M{"$set": bson.M{
		"token": "", "updatedAt": time.Now().UTC()}})

	if err != nil {
		log.Errorf("error occurred during update: error: %v\n", err)
		return errors.CreateErrorWithMsg(500, "update_user_error", err.Error())
	}

	log.Infof("Updated user token and updatedAt")

	return nil
}

// ForgotPassword checks user by email and send email with code
func (Service *Service) ForgotPassword(email string) error {
	if len(strings.TrimSpace(email)) == 0 {
		log.Infof("Invalid email")
		return errors.CreateError(400, "invalid_credentials")
	}

	//Creating a db session and connection
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)
	user := common.User{}
	err := c.Find(bson.M{"email": email}).One(&user)

	if err != nil {
		log.Errorf("cannot find the user with email: %s, error: %v\n", email, err)
		if err == mgo.ErrNotFound {
			return errors.CreateError(404, "not_found")
		}
		return errors.CreateError(500, "user_find_error")
	}

	activationCode := uuid.New().String()
	err = c.Update(bson.M{"email": email}, bson.M{"$set": bson.M{
		"activationCode": activationCode, "updatedAt": time.Now().UTC()}})

	if err != nil {
		log.Errorf("error occurred during update: error: %v\n", err)
		return errors.CreateErrorWithMsg(500, "update_user_error", err.Error())
	}

	utils.SendMailViaSES(user.Email, activationCode)

	return nil
}

// ChangePassword change user password
func (Service *Service) ChangePassword(oldPwd string, newPwd string, userID string) error {
	//Creating a db session and connection
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)

	user := common.User{}
	objID := bson.ObjectIdHex(userID)
	err := c.Find(bson.M{"_id": objID}).One(&user)

	if err != nil {
		log.Errorf("cannot find the user with id: %s, error: %v\n", userID, err)
		if err == mgo.ErrNotFound {
			return errors.CreateError(404, "not_found")
		}
		return errors.CreateError(500, "user_find_error")
	}

	if ok, _ := comparePasswords(user.Password, oldPwd); !ok {
		log.Infof("password does not match")
		return errors.CreateError(400, "invalid_password")
	}

	hash, err := hashAndSalt(newPwd)
	if err != nil {
		log.Errorf("Generating password hash throws error, error: %v\n", err)
		return errors.CreateError(500, "internal_error")
	}

	err = c.Update(bson.M{"_id": objID}, bson.M{"$set": bson.M{"password": hash,
		"activationCode": uuid.New().String(), "updatedAt": time.Now().UTC()}})

	if err != nil {
		log.Errorf("error occurred during update: error: %v\n", err)
		return errors.CreateErrorWithMsg(500, "update_user_error", err.Error())
	}

	return nil
}

// ConfirmUser activate user and update info
func (Service *Service) ConfirmUser(model UserConfirmationModel) error {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(common.UserCollection)
	user := common.User{}
	err := c.Find(bson.M{"activationCode": model.Token}).One(&user)

	if err != nil {
		log.Errorf("cannot find the user with token: %s, error: %v\n", model.Token, err)
		if err == mgo.ErrNotFound {
			return errors.CreateError(404, "not_found")
		}
		return errors.CreateError(500, "internal_error")
	}

	for _, p := range user.Permission {
		if p.Role == "CC" || p.Role == "SS" {
			return errors.CreateError(400, "bad_request")
		} else if p.Role != "GA" && len(model.SiteGroupName) > 0 {
			return errors.CreateError(400, "bad_request")
		}
	}

	if model.NotificationPreference == common.NotificationPhone && len(model.Phone) == 0 {
		log.Infof("Invalid notification preference %s", model.NotificationPreference)
		return errors.CreateError(400, "invalid_data")
	}

	user = model.ToUser(user)
	user.UpdatedAt = time.Now().UTC()
	user.Password, err = hashAndSalt(model.Password)
	user.ActivationCode = ""
	user.Token = ""
	if err != nil {
		log.Errorf("Generating password hash throws error, error: %v\n", err)
		return errors.CreateError(500, "internal_error")
	}

	err = c.Update(bson.M{"_id": user.ID}, user)
	if err != nil {
		log.Errorf("error occurred during update: error: %v\n", err)
		return errors.CreateErrorWithMsg(500, "update_user_error", err.Error())
	}

	return nil
}

//generateToken create token and returns it
func generateToken(user common.User) (*time.Time, *string, error) {
	//Create jwt key from identity
	jwtKey := utils.GenerateKey(user.ID.Hex())

	// Declare the expiration time of the token
	expirationPeriod, err := strconv.ParseInt(config.GetConfig().GetString("app.token_validation_period_in_minutes"), 10, 64)
	if err != nil {
		log.Errorf("error occurred during expiration time building: error: %v\n", err)
		return nil, nil, err
	}

	expirationTime := time.Now().UTC().Add(time.Duration(expirationPeriod) * time.Minute)

	// Create the JWT claims, which includes the user information and expiry time
	claims := &common.Claims{
		ID:          user.ID.Hex(),
		ClientID:    user.ClientID,
		Permissions: user.Permission,
		SiteID:      user.SiteID,
		Email:       user.Email,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string and return
	jwt, err := token.SignedString(jwtKey)
	return &expirationTime, &jwt, err
}
