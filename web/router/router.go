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
func New(db *database.Database, conf *config.Config) (*echo.Echo, error) {
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

	projectHandler := api.ProjectAPI{DB: db}

	projectGroup := apiRoot.Group("/projects")

	// g.GET("/", projectHandler).Name = "ghz api: list projects"
	projectGroup.POST("/", projectHandler.CreateProject).Name = "ghz api: create project"

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
