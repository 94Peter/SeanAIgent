---
description: Execute AutoProfile performance diagnostics and automatically resolve Go code bottlenecks
---

1. Ensure your current directory is within the target Go project you wish to analyze.
// turbo-all
2. Run the target Go service in the background (e.g., `go run main.go`).
3. Wait for the service to start, then execute the load test against the target endpoint in `dump-only` mode using the instructions provided in the `AutoProfile Usage` Skill (using either the native `autoprofile run` binary or the `94peter/autoprofile:latest` Docker image). This will generate an `autoprofile_prompt.txt` file.
4. Read and analyze the generated `autoprofile_prompt.txt`. Pay close attention to the `hot_cpu_top`, `cold_heap_top`, and `cold_goroutine_top` results, correlating them with the extracted source code context.
5. Based on the analysis, refactor the source code to resolve the identified bottlenecks (e.g., eliminate CPU bottlenecks, fix memory leaks, or resolve abandoned goroutines).
6. Implement Table-Driven Tests to ensure the business logic remains unbroken, and verify that `golangci-lint run` passes without errors.
7. Restart the service and repeat Step 3 to verify that the performance bottlenecks have been successfully eliminated. Repeat this cycle until the optimization goals are met.
