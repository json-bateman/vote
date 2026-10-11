-- name: GetVotesByVoterId :one
SELECT * FROM votes 
WHERE voter_id = ?;

-- name: GetAllVotes :many
SELECT id, vote, voter_id, ip, created_at FROM votes;

-- name: AddVote :one
INSERT INTO votes (vote, voter_id, ip)
VALUES (?,?,?)
RETURNING id, vote, voter_id, ip, created_at;

-- name: CountVotes :one
SELECT COUNT(*) FROM votes;
