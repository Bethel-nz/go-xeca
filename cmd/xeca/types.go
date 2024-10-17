package xeca

import "time"

type APIJob struct {
	ID          string    `json:"id"`
	Command     string    `json:"command"`
	Schedule    string    `json:"schedule"`
	Status      string    `json:"status"`
	NextRunTime time.Time `json:"nextRunTime"`
	Output      string    `json:"output"`
}

type APIJobRequest struct {
	Command      string        `json:"command"`
	Schedule     string        `json:"schedule"`
	IsRecurring  bool          `json:"isRecurring"`
	Priority     int           `json:"priority"`
	Dependencies []string      `json:"dependencies"`
	MaxRetries   int           `json:"maxRetries"`
	RetryDelay   time.Duration `json:"retryDelay"`
	Webhook      string        `json:"webhook"`
	Timeout      time.Duration `json:"timeout"`
	ChainedJobs  []string      `json:"chainedJobs"`
}
