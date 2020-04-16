package login

import (
	"github.com/labstack/echo"
	"main/config"
	"main/handlers/awss3"
	"net/http"
)

func LoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", "")
}

func Login(c echo.Context) error {
	storageProvider := c.FormValue("storage_provider")
	ID := c.FormValue("id")
	SecretKey := c.FormValue("secret_key")
	Region := c.FormValue("region")

	if storageProvider != "" && ID != "" && SecretKey != "" {
		switch storageProvider {
		case "s3":
			config.Conf.Init(true, awss3.Handler{}, ID, SecretKey, Region)
		case "gcs":
			config.Conf.Init(true, awss3.Handler{}, ID, SecretKey, Region)
		}
	}
	return c.Redirect(http.StatusFound, "/list_buckets")
}

func Logout(c echo.Context) error {
	config.Conf.DestroyConfig()
	return c.Redirect(http.StatusFound, "/login")
}
