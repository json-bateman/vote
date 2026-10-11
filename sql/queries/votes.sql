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

-- name: GetVotesPerSecond :many
-- Pre-aggregated per second; Go buckets these into the chart's time chunks.
SELECT created_at,
       CAST(SUM(vote) AS INTEGER) AS cheers,
       CAST(COUNT(*) - SUM(vote) AS INTEGER) AS boos
FROM votes
GROUP BY created_at
ORDER BY created_at;
