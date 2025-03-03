package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testRevoluttera/logger"
	"testing"
	"time"
)

// MockLogger is a simple implementation of logger.Logger for testing
type MockLogger struct{}

func (l *MockLogger) Info(msg string)  {}
func (l *MockLogger) Error(msg string) {}

// TestNewJobScheduler verifies JobScheduler creation
func TestNewJobScheduler(t *testing.T) {
	newLogger := logger.NewLogger("taskScheduler.log", "INFO")
	scheduler := NewJobScheduler(newLogger)
	assert.NotNil(t, scheduler)
	assert.NotNil(t, scheduler.jobs)
	assert.NotNil(t, scheduler.stopChan)
	assert.Equal(t, newLogger, scheduler.logger)
}

// TestCreateJob verifies job creation logic
func TestCreateJob(t *testing.T) {
	newLogger := logger.NewLogger("taskScheduler.log", "INFO")
	scheduler := NewJobScheduler(newLogger)

	// Valid Job
	executeAt := time.Now().Add(1 * time.Hour)
	job, err := scheduler.CreateJob("Test Job", executeAt)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "Test Job", job.Description)
	assert.Equal(t, scheduled, job.Status)

	// Invalid Timing
	executeAtPast := time.Now().Add(-1 * time.Hour)
	job, err = scheduler.CreateJob("Past Job", executeAtPast)
	assert.Error(t, err)
	assert.Equal(t, "time must be in future", err.Error())
	assert.Nil(t, job)
}

// TestGetJobs retrieves a list of all jobs
func TestGetJobs(t *testing.T) {
	newLogger := logger.NewLogger("taskScheduler.log", "INFO")
	scheduler := NewJobScheduler(newLogger)

	// Add jobs
	scheduler.CreateJob("Job 1", time.Now().Add(1*time.Hour))
	scheduler.CreateJob("Job 2", time.Now().Add(2*time.Hour))

	jobs := scheduler.GetJobs()
	assert.Len(t, jobs, 2)
}

// TestGetJob fetch job by ID
func TestGetJob(t *testing.T) {
	newLogger := logger.NewLogger("taskScheduler.log", "INFO")
	scheduler := NewJobScheduler(newLogger)

	job, _ := scheduler.CreateJob("Job 1", time.Now().Add(1*time.Hour))

	retrievedJob, err := scheduler.GetJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)

	// Non-existent Job
	_, err = scheduler.GetJob("invalid-id")
	assert.Error(t, err)
	assert.Equal(t, "job not finded", err.Error())
}

// TestCancelJob cancel an existing job
func TestCancelJob(t *testing.T) {
	newLogger := logger.NewLogger("taskScheduler.log", "INFO")
	scheduler := NewJobScheduler(newLogger)

	job, _ := scheduler.CreateJob("Job 1", time.Now().Add(1*time.Hour))

	// Cancel job
	err := scheduler.CancelJob(job.ID)
	assert.NoError(t, err)

	// Verify cancellation
	_, err = scheduler.GetJob(job.ID)
	assert.Error(t, err)

	// Cancel non-existent job
	err = scheduler.CancelJob("non-existent-id")
	assert.Error(t, err)

	// Cancel already executed job
	job.Status = executed
	err = scheduler.CancelJob(job.ID)
	assert.Error(t, err)
}

// Test HTTP Routes
func TestJobSchedulerRoutes(t *testing.T) {
	newLogger := logger.NewLogger("taskScheduler.log", "INFO")
	scheduler := NewJobScheduler(newLogger)
	router := mux.NewRouter()

	scheduler.SetupRoutes(router)

	t.Run("POST /jobs - Create Job", func(t *testing.T) {
		executeAt := time.Now().Add(1 * time.Hour)
		jobData := map[string]interface{}{
			"description": "Test Job",
			"execute_at":  executeAt,
		}
		body, _ := json.Marshal(jobData)

		req, _ := http.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var job Job
		err := json.Unmarshal(resp.Body.Bytes(), &job)
		assert.NoError(t, err)
		assert.Equal(t, "Test Job", job.Description)
	})

	t.Run("GET /jobs - List Jobs", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/jobs", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var jobs []Job
		err := json.Unmarshal(resp.Body.Bytes(), &jobs)
		assert.NoError(t, err)
		assert.NotEmpty(t, jobs)
	})

	t.Run("GET /jobs/{id} - Get Job", func(t *testing.T) {
		executeAt := time.Now().Add(1 * time.Hour)
		job, _ := scheduler.CreateJob("Get Job", executeAt)

		req, _ := http.NewRequest(http.MethodGet, "/jobs/"+job.ID, nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var retrievedJob Job
		err := json.Unmarshal(resp.Body.Bytes(), &retrievedJob)
		assert.NoError(t, err)
		assert.Equal(t, job.ID, retrievedJob.ID)
	})

	t.Run("DELETE /jobs/{id} - Cancel Job", func(t *testing.T) {
		executeAt := time.Now().Add(1 * time.Hour)
		job, _ := scheduler.CreateJob("Delete Job", executeAt)

		req, _ := http.NewRequest(http.MethodDelete, "/jobs/"+job.ID, nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNoContent, resp.Code)

		// Verify job cancellation
		_, err := scheduler.GetJob(job.ID)
		assert.Error(t, err)
	})
}
