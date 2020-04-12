package handlers

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

type Folder struct {
	Name string
	Url  string
}

type ListObjectsResult struct {
	Bucket  Bucket
	Objects []Object
	Folders []Folder
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

func ListBaseObjects(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]
	var result = new(ListObjectsResult)
	result.Bucket = Bucket{
		Name: bucket,
		Url:  bucket,
	}
	// Get objects
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		//MaxKeys:   aws.Int64(100),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return c.Render(http.StatusOK, "album.html", result)
	}
	// Adding folders
	result.Folders = make([]Folder, len(resp.CommonPrefixes))
	for i, item := range resp.CommonPrefixes {
		result.Folders[i] = Folder{
			Name: *item.Prefix,
			Url:  c.Echo().URI(ListFolderObjects, bucket, strings.Replace(*item.Prefix, "/", ":", -1)),
		}
	}
	// Adding object count
	result.Count = len(resp.Contents)
	// Adding objects
	result.Objects = make([]Object, result.Count)
	for i, item := range resp.Contents {
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
		result.Objects[i] = Object{
			Name: *item.Key,
			Url:  urlStr,
			Type: fileTypeIsValid,
		}
	}
	return c.Render(http.StatusOK, "album.html", result)
}

func ListFolderObjects(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]
	folderKey := strings.Replace(c.ParamValues()[1], ":", "/", -1)
	var result = new(ListObjectsResult)
	result.Bucket = Bucket{
		Name: bucket + "/" + folderKey,
		Url:  bucket,
	}

	// Get objects
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		MaxKeys:   aws.Int64(100),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(folderKey),
	})
	if err != nil {
		return c.Render(http.StatusOK, "album.html", result)
	}

	// Adding folders
	result.Folders = make([]Folder, len(resp.CommonPrefixes))
	for i, item := range resp.CommonPrefixes {
		result.Folders[i] = Folder{
			Name: *item.Prefix,
			Url:  c.Echo().URI(ListFolderObjects, bucket, strings.Replace(*item.Prefix, "/", ":", -1)),
		}
	}

	// Adding object count
	// The first object in the folder is always itself
	if len(resp.Contents) > 0 {
		result.Count = len(resp.Contents) - 1
	}
	// Adding objects
	result.Objects = make([]Object, result.Count)
	for i, item := range resp.Contents {
		if i == 0 {
			continue
		}
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
		// Used [i - 1] because the first object is the folder itself
		result.Objects[i-1] = Object{
			Name: *item.Key,
			Url:  urlStr,
			Type: fileTypeIsValid,
		}
	}
	return c.Render(http.StatusOK, "album.html", result)
}

func ListBuckets(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
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
			Url:  c.Echo().URI(ListBaseObjects, *item.Name, ""),
		})
	}
	return c.Render(http.StatusOK, "buckets.html", buckets)
}

func CreateBucket(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
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
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
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
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
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
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
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