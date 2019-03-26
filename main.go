package main

import (
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"time"
)

func main() {
	db, err := gorm.Open("postgres", "host=localhost user=auto password=auto dbname=auto sslmode=disable")
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}
	db.SingularTable(true)
	db.LogMode(true)
	db.AutoMigrate(&Server{}, &Job{}, &JobServer{}, &JobHistory{}, &Queue{})

	tpl := template.Must(template.ParseGlob("templates/**/*"))

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/jobs/", jobListHandler(db, tpl)).Methods("GET")
	router.HandleFunc("/jobs/{id:[0-9]+}/", jobShowHandler(db, tpl)).Methods("GET")
	router.HandleFunc("/jobs/new/", jobCreateHandler(db, tpl)).Methods("GET", "POST")

	hs := &http.Server{
		Addr:         ":8000",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("server running on %s", hs.Addr)
	log.Fatal(hs.ListenAndServe())
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
