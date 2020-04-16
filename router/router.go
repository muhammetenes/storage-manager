package router

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"html/template"
	"io"
	"main/config"
	"main/handlers"
	"main/handlers/login"
	"net/http"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func New() *echo.Echo {
	t := &Template{
		templates: template.Must(template.ParseGlob("/Users/enes/go/src/storage-manager/templates/*.html")),
	}
	e := echo.New()
	e.Renderer = t
	e.Logger.SetLevel(log.DEBUG)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization},
		AllowMethods: []string{
			echo.GET,
			echo.HEAD,
			echo.PUT,
			echo.PATCH,
			echo.POST,
			echo.DELETE},
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${status} ${method} ${host}${path} ${latency} ${latency_human}\n",
	}))
	e.Static("/static", "static")
	e.Use(credentialControl)

	e.GET("/login", login.LoginPage)
	e.POST("/login", login.Login)
	e.GET("/logout", login.Logout)
	e.GET("/list_buckets", handlers.ListBuckets)
	e.POST("/create_bucket", handlers.CreateBucket)
	e.GET("/:bucket/list_objects", handlers.ListBaseObjects)
	e.GET("/:bucket/list_objects/:key", handlers.ListFolderObjects)
	e.POST("/:bucket/upload_file", handlers.UploadFileToBucket)

	e.POST("/delete_buckets", handlers.DeleteBuckets)
	e.POST("/:bucket/delete_objects", handlers.DeleteObjects)

	return e
}

func credentialControl(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		loginRoutePath := c.Echo().URI(login.Login, nil)
		if config.Conf.Status == false && (c.Path() != loginRoutePath && c.Path() != "/static/*" && c.Path() != "") {
			return c.Redirect(http.StatusFound, loginRoutePath)
		}
		return next(c)
	}
}
