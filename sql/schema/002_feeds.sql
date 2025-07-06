-- +goose Up
CREATE TABLE feeds (
	id 	UUID PRIMARY KEY,
	name	TEXT NOT NULL,
	url 	TEXT UNIQUE NOT NULL,
	user_id UUID NOT NULL REFERENCES users
		ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;
