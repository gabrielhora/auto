package main

import (
	"time"
)

type Server struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time
	Hostname  string `gorm:"not null"`
}
