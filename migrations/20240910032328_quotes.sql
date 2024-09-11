-- +goose Up
-- +goose StatementBegin
CREATE TABLE quotes (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    author TEXT,
    category TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE quotes;
-- +goose StatementEnd
