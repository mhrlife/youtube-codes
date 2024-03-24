package main

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	c := make(chan struct{})
	go ImageProcessingServer()
	go AppServer()
	<-c
}

/**
This part is for web server definitions
*/

func ImageProcessingServer() {
	addr := "localhost:8081"
	jobs := NewJobScheduler()
	e := echo.New()

	type jobRequest struct {
		SomeInfo string `json:"some_info"`
	}
	// this route creates a new job or if correlation id already exists, fetches that job
	e.POST("/job", func(c echo.Context) error {
		// extract post data
		var request jobRequest
		if err := c.Bind(&request); err != nil {
			return err
		}
		correlationId := c.Request().Header.Get("X-Correlation-ID")
		if correlationId == "" {
			return c.String(401, "bad correlation id")
		}
		// check whether the request already exists
		_ = jobs.CreateOrGet(correlationId, nil)
		return c.String(202, "http://localhost:8081/job/"+correlationId+"/status")
	})

	type jobRequestWithCallback struct {
		jobRequest
		CallbackURL string `json:"callback_url"`
	}
	// this route creates a new job or if correlation id already exists, fetches that job
	e.POST("/job_callback", func(c echo.Context) error {
		// extract post data
		var request jobRequestWithCallback
		if err := c.Bind(&request); err != nil {
			return err
		}
		correlationId := c.Request().Header.Get("X-Correlation-ID")
		if correlationId == "" {
			return c.String(401, "bad correlation id")
		}
		// check whether the request already exists
		_ = jobs.CreateOrGet(correlationId, func(j Job) {
			_, _, err := post(request.CallbackURL, map[string]any{
				"job": j,
			}, nil)
			if err != nil {
				log.Printf("error while answering the webhook: %v\n", err)
			}
		})
		return c.String(202, "accepted")
	})

	// this route responds with the info the requested job
	e.GET("/job/:jobId/status", func(c echo.Context) error {
		jobId := c.Param("jobId")
		job, ok := jobs.Get(jobId)
		if !ok {
			return c.String(404, "job not found")
		}
		return c.JSON(200, job)
	})
	e.Logger.Error(e.Start(addr))
}

func AppServer() {
	addr := "localhost:8080"
	e := echo.New()

	e.GET("/process", func(c echo.Context) error {
		// doing all business logics here, auth, balance, etc.
		// when everything is ok, dispatches a new job
		correlationId := uuid.New().String()
		code, resp, err := post("http://localhost:8081/job", map[string]any{
			"some_info": "hello world",
		}, map[string]string{
			"X-Correlation-ID": correlationId,
		})
		if err != nil {
			return c.String(500, err.Error())
		}
		return c.String(code, string(resp))
	})
	e.GET("/internal-route", func(c echo.Context) error {
		// doing all business logics here, auth, balance, etc.
		// when everything is ok, dispatches a new job
		correlationId := uuid.New().String()
		code, resp, err := post("http://localhost:8081/job_callback", map[string]any{
			"some_info":    "hello world",
			"callback_url": "http://localhost:8080/callback/" + correlationId,
		}, map[string]string{
			"X-Correlation-ID": correlationId,
		})
		if err != nil {
			return c.String(500, err.Error())
		}
		return c.String(code, string(resp))
	})

	e.POST("/callback/:correlationId", func(c echo.Context) error {
		// doing all business logics here, auth, balance, etc.
		// when everything is ok, dispatches a new job
		correlationId := c.Param("correlationId")
		job := make(map[string]any)
		if err := c.Bind(&job); err != nil {
			return c.String(500, err.Error())
		}
		// now do other business logics
		log.Printf("job done! id=%v, data=%v\n", correlationId, job)
		return c.String(200, "accepted")
	})
	e.Logger.Error(e.Start(addr))
}

/**
This part is for managing jobs
*/

type Job struct {
	State string `json:"state"`
	Value string `json:"value,omitempty"`
}

type JobScheduler struct {
	jobs map[string]*Job
	lock sync.RWMutex
}

func NewJobScheduler() *JobScheduler {
	return &JobScheduler{jobs: make(map[string]*Job)}
}

func (l *JobScheduler) Get(id string) (*Job, bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	job, ok := l.jobs[id]
	return job, ok
}

func (l *JobScheduler) CreateOrGet(id string, onCompleted func(j Job)) *Job {
	l.lock.Lock()
	job, ok := l.jobs[id]
	if !ok {
		job = &Job{State: "created"}
		l.jobs[id] = job
	}
	l.lock.Unlock()

	go func() {
		<-time.After(time.Second * 3)
		job.State = "processing"
		<-time.After(time.Second * 3)
		job.State = "done"

		if onCompleted != nil {
			onCompleted(*job)
		}
	}()
	return job
}

func post(url string, data map[string]any, headers map[string]string) (int, []byte, error) {
	// json marshal the body
	b, err := json.Marshal(data)
	if err != nil {
		return 0, nil, err
	}
	// preparing the post request
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return 0, nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// sending the http request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, body, nil
}
