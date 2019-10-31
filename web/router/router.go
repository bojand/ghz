package router

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bojand/ghz/web/api"
	"github.com/bojand/ghz/web/config"
	"github.com/bojand/ghz/web/database"
	"github.com/rakyll/statik/fs"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	// for bundled resources
	_ "github.com/bojand/ghz/web/router/statik"
)

// New creates new server
func New(db *database.Database, appInfo *api.ApplicationInfo, conf *config.Config) (*echo.Echo, error) {
	s := echo.New()

	s.Logger.SetLevel(getLogLevel(conf))
	output, err := getLogOutput(conf)
	if err != nil {
		return nil, err
	}
	s.Logger.SetOutput(output)

	s.Validator = &CustomValidator{validator: validator.New()}

	s.Pre(middleware.AddTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		Skipper: func(ctx echo.Context) bool {
			path := ctx.Request().URL.Path
			return !strings.Contains(path, "/api")
		},
	}))

	s.Use(middleware.CORS())

	s.Use(middleware.RequestID())
	s.Use(middleware.Logger())
	s.Use(middleware.Recover())

	// API

	apiRoot := s.Group("/api")

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
	reportGroup.DELETE("/:rid/", reportAPI.DeleteReport).Name = "ghz api: delete report"
	reportGroup.GET("/:rid/previous/", reportAPI.GetPreviousReport).Name = "ghz api: get previous report"
	reportGroup.POST("/bulk_delete/", reportAPI.DeleteReportBulk).Name = "ghz api: delete bulk report"

	optionsAPI := api.OptionsAPI{DB: db}
	reportGroup.GET("/:rid/options/", optionsAPI.GetOptions).Name = "ghz api: get options"

	histogramAPI := api.HistogramAPI{DB: db}
	reportGroup.GET("/:rid/histogram/", histogramAPI.GetHistogram).Name = "ghz api: get histogram"

	exportAPI := api.ExportAPI{DB: db}
	reportGroup.GET("/:rid/export/", exportAPI.GetExport).Name = "ghz api: get export"

	// Ingest

	ingestAPI := api.IngestAPI{DB: db}
	apiRoot.POST("/ingest/", ingestAPI.Ingest).Name = "ghz api: ingest"

	// Ingest to project
	projectGroup.POST("/:pid/ingest/", ingestAPI.IngestToProject).Name = "ghz api: ingest to project"

	// Info

	infoAPI := api.InfoAPI{Info: *appInfo}
	apiRoot.GET("/info/", infoAPI.GetApplicationInfo).Name = "ghz api: get info"

	// Frontend

	// load the precompiled statik fs
	statikFS, err := fs.New()
	if err != nil {
		return nil, err
	}

	// get the index file
	indexFile, err := fs.ReadFile(statikFS, "/index.html")
	if err != nil {
		return nil, err
	}

	// wrap the handler
	assetHandler := http.FileServer(statikFS)
	wrapHandler := echo.WrapHandler(assetHandler)

	// our custom handler
	s.GET("/*", func(ctx echo.Context) error {
		// if root just pass through to the fs handler
		path := ctx.Request().URL.Path
		if path == "/" {
			return wrapHandler(ctx)
		}

		// if it has an extension means it's a file
		// so pass through to the fs handler
		ext := filepath.Ext(path)
		if len(ext) > 0 {
			return wrapHandler(ctx)
		}

		// otherwise serve the index file
		// React router will handle the path from there on
		return ctx.HTML(200, string(indexFile))
	})

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

func getLogOutput(config *config.Config) (io.Writer, error) {
	logPath := strings.TrimSpace(config.Log.Path)
	if logPath == "" {
		return os.Stdout, nil
	}

	if _, err := os.Stat(filepath.Dir(logPath)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(logPath), 0777); err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// PrintRoutes prints routes in the server
func PrintRoutes(echoServer *echo.Echo) {
	routes := echoServer.Routes()
	for _, r := range routes {
		index := strings.Index(r.Name, "ghz api:")
		if index >= 0 {
			desc := fmt.Sprintf("[%+v] %+v %+v", r.Name, r.Method, r.Path)
			echoServer.Logger.Info(desc)
		}
	}
}
