package awss3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gammazero/workerpool"
	"github.com/labstack/echo"
	"main/config"
	"main/handlers"
	"main/handlers/base_handlers"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Handler struct{}

// Aws Regions
var AwsRegions = []string{
	"eu-central-1",
	"eu-west-1",
	"eu-west-2",
	"eu-south-1",
	"eu-west-3",
	"eu-north-1",
	"me-south-1",
	"sa-east-1",
	"us-east-2",
	"us-east-1",
	"us-west-1",
	"us-west-2",
	"af-south-1",
	"ap-east-1",
	"ap-south-1",
	"ap-northeast-3",
	"ap-northeast-2",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-northeast-1",
	"ca-central-1",
	"cn-north-1",
	"cn-northwest-1",
}

const maxKeys = 100

const workerNum = 15

func getSession() *session.Session {
	sess, _ := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.Conf.AwsConfig.AwsId, config.Conf.AwsConfig.AwsSecretKey, ""),
		Region:      aws.String(config.Conf.AwsConfig.AwsRegion),
	})
	return sess
}

func (h Handler) ListBaseObjects(c echo.Context) error {
	svc := s3.New(getSession())
	bucket := c.ParamValues()[0]
	var result = new(handlers.ListObjectsResult)
	result.Bucket = handlers.Bucket{
		Name: bucket,
		Url:  bucket,
	}
	// Get objects
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		MaxKeys:   aws.Int64(maxKeys),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		// Region Control
		if _, ok := err.(awserr.RequestFailure); ok {
			region, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{
				Bucket: aws.String(bucket),
			})
			config.Conf.UpdateAwsRegion(*region.LocationConstraint)
			svc = s3.New(getSession())
			resp, err = svc.ListObjectsV2(&s3.ListObjectsV2Input{
				Bucket:    aws.String(bucket),
				MaxKeys:   aws.Int64(maxKeys),
				Delimiter: aws.String("/"),
			})
			if err != nil {
				return c.Render(http.StatusOK, "album.html", result)
			}
		} else {
			return c.Render(http.StatusOK, "album.html", result)
		}
	}
	// Adding folders
	result.Folders = make([]handlers.Folder, len(resp.CommonPrefixes))
	for i, item := range resp.CommonPrefixes {
		result.Folders[i] = handlers.Folder{
			Name: *item.Prefix,
			Url:  c.Echo().URI(base_handlers.ListFolderObjects, bucket, strings.Replace(*item.Prefix, "/", ":", -1)),
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
			Name:    *item.Key,
			Url:     urlStr,
			Type:    fileType,
			IsValid: fileTypeIsValid,
		}
	}
	return c.Render(http.StatusOK, "album.html", result)
}

func getPreviousUrl(f string, c echo.Context, b string) string {
	splitFolder := strings.Split(f, ":")
	folder := strings.Join(splitFolder[0:len(splitFolder)-2], ":") + ":"
	if folder == ":" {
		return c.Echo().URI(base_handlers.ListBaseObjects, b)
	} else {
		return c.Echo().URI(base_handlers.ListFolderObjects, b, folder)
	}
}

func (h Handler) ListFolderObjects(c echo.Context) error {
	svc := s3.New(getSession())
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
		MaxKeys:   aws.Int64(maxKeys),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(folderKey),
	})
	if err != nil {
		// Region Control
		if _, ok := err.(awserr.RequestFailure); ok {
			region, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{
				Bucket: aws.String(bucket),
			})
			config.Conf.UpdateAwsRegion(*region.LocationConstraint)
			svc = s3.New(getSession())
			resp, err = svc.ListObjectsV2(&s3.ListObjectsV2Input{
				Bucket:    aws.String(bucket),
				MaxKeys:   aws.Int64(maxKeys),
				Delimiter: aws.String("/"),
			})
			if err != nil {
				return c.Render(http.StatusOK, "album.html", result)
			}
		} else {
			return c.Render(http.StatusOK, "album.html", result)
		}
	}

	// Adding folders
	result.Folders = make([]handlers.Folder, len(resp.CommonPrefixes))
	for i, item := range resp.CommonPrefixes {
		result.Folders[i] = handlers.Folder{
			Name: *item.Prefix,
			Url:  c.Echo().URI(base_handlers.ListFolderObjects, bucket, strings.Replace(*item.Prefix, "/", ":", -1)),
		}
	}

	if resp.Contents == nil {
		return c.Render(http.StatusOK, "album.html", result)
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
			Name:    *item.Key,
			Url:     urlStr,
			Type:    fileType,
			IsValid: fileTypeIsValid,
		}
	}
	return c.Render(http.StatusOK, "album.html", result)
}

func (h Handler) CreateFolder(c echo.Context) error {
	svc := s3.New(getSession())
	newFolderName := c.FormValue("new_folder_name")
	folderName := c.FormValue("folder_name")
	bucket := c.ParamValues()[0]

	// Folder create
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(folderName + newFolderName + "/"),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	return c.JSON(http.StatusOK, handlers.JsonResponse{Error: false, Message: "Success"})

}

