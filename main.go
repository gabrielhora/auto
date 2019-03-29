package main

import (
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
	db.AutoMigrate(&Server{}, &Job{}, &JobServer{}, &JobExecution{}, &Queue{})

	tpl := template.Must(template.ParseGlob("templates/**/*"))

	// register the server if not registered yet
	server, err := serverRegisterSelf(db)
	if err != nil {
		log.Fatalf("error registering server: %v", err)
	}

	// start the scheduler background job for this server
	go schedulerRun(db, server)

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
