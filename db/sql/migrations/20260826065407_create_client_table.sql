-- +goose Up
CREATE TABLE client (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    secret_hash VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE client;
