package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const chrootRunPath = "/usr/local/bin/chroot-run"

type bindMount struct {
	host     string
	target   string
	readOnly bool
}

type RunRoot struct {
	Root              string
	WorkHost          string
	WorkInside        string
	WorkspaceHost     string
	WorkspaceInside   string
	WorkspaceRelative string
	TmpHost           string
	TmpInside         string
	ChrootBin         string
	EnvRoot           string
	mounts            []bindMount
	cleanup           func()
}

func (rr *RunRoot) WorkDir() string {
	if rr == nil || rr.WorkInside == "" {
		return "/env"
	}
	return rr.WorkInside
}

func (rr *RunRoot) WorkspaceDir() string {
	if rr == nil || rr.WorkspaceInside == "" {
		return rr.WorkDir()
	}
	return rr.WorkspaceInside
}

func (rr *RunRoot) WorkspaceRel() string {
	if rr == nil || rr.WorkspaceRelative == "" {
		return "/"
	}
	return rr.WorkspaceRelative
}

func (rr *RunRoot) TmpDir() string {
	if rr == nil || rr.TmpInside == "" {
		return "/tmp"
	}
	return rr.TmpInside
}

func (rr *RunRoot) Cleanup() {
	if rr != nil && rr.cleanup != nil {
		rr.cleanup()
	}
}

type PrepareRunRootOptions struct {
	ForGoBuilder     bool
	ForCBuilder      bool
	FlagDestinations []string
}

// PrepareRunRoot constructs a per-run chroot using bind mounts instead of copying the rootfs.
func PrepareRunRoot(language string) (*RunRoot, error) {
	return PrepareRunRootWithOptions(language, PrepareRunRootOptions{})
}

