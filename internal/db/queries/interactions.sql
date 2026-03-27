-- name: IncrementStatByAmount :exec
INSERT INTO interaction_stats (type, current_count)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE current_count = current_count + VALUES(current_count);

-- name: GetStats :one
SELECT 
    CAST(COALESCE(SUM(CASE WHEN type = 'visit' THEN current_count ELSE 0 END), 0) AS SIGNED) as visits,
    CAST(COALESCE(SUM(CASE WHEN type = 'like' THEN current_count ELSE 0 END), 0) AS SIGNED) as likes,
    CAST(COALESCE(SUM(CASE WHEN type = 'dislike' THEN current_count ELSE 0 END), 0) AS SIGNED) as dislikes
FROM interaction_stats;

-- name: IncrementStat :exec
INSERT INTO interaction_stats (type, current_count)
VALUES (?, 1)
ON DUPLICATE KEY UPDATE current_count = current_count + 1;