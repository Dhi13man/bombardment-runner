package services

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a bombardment job
type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
)

// JobSnapshot is the JSON-serializable view of a job
type JobSnapshot struct {
	ID              string     `json:"id"`
	Status          JobStatus  `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	TotalRows       int64      `json:"total_rows"`
	ProcessedRows   int64      `json:"processed_rows"`
	FailedRows      int64      `json:"failed_rows"`
	ProgressPercent float64    `json:"progress_percent"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

// Job is the thread-safe mutable job that tracks bombardment progress
type Job struct {
	ID          string
	Status      JobStatus
	CreatedAt   time.Time
	CompletedAt *time.Time
	ErrorMessage string

	processedAtomic atomic.Int64
	failedAtomic    atomic.Int64
	totalAtomic     atomic.Int64
	mu              sync.Mutex
}

// IncrementProcessed atomically increments the processed row count
func (j *Job) IncrementProcessed() {
	j.processedAtomic.Add(1)
}

// IncrementFailed atomically increments the failed row count
func (j *Job) IncrementFailed() {
	j.failedAtomic.Add(1)
}

// SetTotal sets the total row count for progress calculation
func (j *Job) SetTotal(n int64) {
	j.totalAtomic.Store(n)
}

// SetRunning marks the job as running
func (j *Job) SetRunning() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = JobStatusRunning
}

// Complete marks the job as completed
func (j *Job) Complete() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = JobStatusCompleted
	now := time.Now()
	j.CompletedAt = &now
}

// Fail marks the job as failed with an error message
func (j *Job) Fail(errMsg string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = JobStatusFailed
	j.ErrorMessage = errMsg
	now := time.Now()
	j.CompletedAt = &now
}

// Snapshot returns a JSON-safe copy of the job's current state
func (j *Job) Snapshot() JobSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()

	processed := j.processedAtomic.Load()
	failed := j.failedAtomic.Load()
	total := j.totalAtomic.Load()
	var progress float64
	if total > 0 {
		progress = float64(processed+failed) / float64(total) * 100
	}

	return JobSnapshot{
		ID:              j.ID,
		Status:          j.Status,
		CreatedAt:       j.CreatedAt,
		CompletedAt:     j.CompletedAt,
		TotalRows:       total,
		ProcessedRows:   processed,
		FailedRows:      failed,
		ProgressPercent: progress,
		ErrorMessage:    j.ErrorMessage,
	}
}

// JobStore is a thread-safe in-memory store for bombardment jobs
type JobStore struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// NewJobStore creates a new in-memory job store
func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

// Create creates a new pending job and returns it
func (s *JobStore) Create() *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	job := &Job{
		ID:        uuid.New().String(),
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
	}
	s.jobs[job.ID] = job
	return job
}

// Get retrieves a job by ID, returning a snapshot
func (s *JobStore) Get(id string) (JobSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return JobSnapshot{}, false
	}
	return job.Snapshot(), true
}

// List returns snapshots of all jobs, sorted by creation time (newest first)
func (s *JobStore) List() []JobSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jobs := make([]JobSnapshot, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j.Snapshot())
	}
	sort.Slice(jobs, func(i, k int) bool {
		return jobs[i].CreatedAt.After(jobs[k].CreatedAt)
	})
	return jobs
}
