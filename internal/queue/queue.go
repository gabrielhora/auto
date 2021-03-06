package queue

import (
	"auto/internal/job"
	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
	"log"
	"time"
)

type Queue struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time
	Job       job.Job
	JobID     int64     `gorm:"not null;index;type:bigint references job(id)"`
	Date      time.Time `gorm:"not null;index"`
}

func Pending(db *gorm.DB, serverID int64) ([]job.Job, error) {
	tx := db.Begin()

	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// lock the table so only this routine can access it, rollbacks or
	// commits will release this lock, will also timeout in 5 seconds
	if err = tx.Exec(`SET statement_timeout = 5000; LOCK TABLE "queue" IN EXCLUSIVE MODE`).Error; err != nil {
		return nil, err
	}

	// get all pending jobs
	var pending []Queue
	if err = tx.Where("date <= ?", time.Now().UTC()).Preload("Job").Find(&pending).Error; err != nil {
		return nil, err
	}

	// collect pending jobs that can run in this server
	var runnable []Queue
	for _, q := range pending {
		assigned, err := job.IsAssignedToServer(tx, q.Job, serverID)
		if err != nil {
			return nil, err
		}
		if assigned {
			runnable = append(runnable, q)
		}
	}

	// remove from the queue table jobs that will be run
	var idsToDelete []int64
	var jobs []job.Job
	for _, q := range runnable {
		idsToDelete = append(idsToDelete, q.ID)
		jobs = append(jobs, q.Job)
	}
	if len(idsToDelete) > 0 {
		if err = tx.Where(idsToDelete).Delete(Queue{}).Error; err != nil {
			return nil, err
		}
	}

	// schedule next execution for selected jobs
	for _, j := range jobs {
		if err = ScheduleNext(tx, j); err != nil {
			return nil, err
		}
	}

	return jobs, nil
}

func ScheduleNext(db *gorm.DB, job job.Job) error {
	if job.Cron == nil || *job.Cron == "" {
		log.Printf(`job "%s" do not have a cron expression, will not be scheduled`, job.Name)
		return nil
	}

	s, err := cron.ParseStandard(*job.Cron)
	if err != nil {
		log.Printf(`error parsing cron expression "%s" for job "%s"`, *job.Cron, job.Name)
		return err
	}

	t := s.Next(time.Now().UTC())
	return ScheduleTo(db, job, t)
}

func ScheduleTo(db *gorm.DB, j job.Job, t time.Time) error {
	item := Queue{JobID: j.ID, Date: t.UTC()}
	return db.Create(&item).Error
}
