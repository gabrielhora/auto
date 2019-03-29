package job

import (
	"auto/internal/form"
	"auto/internal/server"
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

// AssignedServer specifies in what Servers a Job can run
type AssignedServer struct {
	ID int64 `gorm:"type:bigserial;primary_key"`

	Job   Job
	JobID int64 `gorm:"not null;index;type:bigint references job(id)"`

	Server   server.Server
	ServerID int64 `gorm:"not null;index;type:bigint references server(id)"`
}

func Create(db *gorm.DB, f form.Job) (Job, error) {
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
		if err = AssignToServer(tx, job.ID, id); err != nil {
			tx.Rollback()
			return Job{}, err
		}
	}

	tx.Commit()
	return job, err
}

func AssignToServer(db *gorm.DB, jobID, serverID int64) error {
	js := AssignedServer{JobID: jobID, ServerID: serverID}
	return db.Create(&js).Error
}

func IsAssignedToServer(db *gorm.DB, job Job, serverID int64) (bool, error) {
	if job.AnyServer {
		return true, nil
	}

	var js AssignedServer
	err := db.Where("job_id = ? AND server_id = ?", job.ID, serverID).First(&js).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func List(db *gorm.DB) ([]Job, error) {
	var jobs []Job
	err := db.Order("name").Find(&jobs).Error
	return jobs, err
}

func Get(db *gorm.DB, jobID int64) (Job, error) {
	var job Job
	err := db.First(&job, "id = ?", jobID).Error
	if gorm.IsRecordNotFoundError(err) {
		return Job{}, nil
	}
	return job, err
}

func Servers(db *gorm.DB, jobID int64) ([]server.Server, error) {
	var servers []server.Server
	err := db.
		Select(`"server".*`).
		Joins(`inner join "assigned_server" on "assigned_server"."server_id" = "server"."id"`).
		Where(`"assigned_server"."job_id" = ?`, jobID).
		Find(&servers).
		Error
	return servers, err
}

type State int

const (
	Running State = iota
	Success
	Fail
)

func (j State) String() string {
	states := []string{"Running", "Success", "Fail"}
	if j >= 0 && int(j) < len(states) {
		return states[j]
	}
	return ""
}

type Execution struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time

	Job   Job
	JobID int64 `gorm:"not null;index;type:bigint references job(id)"`

	// In which server this job was executed
	Server   server.Server
	ServerID int64 `gorm:"not null;index"`

	StartDate time.Time `gorm:"not null"`
	EndDate   pq.NullTime
	State     State
	Log       string `gorm:"type:text"`
}

func (h Execution) Duration() string {
	if !h.EndDate.Valid {
		return ""
	}
	return h.EndDate.Time.Sub(h.StartDate).String()
}

func Executions(db *gorm.DB, jobID int64) ([]Execution, error) {
	var ex []Execution
	err := db.
		Where("job_id = ?", jobID).
		Preload("Server").
		Find(&ex).
		Error
	return ex, err
}

func CreateExecution(db *gorm.DB, jobID, serverID int64) (Execution, error) {
	ex := Execution{
		JobID:     jobID,
		ServerID:  serverID,
		State:     Running,
		StartDate: time.Now().UTC(),
	}
	err := db.Create(&ex).Error
	return ex, err
}

func ExecutionLog(db *gorm.DB, ex *Execution, state State, msg string, args ...interface{}) error {
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
