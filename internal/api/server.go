package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GauravKarakoti/flowforge/internal/queue"
	"github.com/GauravKarakoti/flowforge/internal/store"
)

// Job represents a workflow task execution record.
type Job struct {
	ID         string                 `json:"id"`
	TaskName   string                 `json:"task_name"`
	Args       []interface{}          `json:"args"`
	Kwargs     map[string]interface{} `json:"kwargs"`
	Queue      string                 `json:"queue"`
	RetryLimit int                    `json:"retry_limit"`
	Status     string                 `json:"status"` // pending, running, completed, failed
	Result     interface{}            `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// WorkerInfo represents a connected worker node.
type WorkerInfo struct {
	ID       string    `json:"id"`
	Queue    string    `json:"queue"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

// Server provides the control plane HTTP API and serves static UI assets.
type Server struct {
	// Store and Broker dependencies
	store  store.Store
	broker queue.Broker

	// In-memory state tracking
	mu      sync.RWMutex
	jobs    map[string]*Job
	workers map[string]*WorkerInfo

	uiDir string
}

// NewServer initializes and returns a new Server instance.
func NewServer(s store.Store, b queue.Broker) *Server {
	return &Server{
		store:   s,
		broker:  b,
		jobs:    make(map[string]*Job),
		workers: make(map[string]*WorkerInfo),
		uiDir:   "ui/dist",
	}
}

// SetUIDir configures the directory path for UI static assets.
func (s *Server) SetUIDir(dir string) {
	s.uiDir = dir
}

// Router configures and returns the HTTP request multiplexer.
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// UI endpoints
	mux.HandleFunc("/api/jobs", s.handleJobs)
	mux.HandleFunc("/api/workers", s.handleWorkers)

	// Client SDK & Worker endpoints
	mux.HandleFunc("/api/jobs/enqueue", s.handleEnqueue)
	mux.HandleFunc("/api/jobs/poll", s.handlePoll)
	mux.HandleFunc("/api/jobs/", s.handleJobStatusOrDetail)

	// Serve React UI static files
	mux.HandleFunc("/", s.handleStaticUI)

	return mux
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start(addr string) error {
	router := s.Router()
	log.Printf("FlowForge Control Plane API listening on %s", addr)
	return http.ListenAndServe(addr, router)
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	jobList := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobList = append(jobList, j)
	}
	s.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"jobs": jobList,
	})
}

func (s *Server) handleWorkers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	workerList := make([]*WorkerInfo, 0, len(s.workers))
	for _, worker := range s.workers {
		workerList = append(workerList, worker)
	}
	s.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"workers": workerList,
	})
}

type EnqueueRequest struct {
	TaskName   string                 `json:"task_name"`
	Args       []interface{}          `json:"args"`
	Kwargs     map[string]interface{} `json:"kwargs"`
	Queue      string                 `json:"queue"`
	RetryLimit int                    `json:"retry_limit"`
}

func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.TaskName == "" {
		http.Error(w, "task_name is required", http.StatusBadRequest)
		return
	}

	if req.Queue == "" {
		req.Queue = "default"
	}
	if req.RetryLimit <= 0 {
		req.RetryLimit = 3
	}
	if req.Args == nil {
		req.Args = make([]interface{}, 0)
	}
	if req.Kwargs == nil {
		req.Kwargs = make(map[string]interface{})
	}

	jobID := generateID()
	now := time.Now().UTC()

	job := &Job{
		ID:         jobID,
		TaskName:   req.TaskName,
		Args:       req.Args,
		Kwargs:     req.Kwargs,
		Queue:      req.Queue,
		RetryLimit: req.RetryLimit,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	ctx := r.Context()
	if s.store != nil {
		_ = s.store.SaveJob(ctx, job)
	}

	if s.broker != nil {
		payload, err := json.Marshal(job)
		if err == nil {
			_ = s.broker.Enqueue(ctx, req.Queue, payload)
		}
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "enqueued",
		"job_id": jobID,
		"job":    job,
	})
}

