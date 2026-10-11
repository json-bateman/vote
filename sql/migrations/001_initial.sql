-- +goose Up
CREATE TABLE votes (
    id INTEGER PRIMARY KEY,
    voter_id TEXT NOT NULL,
    vote BOOLEAN NOT NULL, 
    ip TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX idx_votes_voter_id ON votes (voter_id);
