package utils

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"anacove.com/backend/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var awsSession *session.Session = nil

// InitAWS initializes the global aws session
func InitAWS() error {
	session, err := session.NewSession(&aws.Config{
		Region: aws.String(config.GetConfig().GetString("aws.s3_region")),
		Credentials: credentials.NewStaticCredentials(
			config.GetConfig().GetString("aws.access_key_id"),
			config.GetConfig().GetString("aws.secret_access_key"),
			""),
	})

	if err != nil {
		log.Errorf("Failed to create aws session, error: %v", err)
		return err
	}

	awsSession = session

	return nil
}

// AwsSession returns the aws session
func AwsSession() *session.Session {
	return awsSession.Copy()
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func AddFileToS3(file multipart.File, fh multipart.FileHeader) (interface{}, error) {
	// Read the file content into a buffer
	var size int64 = fh.Size
	buffer := make([]byte, size)
	file.Read(buffer)
	fileName := uuid.New().String() + "-" + fh.Filename

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err := s3.New(AwsSession()).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(config.GetConfig().GetString("aws.s3_bucket")),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	res := struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}{}

	res.Name = fileName
	res.URL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", config.GetConfig().GetString("aws.s3_bucket"), config.GetConfig().GetString("aws.s3_region"), fileName)

	return res, err
}

// SendMailViaSES will send email to SES using some defined configuration
func SendMailViaSES(recipient string, code string) error {
	//Prepare the data
	url := config.GetConfig().GetString("app.forntend_url") + code
	body := strings.Replace(config.GetConfig().GetString("email.activation_body"), "{{url}}", url, -1)
	subject := config.GetConfig().GetString("email.activation_subject")

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(config.GetConfig().GetString("email.sender")),
	}

	// Attempt to send the email.
	_, err := ses.New(AwsSession()).SendEmail(input)

	return err
}
