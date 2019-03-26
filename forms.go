package main

import (
	"github.com/gorilla/schema"
	"net/http"
	"net/url"
)

var schemaDecoder = schema.NewDecoder()

type jobForm struct {
	Name        string
	Description string
	Shell       string
	Script      string
	Servers     []int64
	AnyServer   bool `schema:"-"`
}

func newJobForm(r *http.Request) (*jobForm, url.Values, error) {
	f := &jobForm{}
	errs := url.Values{}

	if err := r.ParseForm(); err != nil {
		return nil, nil, err
	}
	if err := schemaDecoder.Decode(f, r.PostForm); err != nil {
		return nil, nil, err
	}
	for _, s := range f.Servers {
		if s == -1 {
			f.Servers = []int64{}
			f.AnyServer = true
			break
		}
	}

	if f.Name == "" {
		errs.Add("Name", "Field is required")
	}
	if f.Shell == "" {
		errs.Add("Shell", "Field is required")
	}
	if f.Script == "" {
		errs.Add("Script", "Field is required")
	}
	if len(f.Servers) == 0 && !f.AnyServer {
		errs.Add("Servers", "Field is required")
	}

	return f, errs, nil
}
