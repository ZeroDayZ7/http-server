CREATE TABLE interaction_stats (
    type VARCHAR(50) NOT NULL PRIMARY KEY, -- 'like', 'dislike', 'visit'
    current_count BIGINT NOT NULL DEFAULT 0
);

-- Inicjalizacja (robisz to raz)
INSERT INTO interaction_stats (type, current_count) VALUES ('like', 0), ('dislike', 0), ('visit', 0);