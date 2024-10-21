# GoXeca - Job Scheduler & Executor

GoXeca is a flexible and scalable job scheduling and execution system built in Go. GoXeca is designed to handle complex job execution with a focus on reliability, concurrency, and performance.

## Key Features

- Job Scheduling & Execution: Efficiently schedule and execute tasks with configurable timing and concurrency.

- Time Parsing: Leverages powerful natural language time parsing for flexible job timings.

- Job Chaining & Concurrency: Supports executing jobs in sequence or in parallel to improve workflow efficiency.

- Scalability: Designed to scale across different job loads using Go’s goroutines for lightweight task execution.

## Inspiration

Albrow Jobs – Influenced the job processing and execution design. [here](https://github.com/albrow/jobs)

## Use Case

GoXeca is ideal for tasks like:

- Deployment Checks: Automate system and API health checks.

- Task Orchestration: Efficiently run scheduled or ad-hoc background tasks.

- System Monitoring: Periodically run status checks or other monitoring jobs.

Roadmap

Here's what’s planned:

[x] Fully optimize goroutine usage for improved concurrency management.

[ ] Refine job execution time handling for better accuracy.

[ ] Implement advanced job chaining and dependency management.

[ ] Develop a React-based frontend to visualize and interact with scheduled jobs.



## Getting Started

Clone the repo:

```bash

git clone https://github.com/Bethel-nz/goxeca.git
cd goxeca

```

Install Dependencies:

```bash

go mod tidy

```

Run GoXeca:

```bash

go run main.go

```

### Example with api

// Job that runs once in 2 minutes

```bash

curl -X POST http://localhost:8080/api/add-job \
     -H "Content-Type: application/json" \
     -d '{
         "command": "pinger ping",
         "schedule": "in 2 minutes",
         "isRecurring": false,
         "priority": 10,
         "dependencies": [],
         "maxRetries": 3,
         "retryDelay": 5000000000,
         "webhook": "http://localhost:8080/api/webhook",
         "timeout": 30000000000
     }'
```

// Job that runs every 5 seconds

```bash
curl -X POST http://localhost:8080/api/add-job \
     -H "Content-Type: application/json" \
     -d '{
         "command": "pinger ping",
         "schedule": "every 5 seconds",
         "isRecurring": true,
         "priority": 5,
         "dependencies": [],
         "maxRetries": 3,
         "retryDelay": 1000000000,
         "webhook": "http://localhost:8080/api/webhook",
         "timeout": 10000000000
     }'
```

// After running these commands, you can check the status of jobs:

```bash
curl http://localhost:8080/api/jobs
```

// To stop a recurring job, use (replace JOB_ID with the actual job ID):

```bash
 curl -X POST http://localhost:8080/api/stop-job/JOB_ID

```

Future Enhancements

[ ] Improve logging and error handling.

[x] Implement retry logic for failed jobs.

[x] Add support for persistence using Redis and PostgreSQL. - [x] redis - [] postgreSQL

License

This project is licensed under the MIT License.
