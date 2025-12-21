package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// RunnerRequest is sent to the sandbox runner service
type RunnerRequest struct {
	Language  string `json:"language"`
	Code      string `json:"code"`
	Input     string `json:"input,omitempty"`
	Want      string `json:"want,omitempty"`
	Challenge string `json:"challenge,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Sandbox   string `json:"sandbox,omitempty"`
}

// RunnerResponse is returned from the sandbox runner service
// RunnerResponse is returned from the sandbox runner service
type RunnerResponse struct {
	Result      string `json:"result"`
	Output      string `json:"output,omitempty"`
	DurationMs  int    `json:"duration_ms,omitempty"`
	FailedIndex int    `json:"failed_index,omitempty"`
	Expected    string `json:"expected,omitempty"`
}

// ChallengeMeta is fetched from runner for UI rendering
type ChallengeMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Samples     []struct {
		Input  string `json:"input"`
		Output string `json:"output"`
	} `json:"samples"`
}

func getRunnerHTTPClient() http.Client {
	httpTimeoutMs := 40000
	if v := os.Getenv("RUNNER_HTTP_TIMEOUT_MS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			httpTimeoutMs = n
		}
	}
	return http.Client{Timeout: time.Duration(httpTimeoutMs) * time.Millisecond}
}

func fetchChallengeMeta(name string) (ChallengeMeta, error) {
	client := getRunnerHTTPClient()
	endpoint := "http://runner:9000/challenge?name=" + url.QueryEscape(name)
	resp, err := client.Get(endpoint)
	if err != nil {
		return ChallengeMeta{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ChallengeMeta{}, fmt.Errorf("runner returned %d", resp.StatusCode)
	}
	var meta ChallengeMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return ChallengeMeta{}, err
	}
	return meta, nil
}

// executeSubmission sends the code and test case to the sandbox runner service
// executeSubmission sends code to the sandbox runner service and returns result and duration
func executeSubmission(id int64, challenge, language, code string) (string, int, int, string, string) {
	normalized, ok := normalizeLanguage(language)
	if !ok {
		log.Printf("[submission %d] Unsupported language: %s", id, language)
		return "Unsupported language", 0, -1, "", ""
	}
	// Runner HTTP timeout: default 40s, overridable by RUNNER_HTTP_TIMEOUT_MS
	httpTimeoutMs := 40000
	if v := os.Getenv("RUNNER_HTTP_TIMEOUT_MS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			httpTimeoutMs = n
		}
	}
	client := http.Client{Timeout: time.Duration(httpTimeoutMs) * time.Millisecond}
	url := "http://runner:9000/run"
	reqBody := RunnerRequest{
		Language:  normalized,
		Code:      code,
		Challenge: challenge,
		Mode:      "judge",
	}
	buf, _ := json.Marshal(reqBody)
	resp, err := client.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Printf("[submission %d] Runner request failed: %v", id, err)
		return "Runtime Error", 0, -1, "", ""
	}
	defer resp.Body.Close()
	var rr RunnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		log.Printf("[submission %d] Runner decode failed: %v", id, err)
		return "Runtime Error", 0, -1, "", ""
	}
	// Normalize failing index: only set for WA/RE/TLE; otherwise -1
	if rr.Result != "Wrong Answer" && rr.Result != "Runtime Error" && rr.Result != "Time Limit Exceeded" {
		rr.FailedIndex = -1
	}
	return rr.Result, rr.DurationMs, rr.FailedIndex, rr.Output, rr.Expected
}

// executeSample runs only sample tests and returns detailed failure info for UI testing
func executeSample(id int64, challenge, language, code string) (string, int, int, string, string) {
	normalized, ok := normalizeLanguage(language)
	if !ok {
		log.Printf("[test %d] Unsupported language: %s", id, language)
		return "Unsupported language", 0, -1, "", ""
	}
	httpTimeoutMs := 40000
	if v := os.Getenv("RUNNER_HTTP_TIMEOUT_MS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			httpTimeoutMs = n
		}
	}
	client := http.Client{Timeout: time.Duration(httpTimeoutMs) * time.Millisecond}
	url := "http://runner:9000/run"
	reqBody := RunnerRequest{
		Language:  normalized,
		Code:      code,
		Challenge: challenge,
		Mode:      "sample",
	}
	buf, _ := json.Marshal(reqBody)
	resp, err := client.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Printf("[test %d] Runner request failed: %v", id, err)
		return "Runtime Error", 0, -1, "", ""
	}
	defer resp.Body.Close()
	var rr RunnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		log.Printf("[test %d] Runner decode failed: %v", id, err)
		return "Runtime Error", 0, -1, "", ""
	}
	if rr.Result != "Wrong Answer" && rr.Result != "Runtime Error" && rr.Result != "Time Limit Exceeded" {
		rr.FailedIndex = -1
	}
	return rr.Result, rr.DurationMs, rr.FailedIndex, rr.Output, rr.Expected
}

func executeDebugRun(language, code, input, sandbox string) (RunnerResponse, error) {
	normalized, ok := normalizeLanguage(language)
	if !ok {
		return RunnerResponse{}, fmt.Errorf("unsupported language")
	}
	mode := strings.TrimSpace(sandbox)
	if mode == "" {
		mode = "default"
	}
	switch mode {
	case "default", "nsjail_only":
	default:
		return RunnerResponse{}, fmt.Errorf("unsupported sandbox mode")
	}
	httpTimeoutMs := 40000
	if v := os.Getenv("RUNNER_HTTP_TIMEOUT_MS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			httpTimeoutMs = n
		}
	}
	client := http.Client{Timeout: time.Duration(httpTimeoutMs) * time.Millisecond}
	url := "http://runner:9000/run"
	reqBody := RunnerRequest{
		Language: normalized,
		Code:     code,
		Input:    input,
		Sandbox:  mode,
	}
	buf, _ := json.Marshal(reqBody)
	resp, err := client.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return RunnerResponse{}, err
	}
	defer resp.Body.Close()
	var rr RunnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return RunnerResponse{}, err
	}
	if rr.Result != "Wrong Answer" && rr.Result != "Runtime Error" && rr.Result != "Time Limit Exceeded" {
		rr.FailedIndex = -1
	}
	return rr, nil
}
