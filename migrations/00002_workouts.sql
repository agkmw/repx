-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workouts (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  -- user_id
  title VARCHAR (255) NOT NULL,
  description TEXT,
  duration_minutes INTEGER NOT NULL,
  calories_burned INTEGER,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workouts;
-- +goose StatementEnd

