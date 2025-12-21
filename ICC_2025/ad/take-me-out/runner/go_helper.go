package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	helperLogClip        = 4096
	helperTimeoutGraceMs = 3000
)

type helperTest struct {
	Input    string `json:"input"`
	Output   string `json:"output"`
	IsSample bool   `json:"is_sample"`
}

type helperPayload struct {
	Mode  string       `json:"mode"`
	Tests []helperTest `json:"tests"`
}

func clipForLog(s string) string {
	if len(s) <= helperLogClip {
		return s
	}
	return s[:helperLogClip] + fmt.Sprintf("... (truncated %d bytes)", len(s)-helperLogClip)
}

func helperExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

func executeGoViaHelper(req RunRequest) RunResponse {
	globalLimitMs := 30000
	if v := os.Getenv("RUNNER_GLOBAL_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			globalLimitMs = n
		}
	}

	outLimit := 65536
	if v := os.Getenv("RUN_LIMIT_OUTPUT_BYTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			outLimit = n
		}
	}

	helperPath := strings.TrimSpace(os.Getenv("GO_HELPER_PATH"))
	if helperPath == "" {
		helperPath = "/usr/local/bin/go-helper"
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	challengeName := strings.TrimSpace(req.Challenge)
	singleMode := challengeName == ""
	if singleMode {
		mode = "single"
	} else {
		if mode == "" {
			mode = "judge"
		}
	}
	revealExpected := mode == "sample" || mode == "single"
	sanitize := func(res RunResponse) RunResponse {
		if !revealExpected {
			res.Output = ""
			res.Expected = ""
		}
		return sanitizeRunResponse(req, res)
	}

	var tests []runnerTest
	if !singleMode {
		tests = getRunnerTests(challengeName, req.Mode)
		if len(tests) == 0 {
			return sanitize(RunResponse{Result: "Unknown challenge"})
		}
	} else {
		tests = []runnerTest{{
			Input:    req.Input,
			Output:   req.Want,
			IsSample: true,
		}}
	}

	jobDir, err := os.MkdirTemp("", "gohelper-")
	if err != nil {
		log.Printf("go helper client: mkdtemp failed: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	defer os.RemoveAll(jobDir)

	codePath := filepath.Join(jobDir, "code.go")
	if err := os.WriteFile(codePath, []byte(req.Code), 0o600); err != nil {
		log.Printf("go helper client: writing code failed: %v", err)
		return sanitize(RunResponse{Result: "Internal Error"})
	}

	hTests := make([]helperTest, len(tests))
	for i, tc := range tests {
		hTests[i] = helperTest{Input: tc.Input, Output: tc.Output, IsSample: tc.IsSample}
	}
	payload := helperPayload{Mode: mode, Tests: hTests}
	testsPath := filepath.Join(jobDir, "tests.json")
	if data, err := json.Marshal(payload); err != nil {
		log.Printf("go helper client: marshal tests failed: %v", err)
		return sanitize(RunResponse{Result: "Internal Error"})
	} else if err := os.WriteFile(testsPath, data, 0o600); err != nil {
		log.Printf("go helper client: write tests failed: %v", err)
		return sanitize(RunResponse{Result: "Internal Error"})
	}

	sandboxEnv := strings.TrimSpace(os.Getenv("SANDBOX_ENVS_DIR"))

	args := []string{
		"--mode", mode,
		"--global-timeout", strconv.Itoa(globalLimitMs),
		"--output-limit", strconv.Itoa(outLimit),
		"--code-file", codePath,
		"--tests-file", testsPath,
	}
	if sandboxEnv != "" {
		args = append(args, "--sandbox-env", sandboxEnv)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalLimitMs+helperTimeoutGraceMs)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, helperPath, args...)
	cmd.Env = os.Environ()
	var combined strings.Builder
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	err = cmd.Run()
	exitCode := helperExitCode(err)
	rawOutput := combined.String()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("go helper client: helper timed out waiting for completion")
			return sanitize(RunResponse{Result: "Internal Error"})
		}
		log.Printf("go helper client: command failed err=%v exit=%d output=%s", err, exitCode, clipForLog(rawOutput))
		if combined.Len() == 0 {
			return sanitize(RunResponse{Result: "Internal Error"})
		}
	}

	if combined.Len() == 0 {
		return sanitize(RunResponse{Result: "Internal Error"})
	}

	var resp RunResponse
	resp, err = parseHelperResponse(rawOutput)
	if err != nil {
		log.Printf("go helper client: failed parsing helper output: %v payload=%s", err, clipForLog(rawOutput))
		return sanitize(RunResponse{Result: "Internal Error"})
	}
	if resp.Result == "Compile Error" || (resp.Result != "Wrong Answer" && resp.Result != "Runtime Error" && resp.Result != "Time Limit Exceeded") {
		resp.FailedIndex = -1
	}
	return sanitize(resp)
}

func parseHelperResponse(raw string) (RunResponse, error) {
	var resp RunResponse
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return resp, errors.New("empty helper output")
	}
	if err := json.Unmarshal([]byte(trimmed), &resp); err == nil {
		return resp, nil
	}
	if idx := strings.LastIndex(trimmed, "{"); idx >= 0 {
		candidate := trimmed[idx:]
		if err := json.Unmarshal([]byte(candidate), &resp); err == nil {
			return resp, nil
		}
	}
	lines := strings.Split(trimmed, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(strings.Join(lines[i:], "\n"))
		if candidate == "" {
			continue
		}
		if err := json.Unmarshal([]byte(candidate), &resp); err == nil {
			return resp, nil
		}
	}
	return resp, errors.New("no JSON payload detected")
}
