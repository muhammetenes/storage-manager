package views

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo"
	"main/config"
	"net/http"
	"strings"
	"time"
)

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type Bucket struct {
	Name string
	Url  string
}

type Object struct {
	Name string
	Url  string
	Type bool
}

type ListObjectsResult struct {
	Bucket  Bucket
	Objects []Object
}

func getFileType(fileName string) string {
	result := strings.Split(fileName, ".")
	return result[len(result)-1]
}

var validFileType = [...]string{"png", "PNG", "Png", "Jpeg", "JPEG", "Jpg", "JPG", "jpeg"}

func ListObjects(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
		Region:      aws.String(config.AwsRegion),
	})
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
		} else {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: "Error"})
		}
	}
	var result = new(ListObjectsResult)
	result.Bucket = Bucket{
		Name: bucket,
		Url:  bucket,
	}
	for _, item := range resp.Contents {
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(*item.Key),
		})
		fileType := getFileType(*item.Key)
		fileTypeIsValid := false
		for _, val := range validFileType {
			if fileType == val {
				fileTypeIsValid = true
			}
		}
		urlStr, _ := req.Presign(15 * time.Minute)
		result.Objects = append(result.Objects, Object{
			Name: *item.Key,
			Url:  urlStr,
			Type: fileTypeIsValid,
		})
	}
	return c.Render(http.StatusOK, "album.html", result)
}

func ListBuckets(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
		Region:      aws.String(config.AwsRegion),
	})
	svc := s3.New(sess)
	resp, err := svc.ListBuckets(nil)
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
		} else {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: "Error"})
		}
	}
	var buckets []Bucket
	for _, item := range resp.Buckets {
		buckets = append(buckets, Bucket{
			Name: *item.Name,
			Url:  c.Echo().URI(ListObjects, *item.Name),
		})
	}
	return c.Render(http.StatusOK, "buckets.html", buckets)
}

func CreateBucket(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
		Region:      aws.String(config.AwsRegion),
	})
	bucketName := c.FormValue("bucket_name")
	svc := s3.New(sess)
	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	return c.JSON(http.StatusOK, JsonResponse{Error: false, Message: "Success"})
}

func UploadFileToBucket(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
		Region:      aws.String(config.AwsRegion),
	})
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	bucket := c.ParamValues()[0]
	files := form.File["file_input"]
	for _, file := range files {
		// Source
		src, err := file.Open()
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok {
				return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
			}
		}
		defer src.Close()

		// Copy
		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(file.Filename),
			Body:   src,
		})
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok {
				return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
			}
		}
	}
	return c.JSON(http.StatusOK, JsonResponse{Error: false, Message: "Success"})
}
