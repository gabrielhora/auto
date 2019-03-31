package scheduler

import (
	"auto/internal/job"
	"auto/internal/queue"
	"auto/internal/server"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"time"
)

func Run(db *gorm.DB, server server.Server) {
	sleep := randonDurationBetween(5, 10, time.Second)

	for {
		time.Sleep(sleep)
		log.Printf(`Running scheduler for "%s"...`, server.Hostname)

		jobs, err := queue.Pending(db, server.ID)
		if err != nil {
			log.Printf("error getting pending jobs: %v", err)
			continue
		}

		log.Printf(`Found %d jobs to run on "%s"`, len(jobs), server.Hostname)
		for _, j := range jobs {
			go runJob(db, j, server)
		}
	}
}

func runJob(db *gorm.DB, j job.Job, server server.Server) {
	log.Printf(`Running job "%s"`, j.Name)

	ex, err := job.CreateExecution(db, j.ID, server.ID)
	if err != nil {
		log.Printf(`could not create job execution for "%s": %v`, j.Name, err)
		return
	}

	// always run setup script
	if err := runScript(db, &ex, j.Shell, j.Setup); err != nil {
		log.Printf("error executing setup script: %v", err)
	}

	// only run the main script if setup didn't fail
	if ex.State == job.Running {
		if err := runScript(db, &ex, j.Shell, j.Script); err != nil {
			log.Printf("error executing main script: %v", err)
		}
	}

	// always run teardown script
	if err := runScript(db, &ex, j.Shell, j.Teardown); err != nil {
		log.Printf("error executing teardown script: %v", err)
	}

	// if we get to here with state == running it means everything worked
	// if the status == fail some of the script did not run properly
	if ex.State == job.Running {
		if err := job.ExecutionLog(db, &ex, job.Success, "DONE"); err != nil {
			log.Printf(`error updating "%s" state to Success: %v`, j.Name, err)
		}
	}
}

// runScript executes a job script in it's shell updating the JobHistory with the execution log
// and later deleting the generated temporary file
func runScript(db *gorm.DB, ex *job.Execution, shell, script string) error {
	if script == "" {
		log.Print("script is empty, nothing to do")
		return nil
	}

	p, err := createTempFile(script)
	if err != nil {
		_ = job.ExecutionLog(db, ex, job.Fail, "error creating file: %v", err)
		return err
	}
	defer func() {
		if err := os.Remove(p); err != nil {
			log.Printf(`could not delete temp file "%s": %v`, p, err)
		}
	}()

	out, err := exec.Command(shell, p).CombinedOutput()
	if err != nil {
		_ = job.ExecutionLog(db, ex, job.Fail, "%s\n\nERROR: %v", out, err)
		return err
	}

	return job.ExecutionLog(db, ex, job.Running, string(out))
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
