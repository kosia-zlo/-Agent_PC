# Agent Management Server

Utility for managing a fleet of agents on Windows/Linux machines.

## Quick Start

1. Create necessary directories:
   ```bash
   mkdir -p data updates
   ```

2. Start the server:
   ```bash
   docker compose up --build
   ```

3. Open the web interface:
   [http://127.0.0.1:8080](http://127.0.0.1:8080)

## API Examples

### Agent Registration
```bash
curl -X POST http://127.0.0.1:8080/api/agent/register \
     -H "Content-Type: application/json" \
     -d '{"id":"test-1","hostname":"m1","os":"linux","arch":"amd64","version":"v0.1.0"}'
```

### Get Tasks for Agent
```bash
curl "http://127.0.0.1:8080/api/agent/tasks?agent_id=test-1"
```

### Submit Task Result
```bash
curl -X POST http://127.0.0.1:8080/api/agent/tasks/TASK_ID/result \
     -H "Content-Type: application/json" \
     -d '{"status":"done","result":{"found_files":["/etc/passwd"]}}'
```