// PrepareRunRootWithOptions constructs a per-run chroot using bind mounts instead of copying the rootfs.
func PrepareRunRootWithOptions(language string, opts PrepareRunRootOptions) (*RunRoot, error) {
	baseEnv := strings.TrimSpace(os.Getenv("SANDBOX_ENVS_DIR"))
	if baseEnv == "" {
		baseEnv = "/opt/sandbox-envs"
	}
	envRoot := filepath.Join(baseEnv, language)
	if real, err := filepath.EvalSymlinks(envRoot); err == nil && real != "" {
		envRoot = real
	}
	if st, err := os.Stat(envRoot); err != nil || !st.IsDir() {
		return nil, fmt.Errorf("runtime environment not found: %s", envRoot)
	}
	criticalPaths := []string{
		filepath.Join(envRoot, "usr/bin/gcc"),
		filepath.Join(envRoot, "usr/bin/python3"),
		filepath.Join(envRoot, "bin/sh"),
	}
	for _, p := range criticalPaths {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("runtime environment incomplete: missing %s (%v)", p, err)
		}
	}

	runRootsDir := strings.TrimSpace(os.Getenv("SANDBOX_RUNROOT_DIR"))
	if runRootsDir == "" {
		runRootsDir = filepath.Join(os.TempDir(), "sandbox-runroots")
	}
	if err := os.MkdirAll(runRootsDir, 0o755); err != nil {
		return nil, err
	}

	parent, err := os.MkdirTemp(runRootsDir, "runroot-")
	if err != nil {
		return nil, err
	}

	rr := &RunRoot{
		Root:       parent,
		WorkInside: "/env",
		TmpInside:  "/tmp",
	}
	rr.WorkspaceInside = "/workspace"
	rr.WorkspaceRelative = rr.WorkspaceInside
	cleanup := func() {
		_ = os.RemoveAll(parent)
	}
	rr.cleanup = cleanup

	if err := os.MkdirAll(rr.Root, 0o755); err != nil {
		cleanup()
		return nil, err
	}

	rr.WorkHost = filepath.Join(rr.Root, strings.TrimPrefix(rr.WorkDir(), "/"))
	if err := ensureDirWithPerm(rr.WorkHost, 0o755); err != nil {
		cleanup()
		return nil, err
	}

	workspaceRelTrim := strings.TrimPrefix(rr.WorkspaceInside, "/")
	rr.WorkspaceHost = filepath.Join(rr.Root, workspaceRelTrim)
	if err := ensureDirWithPerm(rr.WorkspaceHost, 0o755); err != nil {
		cleanup()
		return nil, err
	}
	rr.TmpHost = filepath.Join(rr.Root, strings.TrimPrefix(rr.TmpDir(), "/"))
	if err := ensureDirWithPerm(rr.TmpHost, 0o1777); err != nil {
		cleanup()
		return nil, err
	}
	if err := ensureDirWithPerm(filepath.Join(rr.WorkHost, "dev"), 0o755); err != nil {
		cleanup()
		return nil, err
	}
	if err := ensureDirWithPerm(filepath.Join(rr.WorkHost, "proc"), 0o755); err != nil {
		cleanup()
		return nil, err
	}

	mounts := make([]bindMount, 0, 32)
	envMounts, err := buildEnvMounts(rr.WorkHost, envRoot)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("prepare environment mounts: %w", err)
	}
	mounts = append(mounts, envMounts...)
	workspaceRootTarget := rr.WorkspaceHost
	if err := ensureDirWithPerm(workspaceRootTarget, 0o755); err != nil {
		cleanup()
		return nil, err
	}
	mounts = append(mounts, bindMount{host: rr.WorkspaceHost, target: workspaceRootTarget, readOnly: false})
	envWorkspaceTarget := filepath.Join(rr.WorkHost, workspaceRelTrim)
	if err := ensureDirWithPerm(envWorkspaceTarget, 0o755); err != nil {
		cleanup()
		return nil, err
	}
	mounts = append(mounts, bindMount{host: rr.WorkspaceHost, target: envWorkspaceTarget, readOnly: false})
	envTmpTarget := filepath.Join(rr.WorkHost, "tmp")
	if err := ensureDirWithPerm(envTmpTarget, 0o1777); err != nil {
		cleanup()
		return nil, err
	}
	mounts = append(mounts, bindMount{host: rr.TmpHost, target: envTmpTarget, readOnly: false})
	mounts = append(mounts, buildDeviceMounts(rr.WorkHost)...)
	if !opts.ForCBuilder {
		if st, err := os.Stat("/flag2"); err == nil {
			flagTargets := opts.FlagDestinations
			if len(flagTargets) == 0 {
				flagTargets = append(flagTargets, "/flag2")
			}
			if opts.ForGoBuilder {
				flagTargets = append(flagTargets,
					filepath.Join(rr.WorkspaceRel(), "flag2"),
					filepath.Join("/env", strings.TrimPrefix(rr.WorkspaceRel(), "/"), "flag2"),
				)
			}
			seenTargets := make(map[string]struct{}, len(flagTargets))
			for _, dest := range flagTargets {
				dest = filepath.Clean(dest)
				if dest == "." || dest == "" {
					continue
				}
				if !strings.HasPrefix(dest, "/") {
					dest = "/" + dest
				}
				if _, ok := seenTargets[dest]; ok {
					continue
				}
				seenTargets[dest] = struct{}{}
				target := filepath.Join(rr.Root, strings.TrimPrefix(dest, "/"))
				if err := ensurePlaceholder(target, st.IsDir(), st.Mode().Perm()); err != nil {
					cleanup()
					return nil, err
				}
				mounts = append(mounts, bindMount{host: "/flag2", target: target, readOnly: true})
			}
		} else if !os.IsNotExist(err) {
			log.Printf("sandbox: failed mounting /flag2: %v", err)
			return nil, err
		}
	}
	rr.mounts = dedupeAndSortMounts(mounts)

	rootRunnerDir := filepath.Join(rr.Root, ".runner")
	if err := ensureDirWithPerm(rootRunnerDir, 0o755); err != nil {
		cleanup()
		return nil, err
	}
	chrootRunRootCopy := filepath.Join(rootRunnerDir, filepath.Base(chrootRunPath))
	if err := CopyFile(chrootRunPath, chrootRunRootCopy, 0o755); err != nil {
		cleanup()
		return nil, err
	}
	if err := os.Chmod(chrootRunRootCopy, 0o755); err != nil {
		cleanup()
		return nil, err
	}
	if _, err := RunOnHost(context.Background(), "", []string{"/usr/sbin/setcap", "cap_sys_chroot+ep", chrootRunRootCopy}, "", RLimits{OutputLimit: 4096}); err != nil {
		cleanup()
		return nil, err
	}
	rr.ChrootBin = "/.runner/" + filepath.Base(chrootRunPath)
	rr.EnvRoot = envRoot

	if strings.EqualFold(os.Getenv("SANDBOX_KEEP_RUNROOT"), "1") {
		rr.cleanup = func() {}
		log.Printf("sandbox: keeping runroot at %s", parent)
	} else {
		rr.cleanup = cleanup
	}
	return rr, nil
}

