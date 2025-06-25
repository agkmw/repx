package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/agkmw/workout-service/internal/models"
	"github.com/agkmw/workout-service/internal/store"
	"github.com/agkmw/workout-service/internal/utils"
)

type registerUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

type UserHandler struct {
	userStore store.UserStore
	logger    *slog.Logger
}

func NewUserHandler(userStore store.UserStore, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    logger,
	}
}

func (uh *UserHandler) HandleRegisterUser(w http.ResponseWriter, r *http.Request) {
	req := &registerUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		uh.logger.Warn("failed to decode user register request", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid request payload. Please ensure all fields are correctly provided.",
		})
		return
	}

	if err := uh.validateUserRequest(req); err != nil {
		uh.logger.Warn("invalid user request", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid request payload. Please ensure all fields are correctly provided.",
		})
		return
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if err := user.PasswordHash.Set(req.Password); err != nil {
		uh.logger.Error("failed to hash password", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "An unexpected error occurred.",
		})
		return
	}

	if err := uh.userStore.CreateUser(user); err != nil {
		uh.logger.Error("failed to execute user registration in store", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "Failed to register the user due to a server error. Please try again later.",
		})
		return
	}

	if err := utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"status": "success",
		"data": map[string]any{
			"user": user,
		},
	}); err != nil {
		uh.logger.Error("failed to write user registration response", "user_id", user.ID, "error", err)
		return
	}

	uh.logger.Info("user created successfully", "user_id", user.ID)
}

func (uh *UserHandler) validateUserRequest(req *registerUserRequest) error {
	// validate username
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 5 {
		return errors.New("username must contain at least 5 characters")
	}
	if len(req.Username) > 50 {
		return errors.New("username can't be greater than 50 characters")
	}

	// validate email
	if req.Email == "" {
		return errors.New("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}

	// validate password
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 10 {
		return errors.New("password must contain at least 10 characters")
	}

	return nil
}
