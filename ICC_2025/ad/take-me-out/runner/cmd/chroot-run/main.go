package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Usage: chroot-run <newroot> <workdir> -- <cmd> [args...]
func main() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "usage: %s <newroot> <workdir> -- <cmd> [args...]\n", filepath.Base(os.Args[0]))
		os.Exit(97)
	}
	newroot := os.Args[1]
	workdir := os.Args[2]
	idx := 3
	for idx < len(os.Args) && os.Args[idx] != "--" {
		fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", os.Args[idx])
		os.Exit(98)
	}
	if idx >= len(os.Args) {
		fmt.Fprintln(os.Stderr, "missing command")
		os.Exit(100)
	}
	idx++ // skip "--"
	if err := os.Chdir(newroot); err != nil {
		fmt.Fprintf(os.Stderr, "chdir(newroot): %v\n", err)
		os.Exit(101)
	}
	if workdir == "" {
		workdir = "/"
	}
	workdir = filepath.Clean(workdir)
	if workdir == "." {
		workdir = "/"
	}
	if !strings.HasPrefix(workdir, "/") {
		workdir = "/" + workdir
	}
	relWork := strings.TrimPrefix(workdir, "/")
	if relWork != "" {
		if err := os.Chdir(relWork); err != nil {
			fmt.Fprintf(os.Stderr, "chdir(workdir=%s): %v\n", relWork, err)
			os.Exit(102)
		}
	}
	if err := syscall.Chroot(newroot); err != nil {
		fmt.Fprintf(os.Stderr, "chroot: %v\n", err)
		os.Exit(103)
	}
	if relWork == "" {
		if err := os.Chdir("/"); err != nil {
			fmt.Fprintf(os.Stderr, "chdir(/): %v\n", err)
			os.Exit(104)
		}
	}
	cmd := os.Args[idx]
	args := os.Args[idx:]
	if err := syscall.Exec(cmd, args, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "exec: %v\n", err)
		os.Exit(127)
	}
}