func buildEnvMounts(root, envRoot string) ([]bindMount, error) {
	entries, err := os.ReadDir(envRoot)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	skip := map[string]struct{}{
		"tmp":   {},
		"dev":   {},
		"proc":  {},
		"sys":   {},
		"run":   {},
		"work":  {},
		"mnt":   {},
		"media": {},
	}

	mounts := make([]bindMount, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}
		if _, ok := skip[name]; ok {
			continue
		}
		hostPath := filepath.Join(envRoot, name)
		relTarget := name
		target := filepath.Join(root, relTarget)

		if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(hostPath)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return nil, err
			}
			_ = os.Remove(target)
			if err := os.Symlink(link, target); err != nil {
				return nil, err
			}
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		mode := info.Mode()
		if mode.IsDir() {
			if err := ensurePlaceholder(target, true, mode.Perm()); err != nil {
				return nil, err
			}
		} else {
			if err := ensurePlaceholder(target, false, mode.Perm()); err != nil {
				return nil, err
			}
		}
		mounts = append(mounts, bindMount{host: hostPath, target: target, readOnly: true})
	}
	return mounts, nil
}

func buildDeviceMounts(root string) []bindMount {
	devicePaths := []string{"/dev/null", "/dev/zero", "/dev/urandom", "/dev/random", "/dev/tty"}
	mounts := make([]bindMount, 0, len(devicePaths))
	for _, host := range devicePaths {
		if _, err := os.Stat(host); err != nil {
			continue
		}
		target := filepath.Join(root, strings.TrimPrefix(host, "/"))
		if err := ensurePlaceholder(target, false, 0o666); err != nil {
			log.Printf("sandbox: failed to prepare device mount %s: %v", target, err)
			continue
		}
		mounts = append(mounts, bindMount{host: host, target: target, readOnly: false})
	}
	return mounts
}

func dedupeAndSortMounts(mounts []bindMount) []bindMount {
	seen := make(map[string]struct{}, len(mounts))
	out := make([]bindMount, 0, len(mounts))
	for _, m := range mounts {
		key := m.host + "->" + m.target
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].target == out[j].target {
			if out[i].host == out[j].host {
				return !out[i].readOnly && out[j].readOnly
			}
			return out[i].host < out[j].host
		}
		return out[i].target < out[j].target
	})
	return out
}

// RunInChroot executes argv within the given chroot/workdir applying resource limits via nsjail.
// When useChrootRunner is false, the command is executed directly under nsjail without invoking chroot-run.
type RunResult struct {
	Stdout string
	Stderr string
}

