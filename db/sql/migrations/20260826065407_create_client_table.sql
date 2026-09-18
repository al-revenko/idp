-- +goose Up
CREATE TABLE client (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    pub_key_url TEXT NOT NULL
);

-- +goose Down
DROP TABLE client;
