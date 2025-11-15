// Package scheduler provides job scheduling functionality for Discord bot actions.
package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/robfig/cron/v3"
)

// JobFunc represents a scheduled job function
type JobFunc func(ctx context.Context) error

// JobInfo contains information about a scheduled job
type JobInfo struct {
	ID       string
	Name     string
	Schedule string
}

// Scheduler manages scheduled jobs
type Scheduler struct {
	cron    *cron.Cron
	logger  logging.Logger
	jobs    map[string]*jobEntry
	jobsMu  sync.RWMutex
	running bool
	runMu   sync.RWMutex
}

type jobEntry struct {
	id       cron.EntryID
	name     string
	schedule string
	fn       JobFunc
}

// New creates a new scheduler
func New(logger logging.Logger) *Scheduler {
	logger.Info("Creating new scheduler")

	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		logger:  logger,
		jobs:    make(map[string]*jobEntry),
		running: false,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.logger.Info("Starting scheduler")
	s.cron.Start()
	s.running = true

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler not running")
	}

	s.logger.Info("Stopping scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.running = false

	return nil
}

// IsRunning returns whether the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.runMu.RLock()
	defer s.runMu.RUnlock()
	return s.running
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(name, schedule string, fn JobFunc) (string, error) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	s.logger.Debug("Adding job", "name", name, "schedule", schedule)

	// Wrap the job function to handle context and errors
	wrappedFn := func() {
		ctx := context.Background()
		if err := fn(ctx); err != nil {
			s.logger.Error("Job execution failed", "name", name, "error", err)
		}
	}

	// Add job to cron
	entryID, err := s.cron.AddFunc(schedule, wrappedFn)
	if err != nil {
		s.logger.Error("Failed to add job", "name", name, "error", err)
		return "", fmt.Errorf("invalid cron expression: %w", err)
	}

	// Generate job ID
	jobID := fmt.Sprintf("job-%d", entryID)

	// Store job entry
	s.jobs[jobID] = &jobEntry{
		id:       entryID,
		name:     name,
		schedule: schedule,
		fn:       fn,
	}

	s.logger.Debug("Job added successfully", "jobID", jobID, "name", name)

	return jobID, nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID string) error {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	s.logger.Debug("Removing job", "jobID", jobID)

	job, exists := s.jobs[jobID]
	if !exists {
		s.logger.Warn("Job not found", "jobID", jobID)
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Remove from cron
	s.cron.Remove(job.id)

	// Remove from jobs map
	delete(s.jobs, jobID)

	s.logger.Debug("Job removed successfully", "jobID", jobID)

	return nil
}

// GetJobInfo returns information about a job
func (s *Scheduler) GetJobInfo(jobID string) (*JobInfo, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return &JobInfo{
		ID:       jobID,
		Name:     job.name,
		Schedule: job.schedule,
	}, nil
}

// ListJobs returns a list of all scheduled jobs
func (s *Scheduler) ListJobs() []JobInfo {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	jobs := make([]JobInfo, 0, len(s.jobs))
	for jobID, job := range s.jobs {
		jobs = append(jobs, JobInfo{
			ID:       jobID,
			Name:     job.name,
			Schedule: job.schedule,
		})
	}

	return jobs
}

// LoadFromConfig loads scheduled actions from configuration
func (s *Scheduler) LoadFromConfig(cfg *config.Config) (int, error) {
	s.logger.Info("Loading scheduled actions from config")

	count := 0
	for _, action := range cfg.Actions {
		if action.Type == "scheduled" && action.Trigger.Schedule != "" {
			s.logger.Debug("Found scheduled action", "name", action.Name, "schedule", action.Trigger.Schedule)
			count++
		}
	}

	s.logger.Info("Scheduled actions loaded", "count", count)
	return count, nil
}
