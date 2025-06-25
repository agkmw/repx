package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/agkmw/workout-service/internal/api"
	"github.com/agkmw/workout-service/internal/middleware"
	"github.com/agkmw/workout-service/internal/store"
	"github.com/agkmw/workout-service/migrations"
)

type Application struct {
	Logger         *slog.Logger
	UserStore      store.UserStore
	UserHandler    *api.UserHandler
	WorkoutStore   store.WorkoutStore
	WorkoutHandler *api.WorkoutHandler
	TokenStore     store.TokenStore
	TokenHandler   *api.TokenHandler
	Middleware     *middleware.UserMiddleware
	DB             *sql.DB
}

func New() (*Application, error) {
	db, err := store.Open()
	if err != nil {
		return nil, err
	}

	if err := store.MigrateFS(db, migrations.FS, "."); err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// stores
	userStore := store.NewPostgresUserStore(db)
	workoutStore := store.NewPostgresWorkoutStore(db)
	tokenStore := store.NewPostgresTokenStore(db)

	// handlers
	userHandler := api.NewUserHandler(userStore, logger)
	workoutHandler := api.NewWorkoutHandler(workoutStore, logger)
	tokenHandler := api.NewTokenHandler(tokenStore, userStore, logger)

	// middleware
	middlewareHandler := middleware.NewUserMiddleware(userStore)

	app := &Application{
		Logger:         logger,
		UserStore:      userStore,
		UserHandler:    userHandler,
		WorkoutStore:   workoutStore,
		WorkoutHandler: workoutHandler,
		TokenStore:     tokenStore,
		TokenHandler:   tokenHandler,
		Middleware:     middlewareHandler,
		DB:             db,
	}

	return app, nil
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Status is available...")
}
