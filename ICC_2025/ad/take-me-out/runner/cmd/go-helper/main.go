package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"goexe-runner/internal/gohelper"
)

type helperPayload struct {
	Mode  string              `json:"mode"`
	Tests []gohelper.TestCase `json:"tests"`
}

func main() {
	log.SetFlags(0)

	mode := flag.String("mode", "", "execution mode")
	globalTimeout := flag.Int("global-timeout", 0, "global timeout in milliseconds")
	outputLimit := flag.Int("output-limit", 0, "output capture limit in bytes")
	codeFile := flag.String("code-file", "", "path to source code file")
	testsFile := flag.String("tests-file", "", "path to JSON tests file")
	sandboxEnv := flag.String("sandbox-env", "", "path to sandbox env directory")
	flag.Parse()

	if *codeFile == "" || *testsFile == "" {
		fatalJSON("missing --code-file or --tests-file")
	}

	code, err := os.ReadFile(*codeFile)
	if err != nil {
		fatalJSON(fmt.Sprintf("failed to read code file: %v", err))
	}

	testBytes, err := os.ReadFile(*testsFile)
	if err != nil {
		fatalJSON(fmt.Sprintf("failed to read tests file: %v", err))
	}

	var payload helperPayload
	if err := json.Unmarshal(testBytes, &payload); err != nil {
		fatalJSON(fmt.Sprintf("failed to decode tests: %v", err))
	}
	if len(payload.Tests) == 0 {
		fatalJSON("no tests provided")
	}

	req := gohelper.Request{
		Code:            string(code),
		Mode:            firstNonEmpty(*mode, payload.Mode),
		GlobalTimeoutMs: *globalTimeout,
		OutputLimit:     *outputLimit,
		SandboxEnv:      *sandboxEnv,
		Tests:           payload.Tests,
	}

	resp := gohelper.Execute(context.Background(), req)
	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		log.Fatalf("go helper: failed to encode response: %v", err)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func fatalJSON(msg string) {
	log.Printf("go helper: %s", msg)
	resp := gohelper.Response{Result: "Internal Error", Output: msg, FailedIndex: -1}
	_ = json.NewEncoder(os.Stdout).Encode(resp)
	os.Exit(1)
}
