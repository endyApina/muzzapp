package models

type Decision struct {
	ActorUserID     string `gorm:"primaryKey"`
	RecipientUserID string `gorm:"primaryKey"`
	Liked           bool
	UnixTimestamp   int64
}
