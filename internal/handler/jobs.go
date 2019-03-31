package handler

import (
	"auto/internal/form"
	"auto/internal/job"
	"auto/internal/queue"
	"auto/internal/server"
	"fmt"
	"github.com/jinzhu/gorm"
	"html/template"
	"net/http"
	"time"
)

func JobList(db *gorm.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := job.List(db)
		if err != nil {
			internalServerError(w, err)
			return
		}
		data := map[string]interface{}{
			"title": "Jobs",
			"jobs":  jobs,
		}
		tpl.ExecuteTemplate(w, "jobs/list", data)
	}
}

func JobShow(db *gorm.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := int64Param(r, "id")
		j, err := job.Get(db, id)

		if err != nil {
			internalServerError(w, err)
			return
		}

		if j.ID == 0 {
			http.NotFound(w, r)
			return
		}

		servers, err := job.Servers(db, j.ID)
		if err != nil {
			internalServerError(w, err)
			return
		}

		executions, err := job.Executions(db, j.ID)
		if err != nil {
			internalServerError(w, err)
			return
		}

		data := map[string]interface{}{
			"title":   j.Name,
			"job":     j,
			"servers": servers,
			"history": executions,
		}
		tpl.ExecuteTemplate(w, "jobs/show", data)
	}
}

func JobCreate(db *gorm.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := server.List(db)
		if err != nil {
			internalServerError(w, err)
			return
		}

		data := map[string]interface{}{
			"title":   "Create new job",
			"form":    &form.Job{Shell: "/bin/bash"},
			"errors":  nil,
			"servers": servers,
		}

		if r.Method == "GET" {
			tpl.ExecuteTemplate(w, "jobs/new", data)
			return
		}

		f, verrors, err := form.NewJob(r)
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

		j, err := job.Create(db, f)
		if err != nil {
			internalServerError(w, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/jobs/%d", j.ID), http.StatusFound)
	}
}

func JobRun(db *gorm.DB, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := int64Param(r, "id")
		j, err := job.Get(db, id)
		if err != nil {
			internalServerError(w, err)
			return
		}

		data := map[string]interface{}{
			"title": j.Name,
			"job":   j,
		}

		if r.Method == "GET" {
			tpl.ExecuteTemplate(w, "jobs/run", data)
			return
		}

		if err := queue.ScheduleTo(db, j, time.Now()); err != nil {
			internalServerError(w, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/jobs/%d/", j.ID), http.StatusFound)
	}
}
