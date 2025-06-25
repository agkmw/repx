package store

import (
	"time"

	"github.com/agkmw/workout-service/internal/models"
	"github.com/agkmw/workout-service/internal/tokens"
)

type UserStore interface {
	CreateUser(*models.User) error
	SearchUsersByUsername(username string) ([]models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(*models.User) error
	GetUserByToken(scope, plaintextToken string) (*models.User, error)
}

type WorkoutStore interface {
	CreateWorkout(*models.Workout) error
	GetWorkoutByID(id int64) (*models.Workout, error)
	UpdateWorkoutByID(*models.Workout) error
	DeleteWorkoutByID(id int64) error
	GetWorkoutOwner(id int64) (int64, error)
}

type TokenStore interface {
	Insert(token *tokens.Token) error
	CreateNewToken(userID int64, ttl time.Duration, scope string) (*tokens.Token, error)
	DeleteAllTokensForUser(userID int64, scope string) error
}
