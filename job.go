package main

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"time"
)

type Job struct {
	ID          int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt   time.Time
	Name        string  `gorm:"not null"`
	Description *string `gorm:"type:text"`
	Shell       string  `gorm:"not null"`
	Script      string  `gorm:"not null"`

	// Cron expression to determine when this Job is executed
	// If null this job will only run on demand
	Cron *string

	// True if this job can run in any server
	RunnableAny bool `gorm:"not null"`

	// Array of server IDs where this Job can run,
	// if RunnableAny is true, this array is not checked
	RunnableIn pq.Int64Array `gorm:"type:bigint[];not null"`
}

func jobCreate(db *gorm.DB, f *jobForm) (*Job, error) {
	newJob := &Job{
		Name:        f.Name,
		Description: &f.Description,
		Shell:       f.Shell,
		Script:      f.Script,
		RunnableAny: f.AnyServer,
		RunnableIn:  f.Servers,
	}
	err := db.Create(newJob).Error
	return newJob, err
}
