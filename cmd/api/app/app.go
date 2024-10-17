package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/grocky/go-api-starter/cmd/api/middleware"
	"github.com/grocky/go-api-starter/cmd/api/response"
	"github.com/grocky/go-api-starter/cmd/api/server"
	"github.com/grocky/go-api-starter/internal/log"
	"github.com/grocky/go-api-starter/internal/mysql"
	"github.com/grocky/go-api-starter/internal/version"
	"net/http"
	"sync"
)

type App struct {
	db *mysql.DB
	sync.WaitGroup
	//service go-api-starter.Service
}

func New(db *mysql.DB) *App {
	return &App{
		db: db,
	}
}

func (app *App) Routes(ctx context.Context) *mux.Router {
	logger := log.FromContext(ctx).Named("app")

	r := mux.NewRouter()
	r.NotFoundHandler = server.NotFoundHandler()
	r.MethodNotAllowedHandler = server.MethodNotAllowedHandler()

	r.Use(middleware.Recovery())
	r.Use(middleware.PopulateLogger(logger))
	r.Use(middleware.PopulateRequestID())

	r.HandleFunc("/status", app.Status)

	return r
}

func (app *App) Status(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}

	dbStats := app.db.Stats()
	data["dbConnection"] = dbStats
	data["version"] = version.Get()

	if err := response.JSON(w, http.StatusOK, data); err != nil {
		server.Error(w, r, err)
	}
}
