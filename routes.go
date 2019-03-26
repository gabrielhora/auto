package main

import (
	"github.com/jinzhu/gorm"
	"html/template"
	"log"
	"net/http"
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

		f, verrs, err := newJobForm(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		if len(verrs) > 0 {
			data["form"] = f
			data["errors"] = verrs
			tpl.ExecuteTemplate(w, "jobs/new", data)
			return
		}

		job, err := jobCreate(db, f)
		if err != nil {
			log.Printf("error creating new job: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		log.Printf(`new job "%s" created`, job.Name)
		http.Redirect(w, r, "/jobs/new/", http.StatusFound)
	}
}
