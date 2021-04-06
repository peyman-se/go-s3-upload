package s3_upload

import (
	"bytes"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var accessKeyID string
var secretAccessKey string
var region string

type S3Upload struct {
	BucketName string
	LocalFilePath string
	Destination string
	IsPublic bool
}



//GetEnvWithKey : get env value
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}

func connectToAws() *session.Session {
	accessKeyID = GetEnvWithKey("AWS_ACCESS_KEY_ID")
	secretAccessKey = GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
	region = GetEnvWithKey("AWS_REGION")

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(region),
			Credentials: credentials.NewStaticCredentials(
				accessKeyID,
				secretAccessKey,
				"", // a token will be created when the session it's used.
			),
		},
	)

	if err != nil {
		panic("cannot connect to AWS")
	}

	return sess
}


func (s *S3Upload) ToBucket(bucket string) *S3Upload {
	s.BucketName = bucket
	return s
}

func (s *S3Upload) FromLocalPath(path string) *S3Upload {
	s.LocalFilePath = path
	return s
}

func (s *S3Upload) MakePublic() *S3Upload {
	s.IsPublic = true
	return s
}

func (s *S3Upload) SaveTo(destination string) (string, error) {
	s.Destination = destination
	sess := connectToAws()

	uploader := s3manager.NewUploader(sess)
	bucketName := (map[bool]string{true: GetEnvWithKey("BUCKET_NAME"), false: s.BucketName})[s.BucketName == ""]

	file, err := os.Open(s.LocalFilePath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    fileInfo, _ := file.Stat()
    var fileSize int64 = fileInfo.Size()
    fileBuffer := make([]byte, fileSize)

    file.Read(fileBuffer)

	uploadInput := s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s.Destination),
		Body:   bytes.NewReader(fileBuffer),
	}

	if s.IsPublic {
		uploadInput.ACL = aws.String("public-read")
	}

	//upload to the s3 bucket
	_, e := uploader.Upload(&uploadInput)

	if e != nil {
		log.Panic(e)
		return "", errors.New("error uploading to s3")
	}

	filePath := "https://" + bucketName + "." + "s3-" + region + ".amazonaws.com/" + s.Destination

	return filePath, nil
}