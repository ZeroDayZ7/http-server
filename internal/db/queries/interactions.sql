-- name: IncrementCounter :exec
UPDATE interaction_stats 
SET current_count = current_count + 1 
WHERE type = ?;

-- name: GetCountByType :one
SELECT current_count FROM interaction_stats 
WHERE type = ?;