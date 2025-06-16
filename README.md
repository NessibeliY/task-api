# task-api

A simple HTTP service for handling long-running asynchronous I/O-bound tasks (e.g. 3–5 minutes) entirely in memory, without any external databases, queues, or services.

##  Features

- Create tasks
- Fetch task by ID
- Fetch all tasks
- Delete task
- Track task status, creation time, and processing duration

##  Getting Started

### Requirements

- Go 1.22 or higher

### Run the Server

```bash
go run .
```

By default, the server listens on port 8080.

⸻

📬 HTTP API

1. Create a Task
```bash
curl --location 'http://localhost:8080/api/v1/tasks' \
--header 'Content-Type: application/json' \
--data '{
"title": "Test Task",
"description": "This is a test task"
}'
```

⸻

2. Get Task by ID
```bash
curl --location 'http://localhost:8080/api/v1/tasks/2'
```

⸻

3. Get All Tasks
```bash
curl --location 'http://localhost:8080/api/v1/tasks'
```

⸻

4. Delete a Task
```bash
curl --location --request DELETE 'http://localhost:8080/api/v1/tasks/1'
```

⸻


 Notes for Developers
	•	All data is stored in memory — restarting the service clears all tasks.
	•	Tasks are processed asynchronously using goroutines.
	•	Task processing duration is simulated and can be configured for real workloads later.
	•	The codebase is clean and extensible: ideal for adding more task types, metrics, persistence, etc.

⸻

 Run Tests
```bash
go test ./...
```
