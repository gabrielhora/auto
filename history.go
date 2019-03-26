package main

import (
	"github.com/lib/pq"
	"time"
)

type JobHistory struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time

	Job   Job
	JobID int64 `gorm:"not null;index"`

	// In which server this job was executed
	Server   Server
	ServerID int64 `gorm:"not null;index"`

	StartDate time.Time `gorm:"not null"`
	EndDate   pq.NullTime

	// True if script exit code is 0
	Success bool

	// Shell output log
	Log string `gorm:"type:text"`
}
