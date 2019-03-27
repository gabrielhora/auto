package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"math/rand"
	"os"
	"time"
)

func schedulerRun(db *gorm.DB) error {
	sleep := _randDurationBetween(10, 60, time.Second)
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	server, err := serverGet(db, hostname)
	if err != nil {
		return err
	}
	if server.ID == 0 {
		return fmt.Errorf("could not find server with hostname %s", hostname)
	}

	log.Printf(`Scheduler for server "%s" will run every %s`, server.Hostname, sleep)

	go func() {
		for {
			time.Sleep(sleep)
			log.Printf(`Running scheduler for "%s"...`, server.Hostname)

			jobs, err := queuePending(db, server.ID)

			if err != nil {
				log.Printf("error getting pending jobs: %v", err)
			} else {
				log.Printf(`Found %d jobs to run on "%s"`, len(jobs), server.Hostname)

				for _, job := range jobs {
					go schedulerRunJob(job)
				}
			}
		}
	}()

	return nil
}

func schedulerRunJob(job Job) {
	log.Printf(`Running job "%s"`, job.Name)
}

func _randDurationBetween(min, max int, d time.Duration) time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(min+rand.Intn(max-min)) * d
}
