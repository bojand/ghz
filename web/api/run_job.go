package api

import (
    "sync"
    "time"
    "github.com/google/uuid"
)

// RunJobStatus represents lifecycle state of a run job
type RunJobStatus string

const (
    JobPending   RunJobStatus = "PENDING"
    JobRunning   RunJobStatus = "RUNNING"
    JobSucceeded RunJobStatus = "SUCCEEDED"
    JobFailed    RunJobStatus = "FAILED"
)

// RunJob holds async job info
type RunJob struct {
    ID        string         `json:"id"`
    ProjectID uint           `json:"project_id"`
    Status    RunJobStatus   `json:"status"`
    Error     string         `json:"error,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    Result    *IngestResponse `json:"result,omitempty"`
}

// RunJobManager manages async jobs in-memory
type RunJobManager struct {
    mu   sync.RWMutex
    jobs map[string]*RunJob
}

func NewRunJobManager() *RunJobManager {
    return &RunJobManager{jobs: map[string]*RunJob{}}
}

func (m *RunJobManager) NewJob(projectID uint) *RunJob {
    m.mu.Lock()
    defer m.mu.Unlock()
    id := uuid.New().String()
    j := &RunJob{ID: id, ProjectID: projectID, Status: JobPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
    m.jobs[id] = j
    return j
}

func (m *RunJobManager) Start(id string) {
    m.mu.Lock()
    if j, ok := m.jobs[id]; ok {
        j.Status = JobRunning
        j.UpdatedAt = time.Now()
    }
    m.mu.Unlock()
}

func (m *RunJobManager) Succeed(id string, res *IngestResponse) {
    m.mu.Lock()
    if j, ok := m.jobs[id]; ok {
        j.Status = JobSucceeded
        j.Result = res
        j.UpdatedAt = time.Now()
    }
    m.mu.Unlock()
}

func (m *RunJobManager) Fail(id, errMsg string) {
    m.mu.Lock()
    if j, ok := m.jobs[id]; ok {
        j.Status = JobFailed
        j.Error = errMsg
        j.UpdatedAt = time.Now()
    }
    m.mu.Unlock()
}

func (m *RunJobManager) Get(id string) (*RunJob, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    j, ok := m.jobs[id]
    return j, ok
}
