package form

import (
	"github.com/gorilla/schema"
	"net/http"
	"net/url"
)

var schemaDecoder = schema.NewDecoder()

type Job struct {
	Name        string
	Description string
	Cron        string
	Shell       string
	Setup       string
	Script      string
	Teardown    string
	Servers     []int64
	AnyServer   bool `schema:"-"`
}

func NewJob(r *http.Request) (Job, url.Values, error) {
	f := Job{}
	errs := url.Values{}

	if err := r.ParseForm(); err != nil {
		return Job{}, nil, err
	}
	if err := schemaDecoder.Decode(&f, r.PostForm); err != nil {
		return Job{}, nil, err
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
