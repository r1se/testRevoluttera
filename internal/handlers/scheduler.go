package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
	"time"
	"user_db_test/logger"
)

// Job struct
type Job struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	ExecuteAt   time.Time  `json:"execute_at"`
	Status      string     `json:"status"`
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
}

// JobScheduler struct for manipulate jobs
type JobScheduler struct {
	jobs     map[string]*Job
	jobLock  sync.RWMutex
	stopChan chan bool
	logger   logger.Logger
}

const (
	scheduled = "scheduled"
	executing = "executing"
	executed  = "executed"
	cancelled = "cancelled"
)

// NewJobScheduler create scheduler
func NewJobScheduler(logger logger.Logger) *JobScheduler {
	return &JobScheduler{
		jobs:     make(map[string]*Job),
		stopChan: make(chan bool),
		logger:   logger,
	}
}

// CreateJob create job
func (s *JobScheduler) CreateJob(description string, executeAt time.Time) (*Job, error) {
	if executeAt.Before(time.Now()) {
		return nil, fmt.Errorf("time must be in future")
	}

	jobID := uuid.New().String()
	job := &Job{
		ID:          jobID,
		Description: description,
		ExecuteAt:   executeAt,
		Status:      scheduled,
	}

	s.jobLock.Lock()
	defer s.jobLock.Unlock()
	s.jobs[jobID] = job

	// Start monitor job
	go s.monitorJob(job)

	return job, nil
}

// GetJobs return jobs list
func (s *JobScheduler) GetJobs() []*Job {
	s.jobLock.RLock()
	defer s.jobLock.RUnlock()

	var jobs []*Job
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetJob return job by id
func (s *JobScheduler) GetJob(id string) (*Job, error) {
	s.jobLock.RLock()
	defer s.jobLock.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not finded")
	}
	return job, nil
}

// CancelJob cancel job
func (s *JobScheduler) CancelJob(id string) error {
	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job not finded")
	}

	if job.Status == executed || job.Status == cancelled {
		return fmt.Errorf("can't cancel completed or canceled job")
	}

	job.Status = cancelled
	delete(s.jobs, id)
	return nil
}

// monitorJob get job status
func (s *JobScheduler) monitorJob(job *Job) {
	<-time.After(job.ExecuteAt.Sub(time.Now()))

	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	if job.Status == cancelled {
		return
	}

	job.Status = executing
	time.Sleep(2 * time.Second) // Sleep for simulate
	job.Status = executed
	executedAt := time.Now()
	job.ExecutedAt = &executedAt
}

// SetupRoutes set http path
func (s *JobScheduler) SetupRoutes(router *mux.Router) {
	router.HandleFunc("/jobs", s.handleCreateJob).Methods(http.MethodPost)
	router.HandleFunc("/jobs", s.handleGetJobs).Methods(http.MethodGet)
	router.HandleFunc("/jobs/{id}", s.handleGetJob).Methods(http.MethodGet)
	router.HandleFunc("/jobs/{id}", s.handleCancelJob).Methods(http.MethodDelete)
}

func (s *JobScheduler) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var jobData struct {
		Description string    `json:"description"`
		ExecuteAt   time.Time `json:"execute_at"`
	}

	err := json.NewDecoder(r.Body).Decode(&jobData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := s.CreateJob(jobData.Description, jobData.ExecuteAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (s *JobScheduler) handleGetJobs(w http.ResponseWriter, r *http.Request) {
	jobs := s.GetJobs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (s *JobScheduler) handleGetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	job, err := s.GetJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (s *JobScheduler) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	err := s.CancelJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
