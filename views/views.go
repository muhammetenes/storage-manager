package views

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Bucket struct {
	Name string
	Url  string
}

type Object struct {
	Name string
	Url  string
	Type bool
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func getFileType(fileName string) string {
	result := strings.Split(fileName, ".")
	return result[len(result)-1]
}

func ListObjects(c echo.Context) error {
	validFileType := []string{"png", "PNG", "Png", "Jpeg", "JPEG", "Jpg", "JPG"}
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials("", "", ""),
		Region:      aws.String("eu-central-1"),
	})
	svc := s3.New(sess)
	bucket := c.ParamValues()[0]

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	//fmt.Println(resp)
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
	}
	var imageList []Object
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
		imageList = append(imageList, Object{
			Name: *item.Key,
			Url:  urlStr,
			Type: fileTypeIsValid,
		})
	}

	if err != nil {
		log.Println("Failed to sign request", err)
	}
	return c.Render(http.StatusOK, "album.html", imageList)
}

func ListBuckets(c echo.Context) error {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials("", "", ""),
		Region:      aws.String("eu-central-1"),
	})
	svc := s3.New(sess)
	bucket := ""

	resp, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
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
