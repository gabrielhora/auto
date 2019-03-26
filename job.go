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
	AnyServer bool `gorm:"not null"`
}

// JobServer specifies in what Servers a Job can run
type JobServer struct {
	ID int64 `gorm:"type:bigserial;primary_key"`

	Job   Job
	JobID int64 `gorm:"not null;index;type:bigint references job(id)"`

	Server   Server
	ServerID int64 `gorm:"not null;index;type:bigint references server(id)"`
}

type JobHistory struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time

	Job   Job
	JobID int64 `gorm:"not null;index;type:bigint references job(id)"`

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

func jobCreate(db *gorm.DB, f *jobForm) (*Job, error) {
	tx := db.Begin()

	job := &Job{
		Name:        f.Name,
		Description: &f.Description,
		Shell:       f.Shell,
		Script:      f.Script,
		AnyServer:   f.AnyServer,
	}

	var err error
	if err = tx.Create(job).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, id := range f.Servers {
		if err = jobAssignToServer(tx, id, job.ID); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()
	return job, err
}

func jobAssignToServer(db *gorm.DB, serverID, jobID int64) error {
	js := &JobServer{JobID: jobID, ServerID: serverID}
	if err := db.Create(js).Error; err != nil {
		return err
	}
	return nil
}
