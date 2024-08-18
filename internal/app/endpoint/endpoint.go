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
	UpdateTask(task *entity.Task) (*entity.Task, error)
	DeleteTask(task *primitive.ObjectID) error
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
		Author:      t.Author,
	}
	t, err = e.s.CreateTask(task)
	return ctx.JSON(http.StatusCreated, t)
}

func (e *Endpoint) GetTasks(ctx echo.Context) (err error) {
	token, err := auth.GetTokenFromCtx(ctx, "user")

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	claims, err := auth.GetClaims(token)

	fmt.Println(claims)

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
		Author:      f.Author, // claims.userID
		DateFrom:    f.DateFrom,
		DateTo:      f.DateTo,
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
		Author:      t.Author,
	}
	t, err = e.s.UpdateTask(task)
	return ctx.JSON(http.StatusOK, t)
}

func (e *Endpoint) DeleteTask(ctx echo.Context) (err error) {
	var id primitive.ObjectID
	id, err = primitive.ObjectIDFromHex(ctx.Param("id"))

	err = e.s.DeleteTask(&id)
	return ctx.JSON(http.StatusOK, "ok")
}

func (e *Endpoint) Login(ctx echo.Context) (err error) {
	pass := []byte("test")

	var hash []byte
	hash, err = bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	storedUser := entity.User{
		Name:     "test",
		Password: string(hash),
		ID:       primitive.NewObjectID(),
	}

	u := new(entity.User)
	if err = ctx.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(u.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Password is incorrect")
	}
	// If password is correct, generate tokens and set cookies.
	var jwt = new(entity.JWT)
	jwt, err = auth.GenerateTokens(&storedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token is incorrect")
	}

	return ctx.JSON(http.StatusOK, jwt)
}

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

	var id primitive.ObjectID
	id, err = primitive.ObjectIDFromHex(claims.ID)
	if err != nil {
		log.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	fmt.Println(id)
	//
	//token, err := jwt.ParseWithClaims(stringToken, claims, func(token *jwt.Token) (interface{}, error) {
	//	// since we only use the one private key to sign the tokens,
	//	// we also only use its public counter part to verify
	//	return []byte(auth.GetRefreshJWTSecret()), nil
	//})
	//
	//fmt.Println(token.Valid)
	//for key, val := range claims {
	//	fmt.Printf("Key: %v, value: %v\n", key, val)
	//}

	//claims = token.Claims.(*auth.JwtCustomClaims)
	//token := ctx.Request().Header.Get("Refresh-Token").(*jwt.Token)
	//claims, err := auth.GetClaims(token)
	//
	//fmt.Println(claims)

	return ctx.JSON(http.StatusOK, claims)
}
