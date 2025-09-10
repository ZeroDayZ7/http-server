package model

import "time"

type Interaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `json:"user_id,omitempty"`
	IP        string    `gorm:"size:45" json:"ip"`
	Type      string    `gorm:"size:20;not null" json:"type"`
	Value     int       `gorm:"default:0" json:"value"`
	Content   *string   `gorm:"type:text" json:"content,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
