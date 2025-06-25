package store

import (
	"database/sql"

	"github.com/agkmw/workout-service/internal/models"
)

type PostgresWorkoutStore struct {
	db *sql.DB
}

func NewPostgresWorkoutStore(db *sql.DB) *PostgresWorkoutStore {
	return &PostgresWorkoutStore{
		db: db,
	}
}

func (pg *PostgresWorkoutStore) GetWorkoutByID(id int64) (*models.Workout, error) {
	queryWorkout := `
		SELECT id, title, description, duration_minutes, calories_burned
		FROM workouts
		WHERE id = $1
	`
	workout := &models.Workout{}
	err := pg.db.QueryRow(queryWorkout, id).Scan(
		&workout.ID,
		&workout.Title,
		&workout.Description,
		&workout.DurationMinutes,
		&workout.CaloriesBurned,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// fetch entries
	queryEntry := `
		SELECT 
			id, workout_id, exercise_name, sets, reps, duration_seconds, 
			weight, notes, order_index, created_at, updated_at
		FROM workout_entries
		WHERE workout_id = $1
		ORDER BY order_index
	`
	rows, err := pg.db.Query(queryEntry, workout.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := models.WorkoutEntry{}
		if err := rows.Scan(
			&e.ID,
			&e.WorkoutID,
			&e.ExerciseName,
			&e.Sets,
			&e.Reps,
			&e.DurationSeconds,
			&e.Weight,
			&e.Notes,
			&e.OrderIndex,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		workout.Entries = append(workout.Entries, e)
	}

	return workout, nil
}

func (pg *PostgresWorkoutStore) CreateWorkout(workout *models.Workout) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertWorkout := `
		INSERT INTO workouts
		(user_id, title, description, duration_minutes, calories_burned)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	if err := tx.QueryRow(
		insertWorkout,
		workout.UserID,
		workout.Title,
		workout.Description,
		workout.DurationMinutes,
		workout.CaloriesBurned,
	).Scan(
		&workout.ID,
		&workout.CreatedAt,
		&workout.UpdatedAt,
	); err != nil {
		return err
	}

	// insert entries
	insertEntry := `
		INSERT INTO workout_entries
		(
			workout_id, exercise_name, sets, reps,
			duration_seconds, weight, notes, order_index
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	for i := range workout.Entries {
		entry := &workout.Entries[i]
		if err := tx.QueryRow(insertEntry,
			workout.ID,
			entry.ExerciseName,
			entry.Sets,
			entry.Reps,
			entry.DurationSeconds,
			entry.Weight,
			entry.Notes,
			entry.OrderIndex,
		).Scan(
			&entry.ID,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (pg *PostgresWorkoutStore) UpdateWorkoutByID(workout *models.Workout) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateWorkout := `
		UPDATE workouts 
		SET 
		title = $1, 
		description = $2, 
		duration_minutes = $3, 
		calories_burned = $4,
		updated_at = now()
		WHERE id = $5
		RETURNING updated_at
	`
	err = tx.QueryRow(
		updateWorkout,
		workout.Title,
		workout.Description,
		workout.DurationMinutes,
		workout.CaloriesBurned,
		workout.ID,
	).Scan(
		&workout.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// TODO: modify updating entries to use upsert
	_, err = tx.Exec("DELETE FROM workout_entries WHERE workout_id = $1", workout.ID)
	if err != nil {
		return err
	}

	insertEntry := `
		INSERT INTO workout_entries
		(
			workout_id, exercise_name, sets, reps,
			duration_seconds, weight, notes, order_index
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	for i := range workout.Entries {
		entry := &workout.Entries[i]
		err = tx.QueryRow(insertEntry,
			workout.ID,
			entry.ExerciseName,
			entry.Sets,
			entry.Reps,
			entry.DurationSeconds,
			entry.Weight,
			entry.Notes,
			entry.OrderIndex,
		).Scan(
			&entry.ID,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (pg *PostgresWorkoutStore) DeleteWorkoutByID(id int64) error {
	result, err := pg.db.Exec(`DELETE FROM workouts WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (pg *PostgresWorkoutStore) GetWorkoutOwner(workoutID int64) (int64, error) {
	var userID int64

	query := `
		SELECT user_id
		FROM workouts
		WHERE id = $1
	`
	if err := pg.db.QueryRow(query, workoutID).Scan(&userID); err != nil {
		return 0, err
	}

	return userID, nil
}
