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
	Count   int
}

type ListBucketsResult struct {
	Buckets []Bucket
	Count   int
}

func getFileType(fileName string) string {
	result := strings.Split(fileName, ".")
	return result[len(result)-1]
}

var validFileType = [...]string{"png", "PNG", "Png", "Jpeg", "JPEG", "Jpg", "JPG", "jpg", "jpeg"}

func ListObjects(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
		Region:      aws.String(config.AwsRegion),
	})
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]
	var result = new(ListObjectsResult)
	result.Bucket = Bucket{
		Name: bucket,
		Url:  bucket,
	}

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil {
		return c.Render(http.StatusOK, "album.html", result)
		//if awsErr, ok := err.(awserr.RequestFailure); ok {
		//	return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
		//} else {
		//	return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: "Error"})
		//}
	}

	result.Count = len(resp.Contents)
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
				break
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
	var buckets ListBucketsResult
	buckets.Count = len(resp.Buckets)
	for _, item := range resp.Buckets {
		buckets.Buckets = append(buckets.Buckets, Bucket{
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

func DeleteBuckets(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
		Region:      aws.String(config.AwsRegion),
	})
	svc := s3.New(sess)
	_ = c.FormValue("buckets[]")
	buckets := c.Request().Form["buckets[]"]
	for _, bucket := range buckets {
		_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok {
				return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
			}
		}
		err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok {
				return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
			}
		}
	}
	return c.JSON(http.StatusOK, JsonResponse{Error: false, Message: "Success"})
}

func getObjectsToDelete(keys []string) []*s3.ObjectIdentifier {
	var objects []*s3.ObjectIdentifier
	for _, key := range keys {
		objects = append(objects, &s3.ObjectIdentifier{
			Key: aws.String(key),
		})
	}
	return objects
}

func DeleteObjects(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
		Region:      aws.String(config.AwsRegion),
	})
	_ = c.FormValue("keys[]")
	keys := c.Request().Form["keys[]"]
	otd := getObjectsToDelete(keys)
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]
	_, err = svc.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: otd,
		},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(keys[len(keys)-1]),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	return c.JSON(http.StatusOK, JsonResponse{Error: false, Message: "Objects Deleted"})
}
