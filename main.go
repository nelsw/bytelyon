package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
)

func main() {
	// 1. Define the command (e.g., pinging a website 4 times)
	cmd := exec.Command("./main.py", "{\"id\":9999,\"type\":\"news\",\"query\":\"btc forecast\",\"blacklist\":[\"foo\",\"bar\",\"baz\"],\"headless\":false,\"last_run_at\":\"2025-08-18T18:28:24.635990Z\"}")

	// 2. Create a pipe to read standard output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// 3. Start the process without waiting for it to finish
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// 4. Use bufio.Scanner to read the stream line-by-line in real time
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("[STREAMED]: %s\n", line)
	}

	// Check for any scanning errors
	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}

	// 5. Always wait for the command to finish after cleaning up the pipe
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Command finished with error: %v", err)
	}
}
