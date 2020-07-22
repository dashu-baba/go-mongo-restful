package main

import (
	"net/http"
	"os"

	"anacove.com/backend/rest/dummy"
	"anacove.com/backend/rest/user"

	"anacove.com/backend/rest/client"
	"anacove.com/backend/rest/file"

	"anacove.com/backend/config"
	"anacove.com/backend/rest/security"
	"anacove.com/backend/utils"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

func init() {
	// init the config
	config.Init()

	// setting up logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	logFile := config.GetConfig().GetString("log.file")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(os.Stdout)
	log.SetOutput(file)
	logLevel, err := log.ParseLevel(config.GetConfig().GetString("log.level"))
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(logLevel)
	log.SetReportCaller(true)
}

func main() {
	// init database
	err := utils.InitDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
		return
	}

	err = utils.InitAWS()
	if err != nil {
		log.Fatalf("failed to initialize aws: %v", err)
		return
	}

	// init routing
	wsContainer := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Path("/api/v1/").Consumes("multipart/form-data", restful.MIME_JSON).Produces(restful.MIME_JSON)

	// add all controllers
	security.SecurityController{}.AddRouters(ws)
	user.Controller{}.AddRouters(ws)
	client.Controller{}.AddRouters(ws)
	file.Controller{}.AddRouters(ws)
	dummy.Controller{}.AddRouters(ws)
	wsContainer.Add(ws)

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept", "Authorization"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)

	log.Debugf("Listening " + config.GetConfig().GetString("server.listen"))

	mux := http.NewServeMux()
	mux.Handle("/", wsContainer)
	log.Fatal(http.ListenAndServe(config.GetConfig().GetString("server.listen"), mux))
}
