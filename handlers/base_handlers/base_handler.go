package base_handlers

import (
	"github.com/labstack/echo"
	"main/config"
)

func ListBaseObjects(c echo.Context) error {
	return config.Conf.Provider.ListBaseObjects(c)
}

func ListFolderObjects(c echo.Context) error {
	return config.Conf.Provider.ListFolderObjects(c)
}

func ListObjectsWithKey(c echo.Context) error {
	return config.Conf.Provider.ListObjectsWithKey(c)
}

func CreateFolder(c echo.Context) error {
	return config.Conf.Provider.CreateBucket(c)
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
