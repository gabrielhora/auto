package main

import "time"

type Queue struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time
	Job       Job
	JobID     int64     `gorm:"not null"`
	Date      time.Time `gorm:"not null;index"`
}
