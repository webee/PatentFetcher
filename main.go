package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/webee/x/xconfig"
)

type (
	// Status is the status of application
	Status struct {
		Startup time.Time `json:"startup"`
	}
)

var (
	config = new(Config)
	status = new(Status)
)

func init() {
	// config
	xconfig.Load(config)
}

func main() {
	e := echo.New()

	// Echo debug setting
	if config.Debug {
		e.Debug = true
	}

	// common middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// routes
	e.GET("/status", getStatus)

	status.Startup = time.Now()
	e.Logger.Fatal(e.Start(config.Address))
}

// API状态 成功204 失败500
func getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, status)
}
