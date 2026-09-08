-- +goose Up
CREATE TABLE "user" (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE "user";
