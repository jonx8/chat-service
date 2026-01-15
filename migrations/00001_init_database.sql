-- +goose Up
-- +goose StatementBegin

CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_messages_chat_id_created_at ON messages(chat_id, created_at DESC);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;

-- +goose StatementEnd