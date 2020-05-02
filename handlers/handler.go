package handlers

import (
	"github.com/labstack/echo"
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
	Bucket            Bucket   `json:"bucket"`
	Objects           []Object `json:"objects"`
	Folders           []Folder `json:"folders"`
	Count             int      `json:"count"`
	PreviousFolderUrl string   `json:"previous_folder_url"`
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

type Handler interface {
	ListBaseObjects(c echo.Context) error
	ListFolderObjects(c echo.Context) error
	ListObjectsWithKey(c echo.Context) error
	ListBuckets(c echo.Context) error
	CreateBucket(c echo.Context) error
	UploadFileToBucket(c echo.Context) error
	CreateFolder(c echo.Context) error
	DeleteBuckets(c echo.Context) error
	DeleteObjects(c echo.Context) error
	DeleteFolders(c echo.Context) error
}
