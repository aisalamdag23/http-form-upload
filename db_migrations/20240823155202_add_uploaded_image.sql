-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS uploaded_image_metadata (
    id serial PRIMARY KEY NOT NULL,
    filename VARCHAR(255) NOT NULL,
    filepath VARCHAR(500) NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    size INTEGER NOT NULL,
    ip_address VARCHAR(255) NULL,
    user_agent VARCHAR(255) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE uploaded_image_metadata;
-- +goose StatementEnd
