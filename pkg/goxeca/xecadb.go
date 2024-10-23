package goxeca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"os"

	"github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"
)

type XecaDB struct {
	client *redis.Client
	ctx    context.Context
}

func NewXecaDB(addr string, password string, db int) *XecaDB {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Check the connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	log.Info("Successfully connected to Redis")

	return &XecaDB{
		client: client,
		ctx:    context.Background(),
	}
}

// StoreJob stores a job in Redis
func (x *XecaDB) StoreJob(job *Job) error {
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	return x.client.Set(x.ctx, "job:"+job.ID, jobJSON, 0).Err()
}

// GetJob retrieves a single job from Redis
func (x *XecaDB) GetJob(jobID string) (*Job, error) {
	jobJSON, err := x.client.Get(x.ctx, "job:"+jobID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	var job Job
	err = json.Unmarshal([]byte(jobJSON), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}
	return &job, nil
}

// FetchAllJobs retrieves all jobs from Redis
func (x *XecaDB) FetchAllJobs() ([]*Job, error) {
	keys, err := x.client.Keys(x.ctx, "job:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job keys: %w", err)
	}

	jobs := make([]*Job, 0, len(keys))
	for _, key := range keys {
		// Skip ActiveRuns and TotalRuns keys
		if strings.HasSuffix(key, ":ActiveRuns") || strings.HasSuffix(key, ":TotalRuns") {
			continue
		}
		job, err := x.GetJob(key[4:]) // Remove "job:" prefix
		if err != nil {
			log.Warn("Failed to fetch job", "key", key, "error", err)
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// DeleteJob removes a job from Redis
func (x *XecaDB) DeleteJob(jobID string) error {
	return x.client.Del(x.ctx, "job:"+jobID).Err()
}

// UpdateJobStatus updates the status of a job in Redis
func (x *XecaDB) UpdateJobStatus(jobID string, status string) error {
	job, err := x.GetJob(jobID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Job doesn't exist, possibly already expired
			return nil
		}
		return err
	}
	job.Status = status
	return x.StoreJob(job)
}

// GetPendingJobs retrieves all pending jobs from Redis
func (x *XecaDB) GetPendingJobs() ([]*Job, error) {
	keys, err := x.client.Keys(x.ctx, "job:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job keys: %w", err)
	}

	var pendingJobs []*Job
	for _, key := range keys {
		// Skip ActiveRuns and TotalRuns keys
		if strings.HasSuffix(key, ":ActiveRuns") || strings.HasSuffix(key, ":TotalRuns") {
			continue
		}
		job, err := x.GetJob(key[4:]) // Remove "job:" prefix
		if err != nil {
			log.Warn("Failed to fetch job", "key", key, "error", err)
			continue
		}
		if job.Status == JobStatusPending && !job.Paused && job.NextRunTime.Before(time.Now()) {
			pendingJobs = append(pendingJobs, job)
		}
	}
	return pendingJobs, nil
}

// Close closes the Redis connection
func (x *XecaDB) Close() error {
	return x.client.Close()
}

// StoreJobWithTTL stores a job in Redis with a specified TTL
// this is temporary and will be removed when we have a proper job storage with postgres
func (x *XecaDB) StoreJobWithTTL(job *Job, ttl time.Duration) error {
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	return x.client.Set(x.ctx, "job:"+job.ID, jobJSON, ttl).Err()
}

// GetJobStatus retrieves the status of a job from Redis
func (x *XecaDB) GetJobStatus(jobID string) (string, error) {
	job, err := x.GetJob(jobID)
	if err != nil {
		return "", err
	}
	return job.Status, nil
}

func (x *XecaDB) GetPendingJobsByParentID(parentID string) ([]*Job, error) {
	keys, err := x.client.Keys(x.ctx, "job:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job keys: %w", err)
	}

	var pendingJobs []*Job
	for _, key := range keys {
		// Skip ActiveRuns and TotalRuns keys
		if strings.HasSuffix(key, ":ActiveRuns") || strings.HasSuffix(key, ":TotalRuns") {
			continue
		}
		job, err := x.GetJob(key[4:]) // Remove "job:" prefix
		if err != nil {
			log.Warn("Failed to fetch job", "key", key, "error", err)
			continue
		}
		// Add logic here to filter by parentID if needed
		pendingJobs = append(pendingJobs, job)
	}
	return pendingJobs, nil
}

func (x *XecaDB) IncrementActiveRuns(jobID string) error {
	return x.client.IncrBy(x.ctx, "job:"+jobID+":ActiveRuns", 1).Err()
}

func (x *XecaDB) DecrementActiveRuns(jobID string) error {
	return x.client.DecrBy(x.ctx, "job:"+jobID+":ActiveRuns", 1).Err()
}

func (x *XecaDB) IncrementTotalRuns(jobID string) error {
	return x.client.IncrBy(x.ctx, "job:"+jobID+":TotalRuns", 1).Err()
}

func (x *XecaDB) GetActiveRuns(jobID string) (int64, error) {
	return x.client.Get(x.ctx, "job:"+jobID+":ActiveRuns").Int64()
}

func (x *XecaDB) GetTotalRuns(jobID string) (int64, error) {
	return x.client.Get(x.ctx, "job:"+jobID+":TotalRuns").Int64()
}
