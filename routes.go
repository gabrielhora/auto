package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func jobListHandler(db *gorm.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := jobList(db)
		if err != nil {
			log.Printf("error getting job list: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"title": "Jobs",
			"jobs":  jobs,
		}
		tpl.ExecuteTemplate(w, "jobs/list", data)
	}
}

func jobShowHandler(db *gorm.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id, _ := strconv.ParseInt(params["id"], 10, 64)
		job, err := jobGet(db, id)

		if err != nil {
			log.Printf("error getting job with id %d: %v", id, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if job.ID == 0 {
			http.NotFound(w, r)
			return
		}

		servers, err := jobServers(db, job.ID)
		if err != nil {
			log.Printf("error getting job servers: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		history, err := jobHistory(db, job.ID)
		if err != nil {
			log.Printf("error getting job history: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"title":   job.Name,
			"job":     job,
			"servers": servers,
			"history": history,
		}
		tpl.ExecuteTemplate(w, "jobs/show", data)
	}
}

func jobCreateHandler(db *gorm.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := serverList(db)
		if err != nil {
			log.Printf("error getting server list: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"title":   "Create new job",
			"form":    &jobForm{},
			"errors":  nil,
			"servers": servers,
		}

		if r.Method == "GET" {
			tpl.ExecuteTemplate(w, "jobs/new", data)
			return
		}

		f, verrors, err := newJobForm(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		if len(verrors) > 0 {
			data["form"] = f
			data["errors"] = verrors
			tpl.ExecuteTemplate(w, "jobs/new", data)
			return
		}

		job, err := jobCreate(db, f)
		if err != nil {
			log.Printf("error creating new job: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/jobs/%d", job.ID), http.StatusFound)
	}
}
