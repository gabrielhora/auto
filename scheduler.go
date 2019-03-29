package main

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"time"
)

func schedulerRun(db *gorm.DB, server Server) {
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
			go schedulerJobRun(db, job, server)
		}
	}
}

func schedulerJobRun(db *gorm.DB, job Job, server Server) {
	log.Printf(`Running job "%s"`, job.Name)

	ex, err := jobExecutionNew(db, job.ID, server.ID)
	if err != nil {
		log.Printf(`could not create job execution for "%s": %v`, job.Name, err)
		return
	}

	// always run setup script
	if err := runScript(db, &ex, job.Shell, job.Setup); err != nil {
		log.Printf("error executing setup script: %v", err)
	}

	// only run the main script if setup didn't fail
	if ex.State == Running {
		if err := runScript(db, &ex, job.Shell, job.Script); err != nil {
			log.Printf("error executing main script: %v", err)
		}
	}

	// always run teardown script
	if err := runScript(db, &ex, job.Shell, job.Teardown); err != nil {
		log.Printf("error executing teardown script: %v", err)
	}

	// if we get to here with state == running it means everything worked
	// if the status == fail some of the script did not run properly
	if ex.State == Running {
		if err := jobExecutionLog(db, &ex, Success, "DONE"); err != nil {
			log.Printf(`error updating "%s" state to Success: %v`, job.Name, err)
		}
	}
}

// runScript executes a job script in it's shell updating the JobHistory with the execution log
// and later deleting the generated temporary file
func runScript(db *gorm.DB, history *JobExecution, shell, script string) error {
	p, err := createTempFile(script)
	if err != nil {
		_ = jobExecutionLog(db, history, Fail, "error creating file: %v", err)
		return err
	}
	defer func() {
		if err := os.Remove(p); err != nil {
			log.Printf(`could not delete temp file "%s": %v`, p, err)
		}
	}()

	out, err := exec.Command(shell, p).CombinedOutput()
	if err != nil {
		_ = jobExecutionLog(db, history, Fail, "%s\n\nERROR: %v", out, err)
		return err
	}

	return jobExecutionLog(db, history, Running, string(out))
}

// createTempFile creates a temporary executable (755) file containing the script
// passed in as parameter. Returns the full path of the new file.
func createTempFile(script string) (string, error) {
	p := path.Join(os.TempDir(), uuid.New().String())

	f, err := os.Create(p)
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("error closing job script file: %v", err)
		}
	}()

	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(script); err != nil {
		return "", err
	}
	if err := os.Chmod(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}

func randonDurationBetween(min, max int, d time.Duration) time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(min+rand.Intn(max-min)) * d
}
