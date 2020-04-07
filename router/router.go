package router

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"html/template"
	"io"
	"main/views"
)

//func ValidKeyControl(key string, ctx echo.Context) (b bool, err error) {
//	validApiKey := "1234"
//	return key == validApiKey, nil
//}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func New() *echo.Echo {
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
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
	//e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
	//	KeyLookup:  "header:authorization",
	//	Validator: ValidKeyControl,
	//}))
	//e.Use(middleware.KeyAuth(ValidKeyControl))
	//e.Use(createAwsSession())
	e.Static("/static", "static")

	e.GET("/list_buckets", views.ListBuckets)
	e.POST("/create_bucket", views.CreateBucket)
	e.GET("/:bucket/list_objects", views.ListObjects)
	e.POST("/:bucket/upload_file", views.UploadFileToBucket)

	e.GET("/:bucket/delete_bucket", views.DeleteBuckets)
	e.POST("/:bucket/delete_object", views.DeleteObjects)

	return e
}

//func createAwsSession() echo.MiddlewareFunc {
//	return func(next echo.HandlerFunc) echo.HandlerFunc {
//		return func(c echo.Context) error {
//			sess, err := session.NewSession(&aws.Config{
//				Credentials:credentials.NewStaticCredentials(config.AwsId, config.AwsSecretKey, ""),
//				Region: 	aws.String(config.AwsRegion),
//			})
//			svc := s3.New(sess)
//			if err != nil {
//				panic("Credentials is not correct")
//			}
//			c.Set("AwsS3Session", svc)
//			return next(c)
//		}
//	}
//}
