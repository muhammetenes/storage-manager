package handlers

import (
	"github.com/labstack/echo"
	"main/config"
	"strings"
)

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type Bucket struct {
	Name   string
	Prefix string
	Url    string
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
	Bucket            Bucket
	Objects           []Object
	Folders           []Folder
	Count             int
	PreviousFolderUrl string
}

type ListBucketsResult struct {
	Buckets []Bucket
	Count   int
}

var ValidFileType = [...]string{"png", "PNG", "Png", "Jpeg", "JPEG", "Jpg", "JPG", "jpg", "jpeg"}

func GetFileType(fileName string) string {
	result := strings.Split(fileName, ".")
	return result[len(result)-1]
}

func ListBaseObjects(c echo.Context) error {
	return config.Conf.Provider.ListBaseObjects(c)
}

func ListFolderObjects(c echo.Context) error {
	return config.Conf.Provider.ListFolderObjects(c)
}

func ListBuckets(c echo.Context) error {
	return config.Conf.Provider.ListBuckets(c)
}

func CreateBucket(c echo.Context) error {
	return config.Conf.Provider.CreateBucket(c)
}

func UploadFileToBucket(c echo.Context) error {
	return config.Conf.Provider.UploadFileToBucket(c)
}

func DeleteBuckets(c echo.Context) error {
	return config.Conf.Provider.DeleteBuckets(c)
}

func DeleteObjects(c echo.Context) error {
	return config.Conf.Provider.DeleteObjects(c)
}
