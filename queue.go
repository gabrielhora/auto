package main

import (
	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
	"log"
	"time"
)

type Queue struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time
	Job       Job
	JobID     int64     `gorm:"not null;index;type:bigint references job(id)"`
	Date      time.Time `gorm:"not null;index"`
}

func queuePendingJobs(db *gorm.DB, serverID int64) ([]Job, error) {
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
		assigned, err := jobIsAssignedToServer(tx, q.Job, serverID)
		if err != nil {
			return nil, err
		}
		if assigned {
			runnable = append(runnable, q)
		}
	}

	// remove from the queue table jobs that will be run
	var idsToDelete []int64
	var jobs []Job
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
		if err = queueScheduleNext(tx, j); err != nil {
			return nil, err
		}
	}

	return jobs, nil
}

func queueScheduleNext(db *gorm.DB, job Job) error {
	if job.Cron == nil {
		log.Printf(`job "%s" do not have a cron expression, will not be scheduled`, job.Name)
		return nil
	}

	s, err := cron.ParseStandard(*job.Cron)
	if err != nil {
		log.Printf(`error parsing cron expression "%s" for job "%s"`, *job.Cron, job.Name)
		return err
	}

	item := Queue{
		JobID: job.ID,
		Date:  s.Next(time.Now().UTC()).UTC(),
	}
	return db.Create(&item).Error
}
