CREATE TABLE IF NOT EXISTS interaction_stats (
    type VARCHAR(50) NOT NULL PRIMARY KEY,
    current_count BIGINT NOT NULL DEFAULT 0
);

-- Inicjalizacja liczników na start
INSERT IGNORE INTO interaction_stats (type, current_count) 
VALUES ('like', 0), ('dislike', 0), ('visit', 0);