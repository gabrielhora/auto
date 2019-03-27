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

func queuePending(db *gorm.DB, serverID int64) ([]Job, error) {
	tx := db.Begin()

	// lock the table so only this routine can access it, rollbacks or
	// commits will release this lock, will also timeout in 5 seconds
	err := tx.Exec(`SET statement_timeout = 5000; LOCK TABLE "queue" IN EXCLUSIVE MODE`).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// get all pending items
	var pending []Queue
	err = db.Where("date <= ?", time.Now().UTC()).Preload("Job").Find(&pending).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// get pending items that can run in this server
	var toBeProcessed []Queue
	for _, item := range pending {
		assigned, err := jobIsAssignedToServer(tx, &item.Job, serverID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if assigned {
			toBeProcessed = append(toBeProcessed, item)
		}
	}

	// remove from the queue table jobs that will be processed
	var idsToDelete []int64
	var jobs []Job
	for _, item := range toBeProcessed {
		idsToDelete = append(idsToDelete, item.ID)
		jobs = append(jobs, item.Job)
	}

	if len(idsToDelete) > 0 {
		if err = db.Where("id in (?)", &idsToDelete).Delete(Queue{}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// schedule next execution for selected jobs
	for _, item := range jobs {
		if err := queueScheduleNext(tx, item); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

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

	// todo: make sure this job is not already scheduled

	nextTime := s.Next(time.Now().UTC()).UTC()
	item := Queue{
		JobID: job.ID,
		Date:  nextTime,
	}
	return db.Create(&item).Error
}