func (s *Server) handlePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueName := r.URL.Query().Get("queue")
	if queueName == "" {
		queueName = "default"
	}

	workerID := r.URL.Query().Get("worker_id")
	if workerID == "" {
		workerID = "worker-" + queueName
	}

	s.mu.Lock()
	s.workers[workerID] = &WorkerInfo{
		ID:       workerID,
		Queue:    queueName,
		Status:   "active",
		LastSeen: time.Now().UTC(),
	}
	s.mu.Unlock()

	// Try broker dequeue first if available
	if s.broker != nil {
		payload, err := s.broker.Dequeue(r.Context(), []string{queueName})
		if err == nil && len(payload) > 0 {
			var job Job
			if err := json.Unmarshal(payload, &job); err == nil {
				s.mu.Lock()
				if existing, exists := s.jobs[job.ID]; exists {
					existing.Status = "running"
					existing.UpdatedAt = time.Now().UTC()
					job = *existing
				} else {
					job.Status = "running"
					job.UpdatedAt = time.Now().UTC()
					s.jobs[job.ID] = &job
				}
				s.mu.Unlock()
				writeJSON(w, http.StatusOK, map[string]interface{}{"job": job})
				return
			}
		}
	}

	// In-memory queue fallback
	s.mu.Lock()
	var selectedJob *Job
	for _, j := range s.jobs {
		if j.Queue == queueName && j.Status == "pending" {
			j.Status = "running"
			j.UpdatedAt = time.Now().UTC()
			selectedJob = j
			break
		}
	}
	s.mu.Unlock()

	if selectedJob != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"job": selectedJob})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleJobStatusOrDetail(w http.ResponseWriter, r *http.Request) {
	// Pattern: /api/jobs/{id}/status or /api/jobs/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	jobID := parts[0]

	if len(parts) == 2 && parts[1] == "status" {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var update struct {
			Status string      `json:"status"`
			Result interface{} `json:"result,omitempty"`
			Error  string      `json:"error,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		s.mu.Lock()
		job, exists := s.jobs[jobID]
		if !exists {
			// If not in-memory, create placeholder record
			job = &Job{
				ID:        jobID,
				Status:    update.Status,
				Result:    update.Result,
				Error:     update.Error,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			s.jobs[jobID] = job
		} else {
			job.Status = update.Status
			if update.Result != nil {
				job.Result = update.Result
			}
			if update.Error != "" {
				job.Error = update.Error
			}
			job.UpdatedAt = time.Now().UTC()
		}
		s.mu.Unlock()

		if s.store != nil {
			_ = s.store.SaveJob(r.Context(), job)
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "updated",
			"job":    job,
		})
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.mu.RLock()
		job, exists := s.jobs[jobID]
		s.mu.RUnlock()

		if !exists && s.store != nil {
			if stored, err := s.store.GetJob(r.Context(), jobID); err == nil && stored != nil {
				writeJSON(w, http.StatusOK, stored)
				return
			}
		}

		if !exists {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}

		writeJSON(w, http.StatusOK, job)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) handleStaticUI(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cleaned := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	targetPath := filepath.Join(s.uiDir, cleaned)

	fileInfo, err := os.Stat(targetPath)
	if err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, targetPath)
		return
	}

	// Fallback to index.html for Single Page Applications (SPA)
	indexPath := filepath.Join(s.uiDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	// Built-in lightweight fallback dashboard if UI dist is not yet built
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>FlowForge Control Plane</title></head>
<body>
  <h1>FlowForge Control Plane</h1>
  <p>Status: Healthy</p>
  <ul>
    <li><a href="/api/jobs">Jobs API</a></li>
    <li><a href="/api/workers">Workers API</a></li>
  </ul>
</body>
</html>`))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("job-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
