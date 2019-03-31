package main

import (
	"auto/internal/handler"
	"auto/internal/job"
	"auto/internal/queue"
	"auto/internal/scheduler"
	"auto/internal/server"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open("postgres", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}
	db.SingularTable(true)
	db.LogMode(true)
	db.AutoMigrate(&server.Server{}, &job.Job{}, &job.AssignedServer{}, &job.Execution{}, &queue.Queue{})

	tpl := template.Must(template.ParseGlob("templates/**/*"))

	// register the server if not registered yet
	s, err := server.RegisterSelf(db)
	if err != nil {
		log.Fatalf("error registering server: %v", err)
	}

	// start the scheduler background job for this server
	go scheduler.Run(db, s)

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/jobs/", handler.JobList(db, tpl)).Methods("GET")
	router.HandleFunc("/jobs/{id:[0-9]+}/", handler.JobShow(db, tpl)).Methods("GET")
	router.HandleFunc("/jobs/{id:[0-9]+}/run/", handler.JobRun(db, tpl)).Methods("GET", "POST")
	router.HandleFunc("/jobs/new/", handler.JobCreate(db, tpl)).Methods("GET", "POST")

	hs := &http.Server{
		Addr:         ":8000",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("server running on %s", hs.Addr)
	log.Fatal(hs.ListenAndServe())
}
