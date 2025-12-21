package gohelper

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"

	sandbox "goexe-runner/internal/sandbox"
)

const (
	compileTimeout         = 2 * time.Second
	compileLogLimit        = 4096
	defaultExecTimeLimitMs = 1000
)

// TestCase represents a single input/output pair for execution.
type TestCase struct {
	Input    string `json:"input"`
	Output   string `json:"output"`
	IsSample bool   `json:"is_sample"`
}

// Request holds all parameters required to compile and execute Go code.
type Request struct {
	Code            string
	Mode            string
	GlobalTimeoutMs int
	OutputLimit     int
	SandboxEnv      string
	Tests           []TestCase
}

// Response mirrors the runner's RunResponse payload.
type Response struct {
	Result      string `json:"result"`
	Output      string `json:"output,omitempty"`
	DurationMs  int    `json:"duration_ms,omitempty"`
	FailedIndex int    `json:"failed_index,omitempty"`
	Expected    string `json:"expected,omitempty"`
}

// Execute compiles the provided Go code and runs it against the supplied tests.
func Execute(ctx context.Context, req Request) Response {
	dev, err := os.Open("/dev")
	if err != nil {
		return Response{Result: "Internal Error"}
	}
	defer dev.Close()
	devfd := dev.Fd()
	_, err = unix.Openat(int(devfd), "null", os.O_RDONLY, 0)
	if err != nil {
		return Response{Result: "Internal Error"}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	execLimit := defaultExecTimeLimitMs

	globalLimitMs := req.GlobalTimeoutMs
	if globalLimitMs <= 0 {
		globalLimitMs = 30000
		if v := os.Getenv("RUNNER_GLOBAL_TIMEOUT_MS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				globalLimitMs = n
			}
		}
	}

	outLimit := req.OutputLimit
	if outLimit <= 0 {
		outLimit = 65536
		if v := os.Getenv("RUN_LIMIT_OUTPUT_BYTES"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				outLimit = n
			}
		}
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "judge"
	}
	singleMode := mode == "single"
	tests := req.Tests
	if len(tests) == 0 {
		return Response{Result: "Unknown challenge"}
	}
	revealExpected := mode == "sample" || singleMode
	sanitize := func(res Response) Response {
		if !revealExpected {
			res.Output = ""
			res.Expected = ""
		}
		return res
	}

	envBase := strings.TrimSpace(req.SandboxEnv)
	if envBase != "" {
		_ = os.Setenv("SANDBOX_ENVS_DIR", envBase)
	} else if strings.TrimSpace(os.Getenv("SANDBOX_ENVS_DIR")) == "" {
		_ = os.Setenv("SANDBOX_ENVS_DIR", "/opt/sandbox-envs")
	}

	buildRR, err := sandbox.PrepareRunRootWithOptions("go", sandbox.PrepareRunRootOptions{ForGoBuilder: true})
	if err != nil {
		log.Printf("go helper: prepare build runroot failed: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}
	defer buildRR.Cleanup()

	buildWorkspaceHost := buildRR.WorkspaceHost
	buildWorkspaceInside := buildRR.WorkspaceDir()

	if err := resetDir(filepath.Join(buildWorkspaceHost, ".runner"), 0o755); err != nil {
		log.Printf("go helper: failed to prepare build capture dir: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}

	codeHostPath := filepath.Join(buildWorkspaceHost, "code.go")
	if err := os.WriteFile(codeHostPath, []byte(req.Code), 0o644); err != nil {
		log.Printf("go helper: failed to write code: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}

	cacheDirs := []string{
		filepath.Join(buildRR.TmpHost, "go-build-cache"),
		filepath.Join(buildRR.TmpHost, "go-mod-cache"),
	}
	for _, dir := range cacheDirs {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			log.Printf("go helper: failed to prepare go cache dir %s: %v", dir, err)
			return sanitize(Response{Result: "Internal Error"})
		}
	}

	globalCtx := ctx
	var globalCancel context.CancelFunc
	var globalDeadline time.Time
	if globalLimitMs > 0 {
		globalCtx, globalCancel = context.WithTimeout(ctx, time.Duration(globalLimitMs)*time.Millisecond)
		globalDeadline = time.Now().Add(time.Duration(globalLimitMs) * time.Millisecond)
		defer globalCancel()
	}

	compileStdoutHost, compileStderrHost, compileStdoutInside, compileStderrInside := capturePaths(buildWorkspaceHost, buildWorkspaceInside, "compile")
	removeFiles(compileStdoutHost, compileStderrHost)

	compileArgs := []string{
		"env",
		"GOCACHE=/tmp/go-build-cache",
		"GOMODCACHE=/tmp/go-mod-cache",
		"/usr/bin/go", "build",
		"-o", filepath.Join(buildWorkspaceInside, "code"),
		filepath.Join(buildWorkspaceInside, "code.go"),
	}
	compileCmd := buildCaptureCommand(compileArgs, compileStdoutInside, compileStderrInside)
	compileCtx, compileCancel := context.WithTimeout(globalCtx, compileTimeout)
	compileRes, compileErr := sandbox.RunInChroot(compileCtx, buildRR, buildWorkspaceInside, []string{"/bin/sh", "-c", compileCmd}, "", buildGoCompileLimits(outLimit), true)
	compileCancel()

	compileStdout, stdoutErr := readFileLimited(compileStdoutHost, outLimit)
	if stdoutErr != nil {
		log.Printf("go helper: failed to read compile stdout: %v", stdoutErr)
	}
	compileStderr, stderrErr := readFileLimited(compileStderrHost, outLimit)
	if stderrErr != nil {
		log.Printf("go helper: failed to read compile stderr: %v", stderrErr)
	}
	removeFiles(compileStdoutHost, compileStderrHost)
	summary := combineOutputs(compileStdout, compileStderr)
	if summary == "" {
		summary = combineOutputs(compileRes.Stdout, compileRes.Stderr)
	}
	summary = strings.TrimSpace(summary)

	if compileErr != nil {
		if errors.Is(compileCtx.Err(), context.DeadlineExceeded) {
			log.Printf("go helper: compile deadline exceeded after %s", compileTimeout)
		}
		if summary == "" {
			summary = compileErr.Error()
		}
		log.Printf("go helper: compile failed: %s", clipForLog(summary, compileLogLimit))
		return sanitize(Response{Result: "Compile Error", Output: summary, FailedIndex: -1})
	}

	binaryHostPath := filepath.Join(buildWorkspaceHost, "code")
	if _, statErr := os.Stat(binaryHostPath); statErr != nil {
		log.Printf("go helper: compiled binary missing: %v", statErr)
		return sanitize(Response{Result: "Internal Error"})
	}

	runRR, err := sandbox.PrepareRunRootWithOptions("go", sandbox.PrepareRunRootOptions{ForCBuilder: true})
	if err != nil {
		log.Printf("go helper: prepare runtime runroot failed: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}
	defer runRR.Cleanup()

	runWorkspaceHost := runRR.WorkspaceHost
	runWorkspaceRel := runRR.WorkspaceRel()
	runWorkspaceInside := runRR.WorkspaceDir()

	runtimeBinaryHostPath := filepath.Join(runWorkspaceHost, "code")
	if err := sandbox.CopyFile(binaryHostPath, runtimeBinaryHostPath, 0o755); err != nil {
		log.Printf("go helper: failed to copy binary into runtime workspace: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}
	if err := os.Chmod(runtimeBinaryHostPath, 0o755); err != nil {
		log.Printf("go helper: failed to chmod runtime binary: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}
	if _, err := sandbox.RunOnHost(context.Background(), "", []string{"/usr/sbin/setcap", "cap_sys_chroot+ep", runtimeBinaryHostPath}, "", sandbox.RLimits{OutputLimit: 4096}); err != nil {
		log.Printf("go helper: setcap failed: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}

	if err := resetDir(filepath.Join(runWorkspaceHost, ".runner"), 0o755); err != nil {
		log.Printf("go helper: failed to reset runtime capture dir: %v", err)
		return sanitize(Response{Result: "Internal Error"})
	}

	insidePath := func(p string) string {
		if strings.HasPrefix(p, runWorkspaceHost) {
			rel := strings.TrimPrefix(p, runWorkspaceHost)
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
		return p
	}

	runLim := buildGoRunLimits(execLimit, globalLimitMs, outLimit)
	argv := []string{insidePath(filepath.Join(runWorkspaceInside, "code"))}
	totalDuration := 0
	trim := func(s string) string { return strings.TrimSpace(s) }
	lastStdout := ""

	for i, tc := range tests {
		if !globalDeadline.IsZero() && time.Now().After(globalDeadline) {
			expected := ""
			if revealExpected {
				expected = tc.Output
			}
			return sanitize(Response{Result: "Time Limit Exceeded", DurationMs: totalDuration, FailedIndex: i, Expected: expected})
		}

		execCtx, cancel := context.WithTimeout(globalCtx, time.Duration(execLimit)*time.Millisecond)
		start := time.Now()
		stdoutHost, stderrHost, stdoutInside, stderrInside := capturePaths(runWorkspaceHost, runWorkspaceInside, fmt.Sprintf("test-%d", i))
		removeFiles(stdoutHost, stderrHost)
		runCmd := buildCaptureCommand(argv, stdoutInside, stderrInside)
		shellPath := "/env/bin/sh"
		runRes, err := sandbox.RunInChroot(execCtx, runRR, runWorkspaceInside, []string{shellPath, "-c", runCmd}, tc.Input, runLim, false)
		duration := int(time.Since(start).Milliseconds())
		cancel()
		totalDuration += duration

		runStdout, outErr := readFileLimited(stdoutHost, outLimit)
		if outErr != nil {
			log.Printf("go helper: failed to read run stdout: %v", outErr)
		}
		runStderr, errErr := readFileLimited(stderrHost, outLimit)
		if errErr != nil {
			log.Printf("go helper: failed to read run stderr: %v", errErr)
		}
		removeFiles(stdoutHost, stderrHost)

		trimmedStdout := trim(runStdout)
		lastStdout = trimmedStdout
		combined := combineOutputs(runStdout, runStderr)
		if combined == "" {
			combined = trim(runRes.Stdout + "\n" + runRes.Stderr)
		}
		if err != nil {
			if errors.Is(execCtx.Err(), context.DeadlineExceeded) || errors.Is(globalCtx.Err(), context.DeadlineExceeded) {
				expected := ""
				if revealExpected {
					expected = tc.Output
				}
				return sanitize(Response{Result: "Time Limit Exceeded", Output: trimmedStdout, DurationMs: totalDuration, FailedIndex: i, Expected: expected})
			}
			log.Printf("go helper: runtime error: %v output=%s", err, combined)
			expected := ""
			if revealExpected {
				expected = tc.Output
			}
			return sanitize(Response{Result: "Runtime Error", Output: combined, DurationMs: totalDuration, FailedIndex: i, Expected: expected})
		}

		if errors.Is(execCtx.Err(), context.DeadlineExceeded) || errors.Is(globalCtx.Err(), context.DeadlineExceeded) {
			expected := ""
			if revealExpected {
				expected = tc.Output
			}
			return sanitize(Response{Result: "Time Limit Exceeded", Output: trimmedStdout, DurationMs: totalDuration, FailedIndex: i, Expected: expected})
		}

		expectedTrim := trim(tc.Output)
		if singleMode {
			if expectedTrim != "" && trim(trimmedStdout) != expectedTrim {
				expected := ""
				if revealExpected {
					expected = tc.Output
				}
				return sanitize(Response{Result: "Wrong Answer", Output: trimmedStdout, DurationMs: totalDuration, FailedIndex: i, Expected: expected})
			}
		} else {
			if trim(trimmedStdout) != expectedTrim {
				expected := ""
				if revealExpected {
					expected = tc.Output
				}
				return sanitize(Response{Result: "Wrong Answer", Output: trimmedStdout, DurationMs: totalDuration, FailedIndex: i, Expected: expected})
			}
		}

		if err := sandbox.ResetChrootTmp(runRR); err != nil {
			log.Printf("go helper: failed to reset tmp: %v", err)
			return sanitize(Response{Result: "Internal Error"})
		}
	}

	resp := Response{Result: "Success", DurationMs: totalDuration, FailedIndex: -1}
	if singleMode {
		resp.Output = lastStdout
	}
	return sanitize(resp)
}

func buildGoCompileLimits(outputLimit int) sandbox.RLimits {
	toBytes := func(mb int) int { return mb * 1024 * 1024 }

	cpuSeconds := 15
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_CPU_SEC")); err == nil && n > 0 {
		cpuSeconds = n
	}
	asLimit := toBytes(512)
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_AS_MB")); err == nil && n > 0 {
		asLimit = toBytes(n)
	}
	fsizeLimit := toBytes(64)
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_FSIZE_MB")); err == nil && n > 0 {
		fsizeLimit = toBytes(n)
	}
	nproc := 128
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_NPROC")); err == nil && n > 0 {
		nproc = n
	}
	nofile := 512
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_COMPILE_NOFILE")); err == nil && n > 0 {
		nofile = n
	}

	return sandbox.RLimits{
		CPUSeconds:  cpuSeconds,
		ASBytes:     asLimit,
		FSizeBytes:  fsizeLimit,
		NProc:       nproc,
		NOFile:      nofile,
		OutputLimit: outputLimit,
	}
}

func buildGoRunLimits(execLimit, globalLimitMs, outputLimit int) sandbox.RLimits {
	toBytes := func(mb int) int { return mb * 1024 * 1024 }

	execCpu := (execLimit+999)/1000 + 1
	if globalLimitMs > 0 {
		if maxCpu := globalLimitMs / 1000; maxCpu > 0 && execCpu > maxCpu {
			execCpu = maxCpu
		}
	}

	cpuSeconds := execCpu
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_CPU_SEC")); err == nil && n > 0 {
		cpuSeconds = n
	}
	asLimit := toBytes(1024)
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_AS_MB")); err == nil && n > 0 {
		asLimit = toBytes(n)
	}
	fsizeLimit := toBytes(16)
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_FSIZE_MB")); err == nil && n > 0 {
		fsizeLimit = toBytes(n)
	}
	nproc := 64
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_NPROC")); err == nil && n > 0 {
		nproc = n
	}
	nofile := 128
	if n, err := strconv.Atoi(os.Getenv("RUN_LIMIT_NOFILE")); err == nil && n > 0 {
		nofile = n
	}

	return sandbox.RLimits{
		CPUSeconds:  cpuSeconds,
		ASBytes:     asLimit,
		FSizeBytes:  fsizeLimit,
		NProc:       nproc,
		NOFile:      nofile,
		OutputLimit: outputLimit,
	}
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

func resetDir(path string, perm fs.FileMode) error {
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return err
	}
	return os.Chmod(path, perm)
}

func combineOutputs(stdout, stderr string) string {
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

func clipForLog(s string, limit int) string {
	if limit <= 0 || len(s) <= limit {
		return s
	}
	return s[:limit] + "... (truncated)"
}
