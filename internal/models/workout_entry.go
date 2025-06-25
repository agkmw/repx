package models

import "time"

type WorkoutEntry struct {
	ID              int64     `json:"id"`
	WorkoutID       int64     `json:"workout_id"`
	ExerciseName    string    `json:"exercise_name"`
	Sets            int       `json:"sets"`
	Reps            *int      `json:"reps"`
	DurationSeconds *int      `json:"duration_seconds"`
	Weight          *float64  `json:"weight"`
	Notes           string    `json:"notes"`
	OrderIndex      int       `json:"order_index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
