package models

import "time"

type Workout struct {
	ID              int64          `json:"id"`
	UserID          int64          `json:"user_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	DurationMinutes int            `json:"duration_minutes"`
	CaloriesBurned  int            `json:"calories_burned"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Entries         []WorkoutEntry `json:"entries"`
}
