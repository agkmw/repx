package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/agkmw/workout-service/internal/models"
	"github.com/agkmw/workout-service/internal/store"
	"github.com/agkmw/workout-service/internal/tokens"
	"github.com/agkmw/workout-service/internal/utils"
)

type UserMiddleware struct {
	UserStore store.UserStore
}

func NewUserMiddleware(userStore store.UserStore) *UserMiddleware {
	return &UserMiddleware{
		UserStore: userStore,
	}
}

type contextKey string

const UserContextKey = contextKey("use")

func SetUser(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, user)
	return r.WithContext(ctx)
}

func GetUser(r *http.Request) *models.User {
	// asserts that the value retrieving from Context() is a pointer to models.User
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {
		// crash the app as soon as someone is injecting someting (bad actor call)
		panic("missing user in request")
	}
	return user
}

func (um *UserMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			r = SetUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authHeader, " ") // Bearer <TOKEN>
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"status":  "fail",
				"message": "Invalid authorization header.",
			})
			return
		}

		token := headerParts[1]
		user, err := um.UserStore.GetUserByToken(tokens.ScopeAuth, token)
		if err != nil || user == nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"status":  "fail",
				"message": "Token expired, or invalid token.",
			})
			return
		}

		r = SetUser(r, user)
		next.ServeHTTP(w, r)
		return
	})
}

func (um *UserMiddleware) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)

		if user.IsAnonymous() {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"status":  "fail",
				"message": "You must be logged in to access this route.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
