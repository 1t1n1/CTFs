package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sandbox "goexe-runner/internal/sandbox"
)

// RunRequest defines the JSON request for code execution
type RunRequest struct {
	Language  string `json:"language"`
	Code      string `json:"code"`
	Input     string `json:"input"`
	Want      string `json:"want"`
	Challenge string `json:"challenge,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Sandbox   string `json:"sandbox,omitempty"`
}

// RunResponse defines the JSON response
type RunResponse struct {
	Result      string `json:"result"`
	Output      string `json:"output,omitempty"`
	DurationMs  int    `json:"duration_ms,omitempty"`
	FailedIndex int    `json:"failed_index,omitempty"`
	Expected    string `json:"expected,omitempty"`
}

func sanitizeRunResponse(req RunRequest, resp RunResponse) RunResponse {
	if strings.TrimSpace(req.Challenge) == "" {
		return resp
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != "sample" {
		resp.Output = ""
		resp.Expected = ""
	}
	return resp
}

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type job struct {
	req  RunRequest
	resp chan RunResponse
	ctx  context.Context
}

var jobQueue chan job

func initWorkerPool() {
	if jobQueue != nil {
		return
	}
	workerCount := envInt("RUNNER_WORKERS", 4)
	if workerCount <= 0 {
		workerCount = 1
	}
	queueSize := envInt("RUNNER_QUEUE_SIZE", workerCount*4)
	if queueSize <= 0 {
		queueSize = workerCount * 4
	}
	jobQueue = make(chan job, queueSize)
	for i := 0; i < workerCount; i++ {
		go worker(jobQueue)
	}
	log.Printf("Runner worker pool started (workers=%d, queue=%d)", workerCount, queueSize)
}

func worker(queue <-chan job) {
	for j := range queue {
		select {
		case <-j.ctx.Done():
			continue
		default:
		}
		resp := execute(j.req)
		select {
		case j.resp <- resp:
		case <-j.ctx.Done():
		}
	}
}

func envInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// runHandler handles POST /run requests
func runHandler(w http.ResponseWriter, r *http.Request) {
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if jobQueue == nil {
		http.Error(w, "Runner not ready", http.StatusServiceUnavailable)
		return
	}
	ctx := r.Context()
	respCh := make(chan RunResponse, 1)
	j := job{req: req, resp: respCh, ctx: ctx}

	select {
	case jobQueue <- j:
	case <-ctx.Done():
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	}

	select {
	case resp := <-respCh:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case <-ctx.Done():
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
	}
}

// challengeMetaHandler returns description and sample I/O for a challenge
func challengeMetaHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	meta, ok := getChallengeMeta(name)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

// execute compiles (if needed) and runs code inside an isolated chroot sandbox
func execute(req RunRequest) RunResponse {
	sandboxMode := strings.TrimSpace(req.Sandbox)
	if sandboxMode == "" {
		sandboxMode = "default"
	}
	if sandboxMode != "default" && sandboxMode != "nsjail_only" {
		return RunResponse{Result: "Unsupported sandbox mode"}
	}
	if req.Language == "go" {
		if sandboxMode != "default" {
			return RunResponse{Result: "Unsupported sandbox mode"}
		}
		return executeGoViaHelper(req)
	}
	defaultUseChrootRunner := sandboxMode != "nsjail_only"
	if req.Language == "c" {
		return executeCTwoStage(req, false)
	}
	useChrootRunner := defaultUseChrootRunner

	// create a per-run workdir inside shared base rootfs, then chroot
	rr, err := sandbox.PrepareRunRoot(req.Language)
	if err != nil {
		log.Printf("Failed to prepare sandbox: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	defer rr.Cleanup()

	// determine file extension and write code under /work in chroot
	ext := map[string]string{"c": ".c", "python": ".py", "ruby": ".rb"}[req.Language]
	if ext == "" {
		return RunResponse{Result: "Unsupported language: " + req.Language}
	}
	// host path to the work dir (inside chroot this is /work)
	hostWork := rr.WorkspaceHost
	workdir := rr.WorkspaceDir()
	workspaceRel := rr.WorkspaceRel()
	src := filepath.Join(hostWork, "code"+ext)
	if err := ioutil.WriteFile(src, []byte(req.Code), 0644); err != nil {
		log.Printf("Failed to write code: %v", err)
		return RunResponse{Result: "Internal Error"}
	}

	captureDir := filepath.Join(hostWork, ".runner")
	if err := resetDir(captureDir, 0o755); err != nil {
		log.Printf("Failed to prepare capture dir: %v", err)
		return RunResponse{Result: "Internal Error"}
	}

	// For Go: prepare a separate compile workspace under base rootfs (/tmp/comp-*)
	// so that we can use the base's /dev and toolchain even if runroot lacks /dev nodes
	// timeouts: global (compile+run) and exec-only
	execLimit := defaultTimeLimitMs
	globalLimitMs := 5000
	if v := os.Getenv("RUNNER_GLOBAL_TIMEOUT_MS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			globalLimitMs = n
		}
	}
	globalCtx, globalCancel := context.WithTimeout(context.Background(), time.Duration(globalLimitMs)*time.Millisecond)
	defer globalCancel()

	// resource limits
	toBytes := func(mb int) int { return mb * 1024 * 1024 }
	outLimit := 65536
	if v := os.Getenv("RUN_LIMIT_OUTPUT_BYTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			outLimit = n
		}
	}
	comp := sandbox.RLimits{
		CPUSeconds: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_CPU_SEC")); e == nil && n > 0 {
				return n
			}
			return 15
		}(),
		ASBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_AS_MB")); e == nil && n > 0 {
				return n
			}
			return 512
		}()),
		FSizeBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_FSIZE_MB")); e == nil && n > 0 {
				return n
			}
			return 64
		}()),
		NProc: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_NPROC")); e == nil && n > 0 {
				return n
			}
			return 128
		}(),
		NOFile: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_NOFILE")); e == nil && n > 0 {
				return n
			}
			return 512
		}(),
		OutputLimit: outLimit,
	}
	// run limits derived from exec limit
	execCpu := (execLimit+999)/1000 + 1
	// clamp to global
	if execCpu > (globalLimitMs / 1000) {
		execCpu = (globalLimitMs / 1000)
	}
	runLim := sandbox.RLimits{
		CPUSeconds: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_CPU_SEC")); e == nil && n > 0 {
				return n
			}
			return execCpu
		}(),
		ASBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_AS_MB")); e == nil && n > 0 {
				return n
			}
			return 256
		}()),
		FSizeBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_FSIZE_MB")); e == nil && n > 0 {
				return n
			}
			return 16
		}()),
		NProc: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_NPROC")); e == nil && n > 0 {
				return n
			}
			return 64
		}(),
		NOFile: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_NOFILE")); e == nil && n > 0 {
				return n
			}
			return 128
		}(),
		OutputLimit: outLimit,
	}
	insidePath := func(path string) string {
		if strings.HasPrefix(path, hostWork) {
			rel := strings.TrimPrefix(path, hostWork)
			rel = strings.TrimPrefix(rel, "/")
			base := workspaceRel
			if base == "/" {
				if rel == "" {
					return "/"
				}
				return "/" + rel
			}
			if rel == "" {
				return base
			}
			return filepath.Join(base, rel)
		}
		return path
	}

	mapToolPath := func(p string) string {
		if useChrootRunner {
			return p
		}
		if strings.HasPrefix(p, "/") {
			return filepath.Join("/env", strings.TrimPrefix(p, "/"))
		}
		return p
	}
	shellPath := "/bin/sh"
	if !useChrootRunner {
		shellPath = "/env/bin/sh"
	}

	// compile inside chroot if needed (non-root; toolchains write only to work/tmp)
	switch req.Language {
	case "c":
		gccPath := mapToolPath("/usr/bin/gcc")
		compileStdoutHost, compileStderrHost, compileStdoutInside, compileStderrInside := capturePaths(hostWork, workdir, "compile")
		removeFiles(compileStdoutHost, compileStderrHost)
		compileArgs := []string{
			gccPath,
			filepath.Join(workdir, "code.c"),
			"-O2", "-pipe", "-static", "-s", "-lm",
			"-o", filepath.Join(workdir, "code"),
		}
		compileCmd := buildCaptureCommand(compileArgs, compileStdoutInside, compileStderrInside)
		compileRes, err := sandbox.RunInChroot(globalCtx, rr, workdir, []string{shellPath, "-c", compileCmd}, "", comp, useChrootRunner)
		compileStdout, stdoutErr := readFileLimited(compileStdoutHost, outLimit)
		if stdoutErr != nil {
			log.Printf("Failed to read compile stdout: %v", stdoutErr)
		}
		compileStderr, stderrErr := readFileLimited(compileStderrHost, outLimit)
		if stderrErr != nil {
			log.Printf("Failed to read compile stderr: %v", stderrErr)
		}
		summary := combineOutput(compileStdout, compileStderr)
		if summary == "" {
			summary = combineOutput(compileRes.Stdout, compileRes.Stderr)
		}
		if err != nil {
			if summary == "" {
				summary = err.Error()
			}
			return RunResponse{Result: "Compile Error", Output: summary}
		}
		removeFiles(compileStdoutHost, compileStderrHost)
		// Ensure executable for nobody
		_ = os.Chmod(filepath.Join(hostWork, "code"), 0755)
		// Grant chroot capability to compiled binary (setcap outside chroot)
		if out2, err2 := sandbox.RunOnHost(globalCtx, "", []string{"/usr/sbin/setcap", "cap_sys_chroot+ep", filepath.Join(hostWork, "code")}, "", comp); err2 != nil {
			log.Printf("setcap failed: %v, output: %s", err2, out2)
			return RunResponse{Result: "Internal Error"}
		}
	}

	// build argv
	mkArgv := func() ([]string, error) {
		switch req.Language {
		case "c":
			return []string{insidePath(filepath.Join(workdir, "code"))}, nil
		case "python":
			return []string{mapToolPath("/usr/bin/python3"), insidePath(filepath.Join(workdir, "code.py"))}, nil
		case "ruby":
			return []string{mapToolPath("/usr/bin/ruby"), insidePath(filepath.Join(workdir, "code.rb"))}, nil
		}
		return nil, errors.New("unsupported language")
	}
	argv, err := mkArgv()
	if err != nil {
		return RunResponse{Result: "Unsupported language: " + req.Language}
	}

	return runProgramWithTests(req, rr, hostWork, workdir, useChrootRunner, shellPath, argv, runLim, outLimit, execLimit, globalCtx)
}

func executeCTwoStage(req RunRequest, useChrootRunner bool) RunResponse {
	buildRR, err := sandbox.PrepareRunRootWithOptions("c", sandbox.PrepareRunRootOptions{FlagDestinations: []string{"/flag2", "/env/flag2"}})
	if err != nil {
		log.Printf("Failed to prepare C compile sandbox: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	defer buildRR.Cleanup()

	buildHostWork := buildRR.WorkspaceHost
	buildEnvWorkspaceInside := filepath.Join("/env", strings.TrimPrefix(buildRR.WorkspaceRel(), "/"))
	src := filepath.Join(buildHostWork, "code.c")
	if err := ioutil.WriteFile(src, []byte(req.Code), 0644); err != nil {
		log.Printf("Failed to write C source: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	buildCaptureDir := filepath.Join(buildHostWork, ".runner")
	if err := resetDir(buildCaptureDir, 0o755); err != nil {
		log.Printf("Failed to prepare C compile capture dir: %v", err)
		return RunResponse{Result: "Internal Error"}
	}

	execLimit := defaultTimeLimitMs
	globalLimitMs := 5000
	if v := os.Getenv("RUNNER_GLOBAL_TIMEOUT_MS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			globalLimitMs = n
		}
	}
	globalCtx, globalCancel := context.WithTimeout(context.Background(), time.Duration(globalLimitMs)*time.Millisecond)
	defer globalCancel()

	toBytes := func(mb int) int { return mb * 1024 * 1024 }
	outLimit := 65536
	if v := os.Getenv("RUN_LIMIT_OUTPUT_BYTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			outLimit = n
		}
	}
	comp := sandbox.RLimits{
		CPUSeconds: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_CPU_SEC")); e == nil && n > 0 {
				return n
			}
			return 15
		}(),
		ASBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_AS_MB")); e == nil && n > 0 {
				return n
			}
			return 512
		}()),
		FSizeBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_FSIZE_MB")); e == nil && n > 0 {
				return n
			}
			return 64
		}()),
		NProc: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_NPROC")); e == nil && n > 0 {
				return n
			}
			return 128
		}(),
		NOFile: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_NOFILE")); e == nil && n > 0 {
				return n
			}
			return 512
		}(),
		OutputLimit: outLimit,
	}
	execCpu := (execLimit+999)/1000 + 1
	if globalLimitMs/1000 > 0 && execCpu > (globalLimitMs/1000) {
		execCpu = globalLimitMs / 1000
	}
	runLim := sandbox.RLimits{
		CPUSeconds: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_CPU_SEC")); e == nil && n > 0 {
				return n
			}
			return execCpu
		}(),
		ASBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_AS_MB")); e == nil && n > 0 {
				return n
			}
			return 256
		}()),
		FSizeBytes: toBytes(func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_FSIZE_MB")); e == nil && n > 0 {
				return n
			}
			return 16
		}()),
		NProc: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_NPROC")); e == nil && n > 0 {
				return n
			}
			return 64
		}(),
		NOFile: func() int {
			if n, e := strconv.Atoi(os.Getenv("RUN_LIMIT_NOFILE")); e == nil && n > 0 {
				return n
			}
			return 128
		}(),
		OutputLimit: outLimit,
	}

	compileUseChrootRunner := false
	mapToolPath := func(p string) string {
		if compileUseChrootRunner {
			return p
		}
		if strings.HasPrefix(p, "/") {
			return filepath.Join("/env", strings.TrimPrefix(p, "/"))
		}
		return p
	}
	compileShellPath := "/bin/sh"
	if !compileUseChrootRunner {
		compileShellPath = "/env/bin/sh"
	}

	compileStdoutHost, compileStderrHost, compileStdoutInside, compileStderrInside := capturePaths(buildHostWork, buildEnvWorkspaceInside, "compile")
	removeFiles(compileStdoutHost, compileStderrHost)
	compileArgs := []string{
		mapToolPath("/usr/bin/gcc"),
		filepath.Join(buildEnvWorkspaceInside, "code.c"),
		"-O2", "-pipe", "-static", "-s", "-lm",
		"-o", filepath.Join(buildEnvWorkspaceInside, "code"),
	}
	compileCmd := buildCaptureCommand(compileArgs, compileStdoutInside, compileStderrInside)
	compileRes, compileErr := sandbox.RunInChroot(globalCtx, buildRR, buildEnvWorkspaceInside, []string{compileShellPath, "-c", compileCmd}, "", comp, compileUseChrootRunner)
	compileStdout, stdoutErr := readFileLimited(compileStdoutHost, outLimit)
	if stdoutErr != nil {
		log.Printf("Failed to read C compile stdout: %v", stdoutErr)
	}
	compileStderr, stderrErr := readFileLimited(compileStderrHost, outLimit)
	if stderrErr != nil {
		log.Printf("Failed to read C compile stderr: %v", stderrErr)
	}
	removeFiles(compileStdoutHost, compileStderrHost)
	summary := combineOutput(compileStdout, compileStderr)
	if summary == "" {
		summary = combineOutput(compileRes.Stdout, compileRes.Stderr)
	}
	if compileErr != nil {
		if summary == "" {
			summary = compileErr.Error()
		}
		return RunResponse{Result: "Compile Error", Output: summary}
	}
	_ = os.Chmod(filepath.Join(buildHostWork, "code"), 0755)

	runRR, err := sandbox.PrepareRunRootWithOptions("c", sandbox.PrepareRunRootOptions{ForCBuilder: true})
	if err != nil {
		log.Printf("Failed to prepare C run sandbox: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	defer runRR.Cleanup()

	runHostWork := runRR.WorkspaceHost
	runWorkdir := runRR.WorkspaceDir()
	runWorkspaceRel := runRR.WorkspaceRel()
	runCaptureDir := filepath.Join(runHostWork, ".runner")
	if err := resetDir(runCaptureDir, 0o755); err != nil {
		log.Printf("Failed to prepare C run capture dir: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	runtimeBinaryHost := filepath.Join(runHostWork, "code")
	if err := sandbox.CopyFile(filepath.Join(buildHostWork, "code"), runtimeBinaryHost, 0o755); err != nil {
		log.Printf("Failed to copy C binary into runtime sandbox: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	if err := os.Chmod(runtimeBinaryHost, 0755); err != nil {
		log.Printf("Failed to chmod C runtime binary: %v", err)
		return RunResponse{Result: "Internal Error"}
	}
	if out2, err2 := sandbox.RunOnHost(globalCtx, "", []string{"/usr/sbin/setcap", "cap_sys_chroot+ep", runtimeBinaryHost}, "", comp); err2 != nil {
		log.Printf("setcap failed for C runtime binary: %v, output: %s", err2, out2)
		return RunResponse{Result: "Internal Error"}
	}

	insidePathForRun := func(path string) string {
		if strings.HasPrefix(path, runHostWork) {
			rel := strings.TrimPrefix(path, runHostWork)
			rel = strings.TrimPrefix(rel, "/")
			base := runWorkspaceRel
			if base == "/" {
				if rel == "" {
					return "/"
				}
				return "/" + rel
			}
			if rel == "" {
				return base
			}
			return filepath.Join(base, rel)
		}
		return path
	}

	argv := []string{insidePathForRun(filepath.Join(runWorkdir, "code"))}
	runtimeShellPath := "/bin/sh"
	if !useChrootRunner {
		runtimeShellPath = "/env/bin/sh"
	}
	return runProgramWithTests(req, runRR, runHostWork, runWorkdir, useChrootRunner, runtimeShellPath, argv, runLim, outLimit, execLimit, globalCtx)
}

func runProgramWithTests(req RunRequest, rr *sandbox.RunRoot, hostWork, workdir string, useChrootRunner bool, shellPath string, argv []string, runLim sandbox.RLimits, outLimit int, execLimit int, globalCtx context.Context) RunResponse {
	if strings.TrimSpace(req.Challenge) != "" {
		tests := getRunnerTests(req.Challenge, req.Mode)
		if len(tests) == 0 {
			return sanitizeRunResponse(req, RunResponse{Result: "Unknown challenge"})
		}
		total := 0
		for i, tc := range tests {
			execCtx, execCancel := context.WithTimeout(globalCtx, time.Duration(execLimit)*time.Millisecond)
			start := time.Now()
			stdoutHost, stderrHost, stdoutInside, stderrInside := capturePaths(hostWork, workdir, fmt.Sprintf("test-%d", i))
			removeFiles(stdoutHost, stderrHost)
			runCmd := buildCaptureCommand(argv, stdoutInside, stderrInside)
			runRes, err := sandbox.RunInChroot(execCtx, rr, workdir, []string{shellPath, "-c", runCmd}, tc.Input, runLim, useChrootRunner)
			dur := int(time.Since(start).Milliseconds())
			execCancel()
			total += dur
			runStdout, outErr := readFileLimited(stdoutHost, outLimit)
			if outErr != nil {
				log.Printf("Failed to read run stdout: %v", outErr)
			}
			runStderr, errErr := readFileLimited(stderrHost, outLimit)
			if errErr != nil {
				log.Printf("Failed to read run stderr: %v", errErr)
			}
			removeFiles(stdoutHost, stderrHost)
			combined := combineOutput(runStdout, runStderr)
			if execCtx.Err() == context.DeadlineExceeded || globalCtx.Err() == context.DeadlineExceeded {
				if combined == "" {
					combined = strings.TrimSpace(runRes.Stdout)
				}
				if req.Mode == "sample" {
					return sanitizeRunResponse(req, RunResponse{Result: "Time Limit Exceeded", Output: combined, DurationMs: total, FailedIndex: i, Expected: tc.Output})
				}
				return sanitizeRunResponse(req, RunResponse{Result: "Time Limit Exceeded", Output: combined, DurationMs: total, FailedIndex: i})
			}
			trimmed := strings.TrimSpace(runStdout)
			if err != nil {
				if combined == "" {
					combined = combineOutput(runRes.Stdout, runRes.Stderr)
				}
				log.Printf("Runtime error: %v, output: %s", err, combined)
				if req.Mode == "sample" {
					return sanitizeRunResponse(req, RunResponse{Result: "Runtime Error", Output: combined, DurationMs: total, FailedIndex: i, Expected: tc.Output})
				}
				return sanitizeRunResponse(req, RunResponse{Result: "Runtime Error", Output: combined, DurationMs: total, FailedIndex: i})
			}
			if strings.TrimSpace(trimmed) != strings.TrimSpace(tc.Output) {
				if req.Mode == "sample" {
					return sanitizeRunResponse(req, RunResponse{Result: "Wrong Answer", Output: trimmed, DurationMs: total, FailedIndex: i, Expected: tc.Output})
				}
				return sanitizeRunResponse(req, RunResponse{Result: "Wrong Answer", Output: trimmed, DurationMs: total, FailedIndex: i})
			}
			_ = sandbox.ResetChrootTmp(rr)
		}
		return sanitizeRunResponse(req, RunResponse{Result: "Success", Output: "", DurationMs: total})
	}

	execCtx, execCancel := context.WithTimeout(globalCtx, time.Duration(execLimit)*time.Millisecond)
	defer execCancel()
	start := time.Now()
	stdoutHost, stderrHost, stdoutInside, stderrInside := capturePaths(hostWork, workdir, "single")
	removeFiles(stdoutHost, stderrHost)
	runCmd := buildCaptureCommand(argv, stdoutInside, stderrInside)
	runRes, execErr := sandbox.RunInChroot(execCtx, rr, workdir, []string{shellPath, "-c", runCmd}, req.Input, runLim, useChrootRunner)
	durationMs := int(time.Since(start).Milliseconds())
	runStdout, outErr := readFileLimited(stdoutHost, outLimit)
	if outErr != nil {
		log.Printf("Failed to read run stdout: %v", outErr)
	}
	runStderr, errErr := readFileLimited(stderrHost, outLimit)
	if errErr != nil {
		log.Printf("Failed to read run stderr: %v", errErr)
	}
	removeFiles(stdoutHost, stderrHost)
	combined := combineOutput(runStdout, runStderr)
	if execCtx.Err() == context.DeadlineExceeded || globalCtx.Err() == context.DeadlineExceeded {
		if combined == "" {
			combined = strings.TrimSpace(runRes.Stdout)
		}
		return sanitizeRunResponse(req, RunResponse{Result: "Time Limit Exceeded", Output: combined, DurationMs: durationMs})
	}
	output := strings.TrimSpace(runStdout)
	if execErr != nil {
		if combined == "" {
			combined = combineOutput(runRes.Stdout, runRes.Stderr)
		}
		log.Printf("Runtime error: %v, output: %s", execErr, combined)
		if req.Mode == "sample" {
			return sanitizeRunResponse(req, RunResponse{Result: "Runtime Error", Output: combined, DurationMs: durationMs})
		}
		return sanitizeRunResponse(req, RunResponse{Result: "Runtime Error", Output: combined, DurationMs: durationMs})
	}
	wantTrim := strings.TrimSpace(req.Want)
	if wantTrim != "" {
		if strings.TrimSpace(output) == wantTrim {
			return sanitizeRunResponse(req, RunResponse{Result: "Success", Output: output, DurationMs: durationMs})
		}
		return sanitizeRunResponse(req, RunResponse{Result: "Wrong Answer", Output: output, DurationMs: durationMs})
	}
	return sanitizeRunResponse(req, RunResponse{Result: "Success", Output: output, DurationMs: durationMs})
}

func runSandboxShell(lang, command string, args []string, workdir string, keep bool) error {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return errors.New("language is required")
	}
	if command == "" {
		command = "/bin/sh"
	}
	rr, err := sandbox.PrepareRunRoot(lang)
	if err != nil {
		return err
	}
	if keep {
		log.Printf("sandbox shell: keeping runroot at %s", rr.Root)
	} else {
		defer rr.Cleanup()
	}
	if workdir == "" {
		workdir = rr.WorkspaceDir()
	}
	log.Printf("sandbox shell: language=%s workdir=%s host_work=%s", lang, workdir, rr.WorkspaceHost)
	argv := append([]string{command}, args...)
	if err := sandbox.LaunchInteractive(rr, workdir, argv); err != nil {
		return err
	}
	if keep {
		log.Printf("sandbox shell: runroot retained at %s", rr.Root)
	}
	return nil
}

func capturePaths(hostWork, workdir, base string) (stdoutHost, stderrHost, stdoutInside, stderrInside string) {
	stdoutHost = filepath.Join(hostWork, ".runner", base+".stdout")
	stderrHost = filepath.Join(hostWork, ".runner", base+".stderr")
	stdoutInside = filepath.Join(workdir, ".runner", base+".stdout")
	stderrInside = filepath.Join(workdir, ".runner", base+".stderr")
	return
}

func removeFiles(paths ...string) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		_ = os.Remove(p)
	}
}

func buildCaptureCommand(argv []string, stdoutPath, stderrPath string) string {
	var sb strings.Builder
	sb.WriteString("rm -f ")
	sb.WriteString(shellQuote(stdoutPath))
	sb.WriteByte(' ')
	sb.WriteString(shellQuote(stderrPath))
	sb.WriteString(" && exec ")
	sb.WriteString(joinShellArgs(argv))
	sb.WriteString(" > ")
	sb.WriteString(shellQuote(stdoutPath))
	sb.WriteString(" 2> ")
	sb.WriteString(shellQuote(stderrPath))
	return sb.String()
}

func joinShellArgs(argv []string) string {
	var sb strings.Builder
	for i, arg := range argv {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(shellQuote(arg))
	}
	return sb.String()
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func readFileLimited(path string, limit int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if limit > 0 && len(data) > limit {
		data = data[:limit]
	}
	return string(data), nil
}

func resetDir(path string, perm os.FileMode) error {
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return err
	}
	return os.Chmod(path, perm)
}

func combineOutput(stdout, stderr string) string {
	s := strings.TrimSpace(stdout)
	t := strings.TrimSpace(stderr)
	if s == "" {
		return t
	}
	if t == "" {
		return s
	}
	return strings.TrimSpace(s + "\n" + t)
}

func main() {
	loadDotEnv()

	var shellArgsFlag stringSliceFlag
	shellMode := flag.Bool("sandbox-shell", false, "launch an interactive sandbox shell and exit")
	shellLang := flag.String("sandbox-shell-lang", "", "language environment for sandbox shell (c/go/python/ruby/base)")
	shellCmd := flag.String("sandbox-shell-cmd", "/work/bin/sh", "command to execute inside the sandbox shell")
	shellWorkdir := flag.String("sandbox-shell-workdir", "", "working directory inside the sandbox root")
	shellKeep := flag.Bool("sandbox-shell-keep", false, "retain sandbox runroot after the shell exits")
	flag.Var(&shellArgsFlag, "sandbox-shell-arg", "additional argument for the sandbox shell command (repeatable)")
	flag.Parse()

	if *shellMode || *shellLang != "" {
		lang := strings.TrimSpace(*shellLang)
		if lang == "" {
			lang = "base"
		}
		args := append([]string{}, []string(shellArgsFlag)...)
		if err := runSandboxShell(lang, *shellCmd, args, *shellWorkdir, *shellKeep); err != nil {
			log.Fatalf("sandbox shell failed: %v", err)
		}
		return
	}

	if flag.NArg() > 0 {
		log.Fatalf("unexpected arguments: %v", flag.Args())
	}

	initRunnerDB()
	seedInitialChallenges()
	initWorkerPool()
	http.HandleFunc("/run", runHandler)
	http.HandleFunc("/challenge", challengeMetaHandler)
	log.Println("Runner listening on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

// resetChrootTmp clears and recreates /tmp inside a runroot with sticky bit