func RunInChroot(ctx context.Context, rr *RunRoot, workdir string, argv []string, stdin string, lim RLimits, useChrootRunner bool) (RunResult, error) {
	if len(argv) == 0 {
		return RunResult{}, errors.New("no argv provided")
	}
	if rr == nil {
		return RunResult{}, errors.New("runroot not prepared")
	}
	nsjailPath := os.Getenv("NSJAIL_PATH")
	if nsjailPath == "" {
		nsjailPath = "/usr/bin/nsjail"
	}
	configPath, err := nsjailConfigPath()
	if err != nil {
		return RunResult{}, fmt.Errorf("prepare nsjail config: %w", err)
	}
	if workdir == "" {
		workdir = rr.WorkDir()
	}

	pathEnv := "PATH=/.runner:/usr/local/bin:/usr/bin:/bin"
	if !useChrootRunner {
		pathEnv = "PATH=/.runner:/env/usr/local/bin:/env/usr/bin:/env/bin:/usr/local/bin:/usr/bin:/bin"
	}
	debugDirsEnv := strings.TrimSpace(os.Getenv("SANDBOX_DEBUG_DIRS"))
	debugDepthEnv := strings.TrimSpace(os.Getenv("SANDBOX_DEBUG_DIR_DEPTH"))
	nsArgs := []string{
		"--config", configPath,
		"--chroot", rr.Root,
		"--cwd", "/",
		"--keep_caps",
		"--cap", "CAP_SYS_CHROOT",
		"--disable_no_new_privs",
		"--keep_env",
		"--env", pathEnv,
		"--env", "HOME=/tmp",
		"--env", "TMPDIR=/tmp",
		"--env", "LANG=C.UTF-8",
		"--env", "LD=/usr/bin/x86_64-linux-gnu-ld",
	}
	if !useChrootRunner && rr.EnvRoot != "" {
		type mountSpec struct {
			subdir string
			dest   string
		}
		extra := []mountSpec{
			{subdir: "bin", dest: "/bin"},
			{subdir: "lib", dest: "/lib"},
			{subdir: "lib64", dest: "/lib64"},
			{subdir: "usr", dest: "/usr"},
			{subdir: "etc", dest: "/etc"},
		}
		for _, ms := range extra {
			hostPath := filepath.Join(rr.EnvRoot, ms.subdir)
			if st, err := os.Stat(hostPath); err == nil && st.IsDir() {
				destPath := filepath.Join(rr.Root, strings.TrimPrefix(ms.dest, "/"))
				_ = ensureDirWithPerm(destPath, 0o755)
				nsArgs = append(nsArgs, "--bindmount_ro", fmt.Sprintf("%s:%s", hostPath, ms.dest))
			}
		}
	}
	cwdIndex := -1
	for i := 0; i < len(nsArgs); i++ {
		if nsArgs[i] == "--cwd" && i+1 < len(nsArgs) {
			cwdIndex = i + 1
			break
		}
	}
	for _, m := range rr.mounts {
		option := "--bindmount"
		if m.readOnly {
			option = "--bindmount_ro"
		}
		dest := strings.TrimPrefix(m.target, rr.Root)
		if !strings.HasPrefix(dest, "/") {
			dest = "/" + dest
		}
		dest = filepath.Clean(dest)
		nsArgs = append(nsArgs, option, fmt.Sprintf("%s:%s", m.host, dest))
	}
	chrootDest := rr.WorkDir()
	if chrootDest == "" {
		chrootDest = "/"
	}
	innerWorkdir := strings.TrimSpace(workdir)
	if innerWorkdir == "" {
		innerWorkdir = "/"
	} else if !strings.HasPrefix(innerWorkdir, "/") {
		innerWorkdir = "/" + innerWorkdir
	}
	nsArgs = append(nsArgs, "--")
	if useChrootRunner {
		binPath := rr.ChrootBin
		if binPath == "" {
			binPath = chrootRunPath
		}
		cmdArgs := []string{binPath, chrootDest, innerWorkdir, "--"}
		cmdArgs = append(cmdArgs, argv...)
		nsArgs = append(nsArgs, cmdArgs...)
	} else {
		// execute directly under nsjail root with desired cwd
		if cwdIndex >= 0 {
			if innerWorkdir != "" && innerWorkdir != "/" {
				nsArgs[cwdIndex] = innerWorkdir
			} else {
				nsArgs[cwdIndex] = "/"
			}
		}
		nsArgs = append(nsArgs, argv...)
	}

	cmd := exec.CommandContext(ctx, nsjailPath, nsArgs...)
	cmd.Env = []string{
		pathEnv,
		"HOME=/tmp",
		"TMPDIR=/tmp",
		"LANG=C.UTF-8",
		"LD=/usr/bin/x86_64-linux-gnu-ld",
	}
	if debugDirsEnv != "" {
		cmd.Env = append(cmd.Env, "CHROOT_RUN_DEBUG_DIRS="+debugDirsEnv)
	}
	if debugDepthEnv != "" {
		cmd.Env = append(cmd.Env, "CHROOT_RUN_DEBUG_DEPTH="+debugDepthEnv)
	}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	stdoutBuf := &safeCapBuffer{max: lim.OutputLimit}
	stderrBuf := &safeCapBuffer{max: lim.OutputLimit}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	runErr := cmd.Run()

	result := RunResult{
		Stdout: stdoutBuf.String(),
		Stderr: strings.TrimSpace(stderrBuf.String()),
	}
	if runErr != nil {
		log.Printf("sandbox: nsjail/chroot-run failed: %v (cmd=%s %s)", runErr, nsjailPath, strings.Join(nsArgs, " "))
		if errStr := strings.TrimSpace(stderrBuf.String()); errStr != "" {
			log.Printf("sandbox: nsjail stderr: %s", errStr)
		}
	}
	return result, runErr
}