func (h Handler) ListObjectsWithKey(c echo.Context) error {
	svc := s3.New(getSession())
	bucket := c.ParamValues()[0]
	folderKey := c.QueryParam("folder_key")
	lastKey := c.QueryParam("last_key")
	var result = new(handlers.ListObjectsResult)
	result.Bucket = handlers.Bucket{
		Name:   bucket,
		Prefix: folderKey,
		Url:    bucket,
	}
	// Get objects
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(bucket),
		MaxKeys:    aws.Int64(maxKeys),
		Delimiter:  aws.String("/"),
		Prefix:     aws.String(folderKey),
		StartAfter: aws.String(lastKey),
	})
	if err != nil {
		return c.JSON(http.StatusOK, result)
	}

	// Adding folders
	result.Folders = make([]handlers.Folder, len(resp.CommonPrefixes))
	for i, item := range resp.CommonPrefixes {
		result.Folders[i] = handlers.Folder{
			Name: *item.Prefix,
			Url:  c.Echo().URI(base_handlers.ListFolderObjects, bucket, ""),
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
			Name:    *item.Key,
			Url:     urlStr,
			Type:    fileType,
			IsValid: fileTypeIsValid,
		}
	}
	return c.JSON(http.StatusOK, result)
}

func (h Handler) ListBuckets(c echo.Context) error {
	svc := s3.New(getSession())
	// List buckets
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
			Url:  c.Echo().URI(base_handlers.ListBaseObjects, *item.Name, ""),
		})
	}
	return c.Render(http.StatusOK, "buckets.html", buckets)
}

func (h Handler) CreateBucket(c echo.Context) error {
	svc := s3.New(getSession())
	bucketName := c.FormValue("bucket_name")
	// Create bucket
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	// Bucket is exists
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
	sess := getSession()
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	bucket := c.ParamValues()[0]
	folder_key := form.Value["folder_key_input"]
	files := form.File["file_input"]
	response := handlers.DetailedJsonResponse{Error: false, Message: "Success"}
	var wg sync.WaitGroup
	errors := make(chan string, len(files))
	wg.Add(len(files))
	wp := workerpool.New(workerNum)
	for _, file := range files {
		// Upload file func
		func(file *multipart.FileHeader, wg *sync.WaitGroup) {
			wp.Submit(func() {
				src, err := file.Open()
				if err != nil {
					if _, ok := err.(awserr.RequestFailure); ok {
						errors <- file.Filename
						return
					}
				}
				defer src.Close()
				defer wg.Done()
				// Copy file
				uploader := s3manager.NewUploader(sess)
				_, err = uploader.Upload(&s3manager.UploadInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(folder_key[0] + file.Filename),
					Body:   src,
				})
				if err != nil {
					if _, ok := err.(awserr.RequestFailure); ok {
						errors <- file.Filename
						return
					}
				}
			})
		}(file, &wg)
	}
	wg.Wait()
	wp.StopWait()
	close(errors)
	for e := range errors {
		response.Failed = append(response.Failed, e)
	}
	return c.JSON(http.StatusOK, response)
}

func (h Handler) DeleteBuckets(c echo.Context) error {
	svc := s3.New(getSession())
	_ = c.FormValue("buckets[]")
	buckets := c.Request().Form["buckets[]"]
	errors := make(chan string, len(buckets))
	response := handlers.DetailedJsonResponse{Error: false, Message: "Success"}
	var wg sync.WaitGroup
	wg.Add(len(buckets))
	for _, bucket := range buckets {
		// Delete bucket func
		go func(bucket string, wg *sync.WaitGroup) {
			defer wg.Done()
			_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				if _, ok := err.(awserr.RequestFailure); ok {
					errors <- bucket
					return
				}
			}
			// Bucket is delete control
			err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				if _, ok := err.(awserr.RequestFailure); ok {
					errors <- bucket
					return
				}
			}
		}(bucket, &wg)
	}
	wg.Wait()
	close(errors)
	for e := range errors {
		response.Failed = append(response.Failed, e)
	}
	return c.JSON(http.StatusOK, response)
}

// Create object struct for delete object
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
	_ = c.FormValue("keys[]")
	keys := c.Request().Form["keys[]"]
	otd := getObjectsToDelete(keys)
	svc := s3.New(getSession())
	bucket := c.ParamValues()[0]
	// Delete Objects
	_, err := svc.DeleteObjects(&s3.DeleteObjectsInput{
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
	// Exists control
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(keys[len(keys)-1]),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			return c.JSON(http.StatusOK, handlers.JsonResponse{Error: true, Message: awsErr.Message()})
		}
	}
	return c.JSON(http.StatusOK, handlers.JsonResponse{Error: false, Message: "Objects deleted"})
}

func (h Handler) DeleteFolders(c echo.Context) error {
	_ = c.FormValue("keys[]")
	keys := c.Request().Form["keys[]"]
	svc := s3.New(getSession())
	bucket := c.ParamValues()[0]
	response := handlers.DetailedJsonResponse{Error: false, Message: "Success"}
	var wg sync.WaitGroup
	wg.Add(len(keys))
	errors := make(chan string, len(keys))
	wp := workerpool.New(workerNum)
	for _, key := range keys {
		// Delete folder func
		func(bucket string, key string, wg *sync.WaitGroup, wp *workerpool.WorkerPool) {
			wp.Submit(func() {
				defer wg.Done()
				iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
					Bucket: aws.String(bucket),
					Prefix: aws.String(key),
				})
				if err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter); err != nil {
					errors <- key
					return
				}
			})
		}(bucket, key, &wg, wp)
	}
	wg.Wait()
	wp.StopWait()
	close(errors)
	for e := range errors {
		response.Failed = append(response.Failed, e)
	}
	return c.JSON(http.StatusOK, response)
}
