package model

import "time"

type Session struct {
	SessionID string    `gorm:"primaryKey;size:64" json:"session_id"`
	UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	Data      string    `gorm:"type:json" json:"data"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
}