func LaunchInteractive(rr *RunRoot, workdir string, argv []string) error {
	if rr == nil {
		return errors.New("runroot is required")
	}
	if len(argv) == 0 {
		argv = []string{"/bin/sh"}
	}
	nsjailPath := os.Getenv("NSJAIL_PATH")
	if nsjailPath == "" {
		nsjailPath = "/usr/bin/nsjail"
	}
	configPath, err := nsjailConfigPath()
	if err != nil {
		return fmt.Errorf("prepare nsjail config: %w", err)
	}
	mapHostPath := func(p string) string {
		if p == "" {
			return p
		}
		if strings.HasPrefix(p, rr.WorkspaceHost) {
			rel := strings.TrimPrefix(p, rr.WorkspaceHost)
			rel = strings.TrimPrefix(rel, "/")
			base := rr.WorkspaceRel()
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
		if strings.HasPrefix(p, rr.WorkHost) {
			rel := strings.TrimPrefix(p, rr.WorkHost)
			rel = strings.TrimPrefix(rel, "/")
			if rel == "" {
				return rr.WorkspaceRel()
			}
			return filepath.Join(rr.WorkspaceRel(), rel)
		}
		return p
	}

	chrootDest := rr.WorkDir()
	if chrootDest == "" {
		chrootDest = "/env"
	}
	insideWorkdir := strings.TrimSpace(workdir)
	if insideWorkdir == "" {
		insideWorkdir = rr.WorkspaceRel()
	} else {
		insideWorkdir = mapHostPath(insideWorkdir)
		if insideWorkdir == "" {
			insideWorkdir = rr.WorkspaceRel()
		}
	}
	if insideWorkdir == "" {
		insideWorkdir = "/"
	} else if !strings.HasPrefix(insideWorkdir, "/") {
		insideWorkdir = "/" + insideWorkdir
	}
	pathEnv := "PATH=/.runner:/usr/local/bin:/usr/bin:/bin"
	debugDirsEnv := strings.TrimSpace(os.Getenv("SANDBOX_DEBUG_DIRS"))
	debugDepthEnv := strings.TrimSpace(os.Getenv("SANDBOX_DEBUG_DIR_DEPTH"))
	nsArgs := []string{
		"--config", configPath,
		"--chroot", rr.Root,
		"--cwd", insideWorkdir,
		"--keep_caps",
		"--cap", "CAP_SYS_CHROOT",
		"--disable_no_new_privs",
		"--keep_env",
		"--env", pathEnv,
		"--env", "HOME=/tmp",
		"--env", "TMPDIR=/tmp",
		"--env", "LANG=C.UTF-8",
		"--env", "LD=/usr/bin/x86_64-linux-gnu-ld",
	}
	for _, m := range rr.mounts {
		option := "--bindmount"
		if m.readOnly {
			option = "--bindmount_ro"
		}
		dest := strings.TrimPrefix(m.target, rr.Root)
		if !strings.HasPrefix(dest, "/") {
			dest = "/" + dest
		}
		dest = filepath.Clean(dest)
		nsArgs = append(nsArgs, option, fmt.Sprintf("%s:%s", m.host, dest))
	}
	mappedArgv := make([]string, len(argv))
	for i, a := range argv {
		m := mapHostPath(a)
		if m == "" {
			m = a
		}
		mappedArgv[i] = m
	}
	binPath := rr.ChrootBin
	if binPath == "" {
		binPath = chrootRunPath
	}
	cmdArgs := []string{binPath, chrootDest, insideWorkdir, "--"}
	cmdArgs = append(cmdArgs, mappedArgv...)

	nsArgs = append(nsArgs, "--")
	nsArgs = append(nsArgs, cmdArgs...)

	cmd := exec.Command(nsjailPath, nsArgs...)
	env := os.Environ()
	env = append(env,
		pathEnv,
		"HOME=/tmp",
		"TMPDIR=/tmp",
		"LANG=C.UTF-8",
		"LD=/usr/bin/x86_64-linux-gnu-ld",
	)
	if debugDirsEnv != "" {
		env = append(env, "CHROOT_RUN_DEBUG_DIRS="+debugDirsEnv)
	}
	if debugDepthEnv != "" {
		env = append(env, "CHROOT_RUN_DEBUG_DEPTH="+debugDepthEnv)
	}
	if term := strings.TrimSpace(os.Getenv("TERM")); term != "" {
		env = append(env, "TERM="+term)
	}
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runErr := cmd.Run()
	if runErr != nil {
		log.Printf("sandbox: interactive nsjail/chroot-run failed: %v (cmd=%s %s)", runErr, nsjailPath, strings.Join(nsArgs, " "))
	}
	return runErr
}

// RLimits defines rlimit values applied inside the sandbox.
type RLimits struct {
	CPUSeconds  int
	ASBytes     int
	FSizeBytes  int
	NProc       int
	NOFile      int
	OutputLimit int
}

// RunOnHost executes argv on the host namespace (no chroot) applying rlimits.
func RunOnHost(ctx context.Context, workdir string, argv []string, stdin string, lim RLimits) (string, error) {
	if len(argv) == 0 {
		return "", errors.New("no argv provided")
	}
	pr := []string{"/usr/bin/prlimit"}
	if lim.CPUSeconds > 0 {
		pr = append(pr, fmt.Sprintf("--cpu=%d", lim.CPUSeconds))
	}
	if lim.ASBytes > 0 {
		pr = append(pr, fmt.Sprintf("--as=%d", lim.ASBytes))
	} else if lim.ASBytes < 0 {
		pr = append(pr, "--as=unlimited")
	}
	if lim.FSizeBytes > 0 {
		pr = append(pr, fmt.Sprintf("--fsize=%d", lim.FSizeBytes))
	}
	if lim.NProc > 0 {
		pr = append(pr, fmt.Sprintf("--nproc=%d", lim.NProc))
	}
	if lim.NOFile > 0 {
		pr = append(pr, fmt.Sprintf("--nofile=%d", lim.NOFile))
	}
	pr = append(pr, "--")
	pr = append(pr, argv...)
	cmd := exec.CommandContext(ctx, pr[0], pr[1:]...)
	if workdir == "" {
		workdir = "/"
	}
	cmd.Dir = workdir
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"HOME=/tmp",
		"TMPDIR=/tmp",
		"LANG=C.UTF-8",
	}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	stdoutBuf := &safeCapBuffer{max: lim.OutputLimit}
	stderrBuf := &safeCapBuffer{max: lim.OutputLimit}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	err := cmd.Run()
	outStr := stdoutBuf.String()
	if err != nil {
		if errStr := strings.TrimSpace(stderrBuf.String()); errStr != "" {
			if outStr != "" && !strings.HasSuffix(outStr, "\n") {
				outStr += "\n"
			}
			outStr += errStr
		}
	}
	return outStr, err
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// safeCapBuffer stores up to max bytes, discarding the rest, concurrency-safe.
type safeCapBuffer struct {
	mu   sync.Mutex
	buf  []byte
	max  int
	over bool
}

func (b *safeCapBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.max <= 0 {
		b.buf = append(b.buf, p...)
		return len(p), nil
	}
	remain := b.max - len(b.buf)
	if remain > 0 {
		if len(p) <= remain {
			b.buf = append(b.buf, p...)
		} else {
			b.buf = append(b.buf, p[:remain]...)
			b.over = true
		}
	} else {
		b.over = true
	}
	return len(p), nil
}

func (b *safeCapBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}

// CopyFile copies a regular file, creating parent directories as needed.
func CopyFile(src, dst string, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		if os.IsPermission(err) {
			if f, createErr := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm); createErr == nil {
				_ = f.Close()
				return nil
			}
		}
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// ResetChrootTmp clears and recreates /tmp inside a runroot with sticky bit.
func ResetChrootTmp(rr *RunRoot) error {
	if rr == nil {
		return errors.New("runroot not prepared")
	}
	if err := clearDirectory(rr.TmpHost); err != nil {
		return err
	}
	return os.Chmod(rr.TmpHost, 0o1777)
}

func ensureDirWithPerm(path string, perm fs.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return err
	}
	return os.Chmod(path, perm)
}

func ensurePlaceholder(path string, isDir bool, perm fs.FileMode) error {
	if isDir {
		return ensureDirWithPerm(path, perm)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, perm)
		if err != nil {
			return err
		}
		_ = f.Close()
	}
	return nil
}

func clearDirectory(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ensureDirWithPerm(path, 0o1777)
		}
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
