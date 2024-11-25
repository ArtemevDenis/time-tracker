package endpoint

import (
	"fmt"
	"github.com/ArtemevDenis/time-tracker/internal/app/auth"
	"github.com/ArtemevDenis/time-tracker/internal/app/entity"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type Service interface {
	CreateTask(task *entity.Task) (*entity.Task, error)
	GetTasks(filter *entity.TaskFilter) ([]entity.Task, error)
	UpdateTask(task *entity.Task, authorId *primitive.ObjectID) (*entity.Task, error)
	DeleteTask(task, authorId *primitive.ObjectID) error
	GetUserByEmail(string *string) (*entity.User, error)
}

type Endpoint struct {
	s Service
}

func New(s Service) *Endpoint {
	return &Endpoint{
		s: s,
	}
}

func (e *Endpoint) Status(ctx echo.Context) error {
	err := ctx.String(http.StatusOK, "ok")
	if err != nil {
		return err
	}

	return nil
}

func (e *Endpoint) CreateTask(ctx echo.Context) (err error) {
	token, err := auth.GetTokenFromCtx(ctx, "user")

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	claims, err := auth.GetClaims(token)

	t := new(entity.Task)

	if err = ctx.Bind(t); err != nil {
		log.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(t); err != nil {
		log.Error(err.Error())
		return err
	}

	task := &entity.Task{
		Title:       t.Title,
		Date:        t.Date,
		Description: t.Description,
		Duration:    t.Duration,
		Author:      claims.Name,
		AuthorId:    claims.ID,
	}

	fmt.Println(task.Date)
	t, err = e.s.CreateTask(task)
	return ctx.JSON(http.StatusCreated, t)
}

func (e *Endpoint) GetTasks(ctx echo.Context) (err error) {
	token, err := auth.GetTokenFromCtx(ctx, "user")

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	claims, err := auth.GetClaims(token)

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	f := new(entity.TaskFilter)

	if err = ctx.Bind(f); err != nil {
		log.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(f); err != nil {
		log.Error(err.Error())
		return err
	}

	filter := &entity.TaskFilter{
		Title:       f.Title,
		ID:          f.ID,
		Description: f.Description,
		DurationMax: f.DurationMax,
		DurationMin: f.DurationMin,
		Tag:         f.Tag,
		Author:      f.Author,
		DateFrom:    f.DateFrom,
		DateTo:      f.DateTo,
		AuthorId:    claims.ID,
	}

	var tasks []entity.Task
	tasks, err = e.s.GetTasks(filter)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return ctx.JSON(http.StatusOK, tasks)
}

func (e *Endpoint) UpdateTask(ctx echo.Context) (err error) {
	token, err := auth.GetTokenFromCtx(ctx, "user")

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	claims, err := auth.GetClaims(token)

	var id primitive.ObjectID
	id, err = primitive.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		log.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	t := new(entity.Task)
	if err = ctx.Bind(t); err != nil {
		log.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(t); err != nil {
		log.Error(err.Error())
		return err
	}

	task := &entity.Task{
		ID:          id,
		Title:       t.Title,
		Date:        t.Date,
		Description: t.Description,
		Duration:    t.Duration,
		Author:      claims.Name,
		AuthorId:    claims.ID,
	}
	t, err = e.s.UpdateTask(task, &claims.ID)
	return ctx.JSON(http.StatusOK, t)
}

func (e *Endpoint) DeleteTask(ctx echo.Context) (err error) {
	token, err := auth.GetTokenFromCtx(ctx, "user")

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	claims, err := auth.GetClaims(token)

	var id primitive.ObjectID
	id, err = primitive.ObjectIDFromHex(ctx.Param("id"))

	err = e.s.DeleteTask(&id, &claims.ID)
	return ctx.JSON(http.StatusOK, "ok")
}

func (e *Endpoint) Login(ctx echo.Context) (err error) {
	u := new(entity.User)

	hash, err := bcrypt.GenerateFromPassword([](byte)("pass"), bcrypt.DefaultCost)

	fmt.Println(string(hash))
	if err = ctx.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	storedUser, err := e.s.GetUserByEmail(&u.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(u.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Password is incorrect")
	}

	var jwt = new(entity.JWT)
	jwt, err = auth.GenerateTokens(storedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token is incorrect")
	}

	return ctx.JSON(http.StatusOK, jwt)
}

// todo: make new access token and return it
func (e *Endpoint) Refresh(ctx echo.Context) (err error) {

	tokenString := ctx.Request().Header.Get("Refresh-Token")
	fmt.Println(tokenString)

	token, err := jwt.ParseWithClaims(tokenString, &auth.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(auth.GetRefreshJWTSecret()), nil
	})

	fmt.Println(token.Valid)

	claims, ok := token.Claims.(*auth.JwtCustomClaims)
	if !ok || !token.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Password is incorrect")
	}

	fmt.Println(claims.ID)

	return ctx.JSON(http.StatusOK, claims)
}
