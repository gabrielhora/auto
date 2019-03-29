package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"log"
	"time"
)

type Job struct {
	ID          int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt   time.Time
	Name        string  `gorm:"not null"`
	Description *string `gorm:"type:text"`
	Shell       string  `gorm:"not null"`

	Setup    string `gorm:"type:text;not null"`
	Script   string `gorm:"type:text;not null"`
	Teardown string `gorm:"type:text;not null"`

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

func jobCreate(db *gorm.DB, f jobForm) (Job, error) {
	tx := db.Begin()

	job := Job{
		Name:        f.Name,
		Description: &f.Description,
		Shell:       f.Shell,
		Script:      f.Script,
		AnyServer:   f.AnyServer,
	}

	var err error
	if err = tx.Create(&job).Error; err != nil {
		tx.Rollback()
		return Job{}, err
	}

	for _, id := range f.Servers {
		if err = jobAssignToServer(tx, job.ID, id); err != nil {
			tx.Rollback()
			return Job{}, err
		}
	}

	tx.Commit()
	return job, err
}

func jobAssignToServer(db *gorm.DB, jobID, serverID int64) error {
	js := JobServer{JobID: jobID, ServerID: serverID}
	return db.Create(&js).Error
}

func jobIsAssignedToServer(db *gorm.DB, job Job, serverID int64) (bool, error) {
	if job.AnyServer {
		return true, nil
	}

	var js JobServer
	err := db.Where("job_id = ? AND server_id = ?", job.ID, serverID).First(&js).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func jobList(db *gorm.DB) ([]Job, error) {
	var jobs []Job
	err := db.Order("name").Find(&jobs).Error
	return jobs, err
}

func jobGet(db *gorm.DB, jobID int64) (Job, error) {
	var job Job
	err := db.First(&job, "id = ?", jobID).Error
	if gorm.IsRecordNotFoundError(err) {
		return Job{}, nil
	}
	return job, err
}

func jobServers(db *gorm.DB, jobID int64) ([]Server, error) {
	var servers []Server
	err := db.
		Select(`"server".*`).
		Joins(`inner join "job_server" on "job_server"."server_id" = "server"."id"`).
		Where(`"job_server"."job_id" = ?`, jobID).
		Find(&servers).
		Error
	return servers, err
}

type JobState int

const (
	Running JobState = iota
	Success
	Fail
)

func (j JobState) String() string {
	states := []string{"Running", "Success", "Fail"}
	if j >= 0 && int(j) < len(states) {
		return states[j]
	}
	return ""
}

type JobExecution struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time

	Job   Job
	JobID int64 `gorm:"not null;index;type:bigint references job(id)"`

	// In which server this job was executed
	Server   Server
	ServerID int64 `gorm:"not null;index"`

	StartDate time.Time `gorm:"not null"`
	EndDate   pq.NullTime
	State     JobState
	Log       string `gorm:"type:text"`
}

func (h JobExecution) Duration() string {
	if !h.EndDate.Valid {
		return ""
	}
	return h.EndDate.Time.Sub(h.StartDate).String()
}

func jobExecutions(db *gorm.DB, jobID int64) ([]JobExecution, error) {
	var ex []JobExecution
	err := db.
		Where("job_id = ?", jobID).
		Preload("Server").
		Find(&ex).
		Error
	return ex, err
}

func jobExecutionNew(db *gorm.DB, jobID, serverID int64) (JobExecution, error) {
	ex := JobExecution{
		JobID:     jobID,
		ServerID:  serverID,
		State:     Running,
		StartDate: time.Now().UTC(),
	}
	err := db.Create(&ex).Error
	return ex, err
}

func jobExecutionLog(db *gorm.DB, ex *JobExecution, state JobState, msg string, args ...interface{}) error {
	ex.State = state
	if msg != "" {
		ex.Log = fmt.Sprintf(msg, args...)
	}
	if err := db.Save(&ex).Error; err != nil {
		log.Printf(`error job's updating log: %v`, err)
		return err
	}
	return nil
}
