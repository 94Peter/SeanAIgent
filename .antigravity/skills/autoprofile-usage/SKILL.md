---
name: AutoProfile Usage
description: Guide on how to execute AutoProfile to generate performance reports (dump-only) for AI analysis.
---

# AutoProfile Usage Skill

AutoProfile is an AI-driven performance diagnostic tool for Go services. It orchestrates k6 load testing, captures pprof profiles (CPU, Heap, Goroutine), extracts Go source context, and packages them into a `prompt` for AI agents to analyze and optimize.

## When to use this skill
Use this skill when you need to run performance diagnostics on a local Go service to identify CPU bottlenecks, memory leaks, or goroutine leaks. 

## Prerequisites
1. The target Go project must be built or running (e.g., `go run main.go`).
2. The target Go service must have `net/http/pprof` registered and accessible (usually at `http://localhost:<PORT>/debug/pprof/`).

## How to execute AutoProfile (Dump Only Mode)

To generate a performance report without hitting the Gemini LLM API directly (allowing YOU, the AI agent, to read the dump file instead), you have two options: using the native Go binary or executing via Docker.

### Option 1: Native Binary (Preferred if available)
If the `autoprofile` binary is built and available in the environment path, run the following:

```bash
autoprofile run \
  --localhost \
  --target-host localhost \
  --port <TARGET_PORT> \
  --dump-only \
  --k6-path /your/endpoint \
  --k6-method POST \
  --k6-body '{"key": "value"}'
```

### Option 2: Docker Execution (Fallback)
If the native binary is not available, execute the following Docker command exactly as shown. Note that it will automatically pull the `94peter/autoprofile:latest` image from Docker Hub.

**Important Notes for Execution:**
- You must run this command in the root directory of the TARGET Go project that you want to profile.
- Replace `<TARGET_PORT>` with the port your Go service is listening on.
- Adjust the `--k6-path`, `--k6-method`, and `--k6-body` as necessary to hit the specific endpoint you want to test.

```bash
docker run --rm \
  -v $(pwd):/app \
  94peter/autoprofile:latest run \
  --localhost \
  --target-host host.docker.internal \
  --port <TARGET_PORT> \
  --dump-only \
  --k6-path /your/endpoint \
  --k6-method POST \
  --k6-body '{"key": "value"}'
```

## After Execution

1. AutoProfile will output an `autoprofile_prompt.txt` file in the current directory.
2. Read this file. It contains a "Golden Diagnostic Report" detailing the `k6` summary, the `pprof` top allocators/CPU users, and the extracted Go source code snippets.
3. Use the information in this report to autonomously diagnose issues and refactor the Go source code to optimize performance.
