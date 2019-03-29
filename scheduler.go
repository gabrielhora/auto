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

	go schedulerMainLoop(db, server)
	return nil
}

func schedulerMainLoop(db *gorm.DB, server Server) {
	sleep := randonDurationBetween(10, 60, time.Second)

	for {
		time.Sleep(sleep)
		log.Printf(`Running scheduler for "%s"...`, server.Hostname)

		jobs, err := queuePendingJobs(db, server.ID)
		if err != nil {
			log.Printf("error getting pending jobs: %v", err)
			continue
		}

		log.Printf(`Found %d jobs to run on "%s"`, len(jobs), server.Hostname)
		for _, job := range jobs {
			go schedulerRunJob(db, job)
		}
	}
}

func schedulerRunJob(db *gorm.DB, job Job) {
	log.Printf(`Running job "%s"`, job.Name)
}

func randonDurationBetween(min, max int, d time.Duration) time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(min+rand.Intn(max-min)) * d
}

/*
func createScriptFile(script string) (string, error) {
	p := path.Join(os.TempDir(), uuid.New().String())
	f, err := os.Create(p)
	defer f.Close()

	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(script); err != nil {
		return "", err
	}
	if err := os.Chmod(p, 0755); err != nil {
		return "", err
	}

	log.Printf(`created file "%s"`, p)
	return p, nil
}

func setup(job *Job) {
	var err error
	job.ScriptFilePath, err = createScriptFile(job.Script)
	if err != nil {
		log.Fatalf("error creating script file: %v", err)
	}
}

func execute(job *Job) {
	out, err := exec.Command(job.Shell, job.ScriptFilePath).CombinedOutput()
	if err != nil {
		log.Fatalf("error executing script: %v", err)
	}
	log.Printf("%s", out)
}

func cleanup(job *Job) {
	err := os.Remove(job.ScriptFilePath)
	if err != nil {
		log.Printf("error deleting script file: %v", err)
	}
}
*/
