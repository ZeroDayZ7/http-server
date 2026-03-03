-- name: GetStats :one
SELECT 
    MAX(CASE WHEN type = 'visit' THEN current_count ELSE 0 END) as visits,
    MAX(CASE WHEN type = 'like' THEN current_count ELSE 0 END) as likes,
    MAX(CASE WHEN type = 'dislike' THEN current_count ELSE 0 END) as dislikes
FROM interaction_stats;

-- name: IncrementStat :exec
UPDATE interaction_stats 
SET current_count = current_count + 1 
WHERE type = ?;