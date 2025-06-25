package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/agkmw/workout-service/internal/middleware"
	"github.com/agkmw/workout-service/internal/models"
	"github.com/agkmw/workout-service/internal/store"
	"github.com/agkmw/workout-service/internal/utils"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
	logger       *slog.Logger
}

func NewWorkoutHandler(workoutStore store.WorkoutStore, logger *slog.Logger) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore: workoutStore,
		logger:       logger,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		// Handle bad input
		wh.logger.Warn("failed to read or parse workout id parameter", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid workout ID. Please provide a valid numeric identiifer.",
		})
		return
	}

	workout, err := wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		// Handle "Not Found" error
		if errors.Is(err, sql.ErrNoRows) {
			wh.logger.Warn("workout not found for given id", "workout_id", workoutID)
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{
				"status":  "fail",
				"message": "The requested workout could not be found.",
			})
			return
		}

		// Handle other server errors
		wh.logger.Error("failed to fetch workout by id", "workout_id", workoutID, "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "Failed to fetch the workout due to a server error. Please try again later.",
		})
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"status": "success",
		"data": map[string]*models.Workout{
			"workout": workout,
		},
	}); err != nil {
		wh.logger.Error("failed to write success response for get workout", "workout_id", workoutID, "error", err)
		return
	}
	wh.logger.Info("workout served successfully", "workout_id", workout.ID)
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	workout := &models.Workout{}
	if err := json.NewDecoder(r.Body).Decode(workout); err != nil {
		wh.logger.Warn("failed to decode workout create request payload", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid request payload. Please ensure all fields are correctly provided.",
		})
		return
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == models.AnonymousUser {
		wh.logger.Warn("unauthorized attempt to create a new workout", "username", currentUser.Username)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "You must be logged in to create a new workout.",
		})
		return
	}

	workout.UserID = currentUser.ID

	// TODO: Add field validation

	if err := wh.workoutStore.CreateWorkout(workout); err != nil {
		wh.logger.Error("failed to execute workout creation in store", "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "Failed to create the workout due to a server error. Please try again later.",
		})
		return
	}

	if err := utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"status": "success",
		"data": map[string]*models.Workout{
			"workout": workout,
		},
	}); err != nil {
		wh.logger.Error("failed to write success response for create workout", "workout_id", workout.ID, "error", err)
		return
	}
	wh.logger.Info("workout created successfully", "workout_id", workout.ID)
}

func (wh *WorkoutHandler) HandleUpdateWorkoutByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		// Handle bad input
		wh.logger.Warn("failed to read or parse workout id parameter", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid workout ID. Please provide a valid numeric identiifer.",
		})
		return
	}

	// Check if the workout to update exists
	existingWorkout, err := wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		// Handle "Not Found" error
		if errors.Is(err, sql.ErrNoRows) {
			wh.logger.Warn("attempted to update a workout that does not exist", "workout_id", workoutID)
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{
				"status":  "fail",
				"message": "The workout you are trying to update could not be found.",
			})
			return
		}

		// Handle other server errors
		wh.logger.Error("failed to fetch workout for update", "workout_id", workoutID, "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "An unexpected error occurred while preparing to update. Please try again later.",
		})
		return
	}

	var updateWorkoutRequest struct {
		Title           *string               `json:"title"`
		Description     *string               `json:"description"`
		DurationMinutes *int                  `json:"duration_minutes"`
		CaloriesBurned  *int                  `json:"calories_burned"`
		Entries         []models.WorkoutEntry `json:"entries"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateWorkoutRequest); err != nil {
		wh.logger.Warn("failed to decode workout update request", "workout_id", workoutID, "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid request payload. Please ensure all fields are correctly provided.",
		})
		return
	}

	if updateWorkoutRequest.Title != nil {
		existingWorkout.Title = *updateWorkoutRequest.Title
	}
	if updateWorkoutRequest.Description != nil {
		existingWorkout.Description = *updateWorkoutRequest.Description
	}
	if updateWorkoutRequest.DurationMinutes != nil {
		existingWorkout.DurationMinutes = *updateWorkoutRequest.DurationMinutes
	}
	if updateWorkoutRequest.CaloriesBurned != nil {
		existingWorkout.CaloriesBurned = *updateWorkoutRequest.CaloriesBurned
	}
	if updateWorkoutRequest.Entries != nil {
		existingWorkout.Entries = updateWorkoutRequest.Entries
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == models.AnonymousUser {
		wh.logger.Warn("unauthorized attempt to update a workout", "username", currentUser.Username)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "You must be logged in to update a workout.",
		})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			wh.logger.Warn("attempted to update a workout that does not exist", "error", err)
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{
				"status":  "fail",
				"message": "The workout you are trying to update could not be found.",
			})
			return
		}

		wh.logger.Error("failed to fetch workout for update", "workout_id", workoutID, "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "An unexpected error occurred while preparing to update. Please try again later.",
		})
		return
	}

	if workoutOwner != currentUser.ID {
		wh.logger.Warn("unauthorized attempt to update a workout", "user_id", currentUser.ID)
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{
			"status":  "fail",
			"message": "You are not authorized to update this workout.",
		})
		return
	}

	// TODO:Add field validation

	if err := wh.workoutStore.UpdateWorkoutByID(existingWorkout); err != nil {
		wh.logger.Error("failed to execute workout update in store", "workout_id", workoutID, "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "Failed to update the workout due to a server error. Please try again later.",
		})
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"status": "success",
		"data": map[string]*models.Workout{
			"workout": existingWorkout,
		},
	}); err != nil {
		wh.logger.Error("failed to write success response for update workout", "workout_id", workoutID, "error", err)
		return
	}
	wh.logger.Info("workout updated successfully", "workout_id", workoutID)
}

func (wh *WorkoutHandler) HandleDeleteWorkoutByID(w http.ResponseWriter, r *http.Request) {
	workoutID, err := utils.ReadIDParam(r)
	if err != nil {
		wh.logger.Warn("failed to read or parse workout id parameter", "error", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "Invalid workout ID. Please provide a valid numeric identiifer.",
		})
		return
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == models.AnonymousUser {
		wh.logger.Warn("unauthorized attempt to delete a workout", "username", currentUser.Username)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"status":  "fail",
			"message": "You must be logged in to delete a workout.",
		})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			wh.logger.Warn("attempted to delete a workout that does not exist", "error", err)
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{
				"status":  "fail",
				"message": "The workout you are trying to delete could not be found.",
			})
			return
		}

		wh.logger.Error("failed to fetch workout for delete", "workout_id", workoutID, "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "An unexpected error occurred while preparing to delete. Please try again later.",
		})
		return
	}

	if workoutOwner != currentUser.ID {
		wh.logger.Warn("unauthorized attempt to delete a workout", "user_id", currentUser.ID)
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{
			"status":  "fail",
			"message": "You are not authorized to delete this workout.",
		})
		return
	}

	err = wh.workoutStore.DeleteWorkoutByID(workoutID)
	if err != nil {
		if err == sql.ErrNoRows {
			wh.logger.Warn("attempted to delete a workout that does not exist", "workout_id", workoutID, "error", err)
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{
				"status":  "fail",
				"message": "The workout you are tyring to delete could not be found.",
			})
			return
		}

		wh.logger.Error("failed to execute workout deletion in store", "workout_id", workoutID, "error", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"status":  "error",
			"message": "Failed to delete the workout due to a server error. Please try again later.",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
	wh.logger.Info("workout deleted successfully", "workout_id", workoutID)
}
