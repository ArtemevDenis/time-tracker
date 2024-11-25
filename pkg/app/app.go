package app

import (
	"github.com/ArtemevDenis/time-tracker/internal/app/auth"
	"github.com/ArtemevDenis/time-tracker/internal/app/db"
	"github.com/ArtemevDenis/time-tracker/internal/app/endpoint"
	"github.com/ArtemevDenis/time-tracker/internal/app/middleware"
	"github.com/ArtemevDenis/time-tracker/internal/app/service"
	"github.com/go-playground/validator"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

type App struct {
	e    *endpoint.Endpoint
	s    *service.Service
	d    *mongo.Database
	echo *echo.Echo
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func New() (*App, error) {
	a := &App{}

	var url = "mongodb://localhost:27017/tasks?directConnection=true"
	a.d = db.New(url)

	a.s = service.New(a.d)

	a.e = endpoint.New(a.s)

	a.echo = echo.New()

	a.echo.Validator = &CustomValidator{validator: validator.New()}

	a.echo.Use(middleware.RoleCheck)

	a.echo.Use(echoMiddleware.Logger())
	a.echo.Use(echoMiddleware.Recover())
	a.echo.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	api := a.echo.Group("/api")
	api.GET("/status", a.e.Status)

	api.POST("/login", a.e.Login)
	api.POST("/refresh", a.e.Refresh)

	tasks := api.Group("/tasks")
	tasks.Use(echojwt.WithConfig(*auth.GetConfig()))

	tasks.POST("", a.e.CreateTask)
	tasks.GET("", a.e.GetTasks)
	tasks.PUT("/:id", a.e.UpdateTask)
	tasks.DELETE("/:id", a.e.DeleteTask)

	return a, nil
}

func (a *App) Run() error {
	log.Println("server running")

	err := a.echo.Start(":8080")
	if err != nil {
		log.Fatalf("failed to start http server: %w", err)
		return err
	}

	return nil
}
