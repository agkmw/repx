package store

import (
	"crypto/sha256"
	"database/sql"
	"time"

	"github.com/agkmw/workout-service/internal/models"
)

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}

func (pg *PostgresUserStore) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users 
		(username, email, password_hash, bio)
		VALUES 
		($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	if err := pg.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.PasswordHash.Hash,
		user.Bio,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (pg *PostgresUserStore) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{
		PasswordHash: models.Password{},
	}
	query := `
		SELECT
			id, username, email, password_hash, 
			bio, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	if err := pg.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.Hash,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return user, nil
}

func (pg *PostgresUserStore) SearchUsersByUsername(username string) ([]models.User, error) {
	users := []models.User{}

	query := `
		SELECT
			id, username, email, password_hash, 
			bio, created_at, updated_at
		FROM users
		WHERE username ILIKE $1
	`
	rows, err := pg.db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := models.User{
			PasswordHash: models.Password{},
		}

		err = rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash.Hash,
			&user.Bio,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (pg *PostgresUserStore) UpdateUser(user *models.User) error {
	query := `
		UPDATE users 
		SET
			username = $1, email = $2, bio = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at
	`
	err := pg.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.Bio,
		user.ID,
	).Scan(
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgresUserStore) GetUserByToken(scope, plaintextToken string) (*models.User, error) {
	tokenHash := sha256.Sum256([]byte(plaintextToken))

	user := &models.User{
		PasswordHash: models.Password{},
	}

	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.bio, u.created_at, u.updated_at
		FROM users u
		INNER JOIN tokens t ON t.user_id = u.id
		WHERE t.hash = $1 AND t.scope = $2 AND t.expiry > $3
	`
	if err := pg.db.QueryRow(query, tokenHash[:], scope, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.Hash,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return user, nil
}
