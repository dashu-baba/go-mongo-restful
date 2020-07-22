package file

import (
	"anacove.com/backend/errors"
	"anacove.com/backend/utils"
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// Controller type
type Controller struct {
}

// AddRouters allows the endpoints defined in this controller to be added to router
func (controller Controller) AddRouters(ws *restful.WebService) *restful.WebService {
	ws.Route(ws.POST("/files").To(upload))
	return ws
}

// upload godoc
// read multiple files and store to aws s3
func upload(req *restful.Request, resp *restful.Response) {
	response := []interface{}{}
	r := req.Request
	r.ParseMultipartForm(32 << 20)
	fhs := r.MultipartForm.File["myfiles"]
	for _, fh := range fhs {
		file, _ := fh.Open()
		res, err := utils.AddFileToS3(file, *fh)

		if err != nil {
			log.Errorf("error occurred during file upload: error: %v\n", err)
			utils.WriteError(resp, errors.CreateError(500, err.Error()))
		}

		response = append(response, res)
	}
	resp.WriteEntity(response)
}
