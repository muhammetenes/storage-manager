package awss3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo"
	"main/config"
	"main/handlers"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type Handler struct{}

func (h Handler) ListBaseObjects(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]
	var result = new(handlers.ListObjectsResult)
	result.Bucket = handlers.Bucket{
		Name: bucket,
		Url:  bucket,
	}
	// Get objects
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		//MaxKeys:   awss3.Int64(100),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return c.Render(http.StatusOK, "album.html", result)
	}
	// Adding folders
	result.Folders = make([]handlers.Folder, len(resp.CommonPrefixes))
	for i, item := range resp.CommonPrefixes {
		result.Folders[i] = handlers.Folder{
			Name: *item.Prefix,
			Url:  c.Echo().URI(handlers.ListFolderObjects, bucket, strings.Replace(*item.Prefix, "/", ":", -1)),
		}
	}
	// Adding object count
	result.Count = len(resp.Contents)
	// Adding objects
	result.Objects = make([]handlers.Object, result.Count)
	for i, item := range resp.Contents {
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(*item.Key),
		})
		fileType := handlers.GetFileType(*item.Key)
		fileTypeIsValid := false
		for _, val := range handlers.ValidFileType {
			if fileType == val {
				fileTypeIsValid = true
				break
			}
		}
		urlStr, _ := req.Presign(15 * time.Minute)
		result.Objects[i] = handlers.Object{
			Name: *item.Key,
			Url:  urlStr,
			Type: fileTypeIsValid,
		}
	}
	return c.Render(http.StatusOK, "album.html", result)
}

func getPreviousUrl(f string, c echo.Context, b string) string {
	splitFolder := strings.Split(f, ":")
	folder := strings.Join(splitFolder[0:len(splitFolder)-2], "") + ":"
	if folder == ":" {
		return c.Echo().URI(handlers.ListBaseObjects, b)
	} else {
		return c.Echo().URI(handlers.ListFolderObjects, b, folder)
	}
}

func (h Handler) ListFolderObjects(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]
	folderKey := strings.Replace(c.ParamValues()[1], ":", "/", -1)
	var result = new(handlers.ListObjectsResult)
	result.Bucket = handlers.Bucket{
		Name:   bucket,
		Prefix: folderKey,
		Url:    bucket,
	}
	result.PreviousFolderUrl = getPreviousUrl(c.ParamValues()[1], c, bucket)
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
	result.Folders = make([]handlers.Folder, len(resp.CommonPrefixes))
	for i, item := range resp.CommonPrefixes {
		result.Folders[i] = handlers.Folder{
			Name: *item.Prefix,
			Url:  c.Echo().URI(handlers.ListFolderObjects, bucket, strings.Replace(*item.Prefix, "/", ":", -1)),
		}
	}

	// Adding object count
	// The first object in the folder is always itself
	result.Count = len(resp.Contents) - 1
	// Adding objects
	result.Objects = make([]handlers.Object, result.Count)
	// Used [:i] because the first object is the folder itself
	for i, item := range resp.Contents[1:] {
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(*item.Key),
		})
		fileType := handlers.GetFileType(*item.Key)
		fileTypeIsValid := false
		for _, val := range handlers.ValidFileType {
			if fileType == val {
				fileTypeIsValid = true
				break
			}
		}
		urlStr, _ := req.Presign(15 * time.Minute)

		result.Objects[i] = handlers.Object{
			Name: *item.Key,
			Url:  urlStr,
			Type: fileTypeIsValid,
		}
	}
	return c.Render(http.StatusOK, "album.html", result)
}

func (h Handler) ListBuckets(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	svc := s3.New(sess)
	resp, err := svc.ListBuckets(nil)
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		} else {
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: "Error"})
		}
	}
	var buckets handlers.ListBucketsResult
	buckets.Count = len(resp.Buckets)
	for _, item := range resp.Buckets {
		buckets.Buckets = append(buckets.Buckets, handlers.Bucket{
			Name: *item.Name,
			Url:  c.Echo().URI(handlers.ListBaseObjects, *item.Name, ""),
		})
	}
	return c.Render(http.StatusOK, "buckets.html", buckets)
}

func (h Handler) CreateBucket(c echo.Context) error {
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
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	return c.JSON(http.StatusOK, handlers.JsonResponse{Error: false, Message: "Success"})
}

func (h Handler) UploadFileToBucket(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	bucket := c.ParamValues()[0]
	folder_key := form.Value["folder_key_input"]
	files := form.File["file_input"]
	done := make(chan bool, len(files))
	for _, file := range files {
		// Source
		go func(file *multipart.FileHeader) {
			src, err := file.Open()
			if err != nil {
				if _, ok := err.(awserr.RequestFailure); ok {
					done <- false
					return
					//return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
				}
			}
			defer src.Close()

			// Copy
			uploader := s3manager.NewUploader(sess)
			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(folder_key[0] + file.Filename),
				Body:   src,
			})
			if err != nil {
				if _, ok := err.(awserr.RequestFailure); ok {
					done <- false
					return
					//return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
				}
			}
			done <- true
		}(file)
	}
	for i := 0; i < len(files); i++ {
		<-done
	}
	return c.JSON(http.StatusOK, handlers.JsonResponse{Error: false, Message: "Success"})
}

func (h Handler) DeleteBuckets(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	svc := s3.New(sess)
	_ = c.FormValue("buckets[]")
	buckets := c.Request().Form["buckets[]"]
	done := make(chan bool, len(buckets))
	for _, bucket := range buckets {
		go func() {
			_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				if _, ok := err.(awserr.RequestFailure); ok {
					done <- false
					return
					//return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
				}
			}
			err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				if _, ok := err.(awserr.RequestFailure); ok {
					done <- false
					return
					//return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
				}
			}
			done <- true
		}()
	}
	for i := 0; i < len(buckets); i++ {
		<-done
	}
	return c.JSON(http.StatusOK, handlers.JsonResponse{Error: false, Message: "Success"})
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

func (h Handler) DeleteObjects(c echo.Context) error {
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
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(keys[len(keys)-1]),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	return c.JSON(http.StatusOK, handlers.JsonResponse{Error: false, Message: "Objects Deleted"})
}
