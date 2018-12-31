package router

import (
	"fmt"
	"strings"

	"github.com/bojand/ghz/web/api"
	"github.com/bojand/ghz/web/config"
	"github.com/bojand/ghz/web/database"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

// New creates new server
func New(db *database.Database, appInfo *api.ApplicationInfo, conf *config.Config) (*echo.Echo, error) {
	s := echo.New()

	s.Logger.SetLevel(getLogLevel(conf))

	s.Validator = &CustomValidator{validator: validator.New()}

	s.Use(middleware.CORS())

	s.Pre(middleware.AddTrailingSlash())

	root := s.Group(conf.Server.RootURL)

	root.Use(middleware.RequestID())
	root.Use(middleware.Logger())
	root.Use(middleware.Recover())

	// API

	apiRoot := root.Group("/api")

	// Projects

	projectGroup := apiRoot.Group("/projects")

	projectAPI := api.ProjectAPI{DB: db}

	projectGroup.GET("/", projectAPI.ListProjects).Name = "ghz api: list projects"
	projectGroup.POST("/", projectAPI.CreateProject).Name = "ghz api: create project"
	projectGroup.GET("/:pid/", projectAPI.GetProject).Name = "ghz api: get project"
	projectGroup.PUT("/:pid/", projectAPI.UpdateProject).Name = "ghz api: update project"
	projectGroup.DELETE("/:pid/", projectAPI.DeleteProject).Name = "ghz api: delete project"

	// Reports by Project

	reportAPI := api.ReportAPI{DB: db}
	projectGroup.GET("/:pid/reports/", reportAPI.ListReportsForProject).Name = "ghz api: list reports for project"

	// Reports

	reportGroup := apiRoot.Group("/reports")
	reportGroup.GET("/", reportAPI.ListReportsAll).Name = "ghz api: list all reports"
	reportGroup.GET("/:rid/", reportAPI.GetReport).Name = "ghz api: get report"

	optionsAPI := api.OptionsAPI{DB: db}
	reportGroup.GET("/:rid/options/", optionsAPI.GetOptions).Name = "ghz api: get options"

	histogramAPI := api.HistogramAPI{DB: db}
	reportGroup.GET("/:rid/histogram/", histogramAPI.GetHistogram).Name = "ghz api: get histogram"

	// Ingest

	ingestAPI := api.IngestAPI{DB: db}
	apiRoot.POST("/ingest/", ingestAPI.Ingest).Name = "ghz api: ingest"

	// Ingest to project
	projectGroup.POST("/:pid/ingest/", ingestAPI.IngestToProject).Name = "ghz api: ingest to project"

	// Info

	infoAPI := api.InfoAPI{Info: *appInfo}
	apiRoot.GET("/info/", infoAPI.GetApplicationInfo).Name = "ghz api: get info"

	// Frontend

	s.Static("/", "ui/dist").Name = "ghz api: static"

	return s, nil
}

// CustomValidator is our validator for the API
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the input
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func getLogLevel(config *config.Config) log.Lvl {
	if config.Log.Level == "debug" {
		return log.DEBUG
	} else if config.Log.Level == "info" {
		return log.INFO
	} else if config.Log.Level == "warn" {
		return log.WARN
	} else if config.Log.Level == "error" {
		return log.ERROR
	} else {
		return log.OFF
	}
}

// PrintRoutes prints routes in the server
func PrintRoutes(echoServer *echo.Echo) {
	routes := echoServer.Routes()
	for _, r := range routes {
		index := strings.Index(r.Name, "ghz api:")
		if index >= 0 {
			desc := fmt.Sprintf("%+v %+v", r.Method, r.Path)
			fmt.Println(desc)
		}
	}

}
