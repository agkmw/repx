package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/agkmw/workout-service/internal/store"
	"github.com/agkmw/workout-service/internal/tokens"
	"github.com/agkmw/workout-service/internal/utils"
)

type createTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *slog.Logger
}

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *slog.Logger) *TokenHandler {
	return &TokenHandler{
		tokenStore: tokenStore,
		userStore:  userStore,
		logger:     logger,
	}
}

func (th *TokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	req := &createTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		th.logger.Warn("failed to decode token create request", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid request payload. Please ensure all fields are correctly provided.",
		})
		return
	}

	user, err := th.userStore.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		th.logger.Warn("failed to fetch user by username", "error", err)
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{
			"status":  "fail",
			"message": "User not found or incorrect credentials provided.",
		})
		return
	}

	passwordDoMatch, err := user.PasswordHash.Match(req.Password)
	if err != nil {
		th.logger.Warn("error comparing password hash", "error", err)
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{
			"status":  "error",
			"message": "User not found or incorrect credentials provided.",
		})
		return
	}

	if !passwordDoMatch {
		th.logger.Warn("invalid credentials provided")
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
			"status":  "fail",
			"message": "Invalid username or password. Please try again.",
		})
		return
	}

	token, err := th.tokenStore.CreateNewToken(user.ID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		th.logger.Error("failed to create authentication token", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "Failed to create an authentication token due to a server error.",
		})
		return
	}

	if err := utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"status": "success",
		"data": map[string]tokens.Token{
			"auth_token": *token,
		},
	}); err != nil {
		th.logger.Error("failed to write token creation response", "error", err)
	}
}
