package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/geekxflood/gxf-discord-bot/internal/testutil"
	"github.com/geekxflood/gxf-discord-bot/pkg/config"
	"github.com/geekxflood/gxf-discord-bot/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewScheduler_Success(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)

	require.NotNil(t, sched)
	assert.False(t, sched.IsRunning())
}

func TestScheduler_StartStop(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	require.NotNil(t, sched)

	// Start scheduler
	err := sched.Start()
	require.NoError(t, err)
	assert.True(t, sched.IsRunning())

	// Stop scheduler
	err = sched.Stop()
	require.NoError(t, err)
	assert.False(t, sched.IsRunning())
}

func TestScheduler_StartTwice(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)

	err := sched.Start()
	require.NoError(t, err)

	// Starting again should return error
	err = sched.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	sched.Stop()
}

func TestScheduler_StopWithoutStart(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)

	// Stopping without starting should return error
	err := sched.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestScheduler_AddJob(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	err := sched.Start()
	require.NoError(t, err)
	defer sched.Stop()

	// Add a simple job
	executed := false
	job := func(ctx context.Context) error {
		executed = true
		return nil
	}

	// Use @hourly descriptor instead of cron expression
	jobID, err := sched.AddJob("test-job", "@hourly", job)
	require.NoError(t, err)
	assert.NotEmpty(t, jobID)
	assert.False(t, executed) // Job shouldn't execute immediately
}

func TestScheduler_AddJobInvalidCron(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	err := sched.Start()
	require.NoError(t, err)
	defer sched.Stop()

	job := func(ctx context.Context) error {
		return nil
	}

	// Invalid cron expression
	_, err = sched.AddJob("test-job", "invalid", job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cron expression")
}

func TestScheduler_RemoveJob(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	err := sched.Start()
	require.NoError(t, err)
	defer sched.Stop()

	job := func(ctx context.Context) error {
		return nil
	}

	jobID, err := sched.AddJob("test-job", "@daily", job)
	require.NoError(t, err)

	// Remove the job
	err = sched.RemoveJob(jobID)
	assert.NoError(t, err)
}

func TestScheduler_RemoveNonExistentJob(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	err := sched.Start()
	require.NoError(t, err)
	defer sched.Stop()

	err = sched.RemoveJob("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

func TestScheduler_JobExecution(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	err := sched.Start()
	require.NoError(t, err)
	defer sched.Stop()

	var executedMu sync.Mutex
	executed := false
	job := func(ctx context.Context) error {
		executedMu.Lock()
		executed = true
		executedMu.Unlock()
		return nil
	}

	// Schedule job to run every second
	_, err = sched.AddJob("test-job", "@every 1s", job)
	require.NoError(t, err)

	// Wait for job to execute
	time.Sleep(1500 * time.Millisecond)

	executedMu.Lock()
	wasExecuted := executed
	executedMu.Unlock()
	assert.True(t, wasExecuted, "Job should have been executed")
}

func TestScheduler_GetJobInfo(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	err := sched.Start()
	require.NoError(t, err)
	defer sched.Stop()

	job := func(ctx context.Context) error {
		return nil
	}

	jobID, err := sched.AddJob("test-job", "@weekly", job)
	require.NoError(t, err)

	info, err := sched.GetJobInfo(jobID)
	require.NoError(t, err)
	assert.Equal(t, "test-job", info.Name)
	assert.Equal(t, "@weekly", info.Schedule)
	assert.Equal(t, jobID, info.ID)
}

func TestScheduler_ListJobs(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	sched := scheduler.New(logger)
	err := sched.Start()
	require.NoError(t, err)
	defer sched.Stop()

	job := func(ctx context.Context) error {
		return nil
	}

	// Add multiple jobs
	_, err = sched.AddJob("job1", "@daily", job)
	require.NoError(t, err)
	_, err = sched.AddJob("job2", "@hourly", job)
	require.NoError(t, err)

	jobs := sched.ListJobs()
	assert.Len(t, jobs, 2)
}

func TestScheduler_LoadFromConfig(t *testing.T) {
	logger := &testutil.MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	cfg := &config.Config{
		Actions: []config.ActionConfig{
			{
				Name: "scheduled-task",
				Type: "scheduled",
				Trigger: config.TriggerConfig{
					Schedule: "@daily",
				},
				Response: config.ResponseConfig{
					Type:    "text",
					Content: "Daily reminder!",
				},
			},
		},
	}

	sched := scheduler.New(logger)

	// Should be able to load scheduled actions from config
	count, err := sched.LoadFromConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
