# GoXeca - Job Scheduler & Executor

GoXeca is a flexible and scalable job scheduling and execution system built in Go. GoXeca is designed to handle complex job execution with a focus on reliability, concurrency, and performance.

## Key Features

- Job Scheduling & Execution: Efficiently schedule and execute tasks with configurable timing and concurrency.

- Time Parsing: Leverages powerful natural language time parsing for flexible job timings.

- Job Chaining & Concurrency: Supports executing jobs in sequence or in parallel to improve workflow efficiency.

- Scalability: Designed to scale across different job loads using Go’s goroutines for lightweight task execution.

## Inspiration

Albrow Jobs – Influenced the job processing and execution design.

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

Getting Started

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

Future Enhancements

[ ] Improve logging and error handling.

[ ] Implement retry logic for failed jobs.

[ ] Add support for persistence using Redis and PostgreSQL.

License

This project is licensed under the MIT License.
