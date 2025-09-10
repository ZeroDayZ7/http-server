CREATE TABLE interactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ip VARCHAR(45),
    user_id BIGINT NULL,
    type ENUM('visit','like','dislike','comment') NOT NULL,
    value INT DEFAULT 0,
    content TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
