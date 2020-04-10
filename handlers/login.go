package handlers

import (
	"github.com/labstack/echo"
	"main/config"
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
		if storageProvider == "s3" {
			config.Conf = config.Config{
				Status: true,
				AwsConfig: config.AwsConfig{
					AwsId:        ID,
					AwsSecretKey: SecretKey,
					AwsRegion:    Region,
				},
			}
		}

	}
	return c.Redirect(http.StatusMovedPermanently, "/list_buckets")
}
