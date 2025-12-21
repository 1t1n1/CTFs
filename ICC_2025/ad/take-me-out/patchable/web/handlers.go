package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type challengeTestYAML struct {
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
}

const sampleResultCacheTTL = 30 * time.Second

type cachedSampleResult struct {
	Result     string
	DurationMs int
	FailIdx    int
	Output     string
	Expect     string
	CachedAt   time.Time
}

var (
	sampleResultCache   = make(map[string]cachedSampleResult)
	sampleResultCacheMu sync.Mutex
)

func makeSampleResultCacheKey(userID, challengeID int, language, code string) string {
	hashed := sha256.Sum256([]byte(code))
	return fmt.Sprintf("%d:%d:%s:%s", userID, challengeID, language, hex.EncodeToString(hashed[:]))
}

func lookupSampleResultCache(key string) (cachedSampleResult, bool) {
	now := time.Now()
	sampleResultCacheMu.Lock()
	defer sampleResultCacheMu.Unlock()
	entry, ok := sampleResultCache[key]
	if !ok {
		return cachedSampleResult{}, false
	}
	if now.Sub(entry.CachedAt) > sampleResultCacheTTL {
		delete(sampleResultCache, key)
		return cachedSampleResult{}, false
	}
	return entry, true
}

func storeSampleResultCache(key string, entry cachedSampleResult) {
	now := time.Now()
	sampleResultCacheMu.Lock()
	defer sampleResultCacheMu.Unlock()
	for k, v := range sampleResultCache {
		if now.Sub(v.CachedAt) > sampleResultCacheTTL {
			delete(sampleResultCache, k)
		}
	}
	entry.CachedAt = now
	sampleResultCache[key] = entry
}

// parseChallengeTestsYAML converts YAML textarea input into a cleaned list of test specs
func parseChallengeTestsYAML(src string) ([]challengeTestYAML, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, nil
	}
	var tests []challengeTestYAML
	if err := yaml.Unmarshal([]byte(src), &tests); err != nil {
		return nil, err
	}
	cleaned := make([]challengeTestYAML, 0, len(tests))
	for _, t := range tests {
		in := strings.TrimSpace(t.Input)
		out := strings.TrimSpace(t.Output)
		if in == "" && out == "" {
			continue
		}
		cleaned = append(cleaned, challengeTestYAML{Input: t.Input, Output: t.Output})
	}
	return cleaned, nil
}

// challengeTestsToYAML serializes ordered tests back into YAML for prefilling forms
func challengeTestsToYAML(tests []TestCase) string {
	if len(tests) == 0 {
		return ""
	}
	arr := make([]challengeTestYAML, 0, len(tests))
	for _, t := range tests {
		arr = append(arr, challengeTestYAML{Input: t.Input, Output: t.Output})
	}
	buf, err := yaml.Marshal(arr)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(buf))
}

func normalizeLineEndings(src string) string {
	if !strings.Contains(src, "\r") {
		return src
	}
	src = strings.ReplaceAll(src, "\r\n", "\n")
	return strings.ReplaceAll(src, "\r", "\n")
}

var (
	errDuplicateChallenge = errors.New("duplicate challenge")
	errSampleCaseInsert   = errors.New("sample case insert failed")
	errHiddenCaseInsert   = errors.New("hidden case insert failed")
)

func challengeExists(name string) (int, bool, error) {
	id, err := getChallengeIDByName(name)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, err
}

func sanitizeChallengeTests(tests []challengeTestYAML) []challengeTestYAML {
	if len(tests) == 0 {
		return nil
	}
	cleaned := make([]challengeTestYAML, 0, len(tests))
	for _, t := range tests {
		in := strings.TrimSpace(t.Input)
		out := strings.TrimSpace(t.Output)
		if in == "" && out == "" {
			continue
		}
		cleaned = append(cleaned, challengeTestYAML{
			Input:  t.Input,
			Output: t.Output,
		})
	}
	return cleaned
}

func createChallengeRecord(userID int, name, description string, points int, publish bool, samples, hidden []challengeTestYAML) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var newChallengeID int
	if err := tx.QueryRow(
		`INSERT INTO challenges(name, description, created_by, input, output, points, is_public) VALUES($1,$2,$3,'','',$4,$5) RETURNING id`,
		name, description, userID, points, publish,
	).Scan(&newChallengeID); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return 0, errDuplicateChallenge
		}
		return 0, err
	}

	for i, t := range samples {
		if _, err := tx.Exec(`INSERT INTO sample_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, name, i, t.Input, t.Output); err != nil {
			return 0, fmt.Errorf("%w: %w", errSampleCaseInsert, err)
		}
	}
	for i, t := range hidden {
		if _, err := tx.Exec(`INSERT INTO judge_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, name, i, t.Input, t.Output); err != nil {
			return 0, fmt.Errorf("%w: %w", errHiddenCaseInsert, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newChallengeID, nil
}

type apiChallengeRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Points      int                 `json:"points"`
	SampleTests []challengeTestYAML `json:"sample_tests"`
	HiddenTests []challengeTestYAML `json:"hidden_tests"`
	IsPublic    bool                `json:"is_public"`
}

type apiChallengeResponse struct {
	ChallengeID int    `json:"challenge_id"`
	Name        string `json:"name"`
	IsPublic    bool   `json:"is_public"`
	DetailURL   string `json:"detail_url"`
	Status      string `json:"status"`
}

func getBasePageData(r *http.Request) BasePageData {
	return newBasePageData(getUser(r))
}

func renderNotFound(w http.ResponseWriter, r *http.Request, base BasePageData) {
	w.WriteHeader(http.StatusNotFound)
	if err := templates.ExecuteTemplate(w, "404.html", struct {
		BasePageData
		Path string
	}{
		BasePageData: base,
		Path:         r.URL.Path,
	}); err != nil {
		log.Printf("renderNotFound failed: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func summarizeDescription(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	preview := strings.Join(strings.Fields(raw), " ")
	runes := []rune(preview)
	const limit = 160
	if len(runes) > limit {
		preview = string(runes[:limit]) + "..."
	}
	return preview
}

func populateChallengePreviews(chals []ChallengeSummary) {
	for i := range chals {
		meta, err := fetchChallengeMeta(chals[i].Name)
		if err != nil {
			log.Printf("Failed to fetch preview for %s: %v", chals[i].Name, err)
			continue
		}
		summary := summarizeDescription(meta.Description)
		if summary == "" && len(meta.Samples) > 0 {
			summary = "Sample input: " + strings.TrimSpace(meta.Samples[0].Input)
		}
		chals[i].Preview = summary
	}
}

func splitTestsBySample(tests []TestCase) (samples []TestCase, hidden []TestCase) {
	for _, t := range tests {
		if t.IsSample {
			samples = append(samples, t)
		} else {
			hidden = append(hidden, t)
		}
	}
	return samples, hidden
}

// registerHandler handles user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "register.html", getBasePageData(r))
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	wantWriter := r.FormValue("is_writer") == "on"
	// Insert user into DB
	user, err := createUser(username, password, wantWriter)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}
	if err := setSession(w, user.Username); err != nil {
		log.Printf("failed to set session after registration: %v", err)
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// adminHandler shows the YAML upload form for admins
func adminHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil || !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "admin.html", newBasePageData(user))
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// adminUploadHandler processes uploaded YAML to add challenges (insert-only)
func adminUploadHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil || !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	file, _, err := r.FormFile("yaml")
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	var problems []struct {
		Name            string `yaml:"name"`
		Description     string `yaml:"description"`
		LegacyStatement string `yaml:"statement"`
		Input           string `yaml:"input"`
		Output          string `yaml:"output"`
		Points          int    `yaml:"points"`
		Tests           []struct {
			Input    string `yaml:"input"`
			Output   string `yaml:"output"`
			IsSample bool   `yaml:"sample"`
		} `yaml:"tests"`
	}
	if err := yaml.Unmarshal(data, &problems); err != nil {
		http.Error(w, "Invalid YAML", http.StatusBadRequest)
		return
	}
	for _, p := range problems {
		// Insert challenge (insert-only)
		pts := p.Points
		if pts == 0 {
			pts = 100
		}
		desc := strings.TrimSpace(p.Description)
		if desc == "" {
			desc = strings.TrimSpace(p.LegacyStatement)
		}
		if _, err := db.Exec(`INSERT INTO challenges(name, description, created_by, input, output, points, is_public) VALUES($1,$2,$3,'','',$4,FALSE) ON CONFLICT(name) DO NOTHING;`, p.Name, desc, user.ID, pts); err != nil {
			log.Printf("Failed to insert challenge %s: %v", p.Name, err)
		}
		// Insert tests (append-only)
		if len(p.Tests) > 0 {
			sampleIdx := 0
			hiddenIdx := 0
			for _, t := range p.Tests {
				if t.IsSample {
					if _, err := db.Exec(`INSERT INTO sample_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, p.Name, sampleIdx, t.Input, t.Output); err != nil {
						log.Printf("Failed to insert sample case for %s: %v", p.Name, err)
					}
					sampleIdx++
				} else {
					if _, err := db.Exec(`INSERT INTO judge_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, p.Name, hiddenIdx, t.Input, t.Output); err != nil {
						log.Printf("Failed to insert judge case for %s: %v", p.Name, err)
					}
					hiddenIdx++
				}
			}
		} else if p.Input != "" || p.Output != "" {
			if _, err := db.Exec(`INSERT INTO sample_cases(challenge, idx, input, output) VALUES($1,0,$2,$3)`, p.Name, p.Input, p.Output); err != nil {
				log.Printf("Failed to insert fallback sample case for %s: %v", p.Name, err)
			}
			if _, err := db.Exec(`INSERT INTO judge_cases(challenge, idx, input, output) VALUES($1,0,$2,$3)`, p.Name, p.Input, p.Output); err != nil {
				log.Printf("Failed to insert fallback judge case for %s: %v", p.Name, err)
			}
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func adminDebugHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil || !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	base := newBasePageData(user)
	languages := []string{"c", "go", "python", "ruby"}
	data := struct {
		BasePageData
		Language    string
		SandboxMode string
		Code        string
		Input       string
		Result      *RunnerResponse
		Error       string
		Languages   []string
	}{
		BasePageData: base,
		Language:     "python",
		SandboxMode:  "runner",
		Languages:    languages,
	}
	switch r.Method {
	case http.MethodGet:
		templates.ExecuteTemplate(w, "admin_debug.html", data)
		return
	case http.MethodPost:
		lang := strings.TrimSpace(r.FormValue("language"))
		mode := strings.TrimSpace(r.FormValue("sandbox_mode"))
		code := normalizeLineEndings(r.FormValue("code"))
		input := r.FormValue("input")
		adminPassword := strings.TrimSpace(r.FormValue("admin_password"))
		if mode == "" {
			mode = "runner"
		}
		data.Language = lang
		data.SandboxMode = mode
		data.Code = code
		data.Input = input
		if strings.TrimSpace(code) == "" {
			data.Error = "Code is required."
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}
		maxBytes := 131072
		if v := os.Getenv("MAX_CODE_BYTES"); v != "" {
			if n, e := strconv.Atoi(v); e == nil && n > 0 {
				maxBytes = n
			}
		}
		if len([]byte(code)) > maxBytes {
			data.Error = "Code size limit exceeded."
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}
		if adminPassword == "" {
			data.Error = "Admin password is required."
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}
		if adminPassword != user.Password {
			data.Error = "Admin password is incorrect."
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}
		proof, err := powProofFromForm(r)
		if err != nil {
			data.Error = "Proof-of-Work is required."
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}

		var sandbox string
		switch mode {
		case "runner":
			sandbox = "default"
		case "nsjail":
			sandbox = "nsjail_only"
		default:
			data.Error = "Unsupported sandbox mode."
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}

		challengeName := "admin_debug_runner"
		if mode == "nsjail" {
			challengeName = "admin_debug_nsjail"
		}
		if err := powMgr.Verify(user.ID, challengeName, proof, powPurposeAdmin); err != nil {
			_, msg := powErrorToHTTP(err)
			if msg == "" {
				msg = "Proof-of-Work validation failed."
			}
			data.Error = msg
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}

		res, err := executeDebugRun(lang, code, input, sandbox)
		if err != nil {
			data.Error = fmt.Sprintf("Runner request failed: %v", err)
			templates.ExecuteTemplate(w, "admin_debug.html", data)
			return
		}
		data.Result = &res
		templates.ExecuteTemplate(w, "admin_debug.html", data)
		return
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

// writerDashboardHandler shows entry points for writers to add challenges
func writerDashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !user.IsAdmin && !user.IsWriter {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	data := struct {
		BasePageData
	}{
		BasePageData: newBasePageData(user),
	}
	templates.ExecuteTemplate(w, "writer_dashboard.html", data)
}

// writerNewChallengeHandler renders and processes challenge creation for writers
func writerNewChallengeHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if !user.IsAdmin && !user.IsWriter {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	type writerChallengeForm struct {
		Name        string
		Description string
		Points      string
		SampleYAML  string
		HiddenYAML  string
		PublishNow  bool
	}
	form := writerChallengeForm{
		Points: "100",
	}

	data := struct {
		BasePageData
		Error         string
		Success       bool
		CreatedName   string
		CreatedID     int
		CreatedPublic bool
		Form          writerChallengeForm
	}{
		BasePageData: newBasePageData(user),
		Form:         form,
	}

	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Populate form with submitted values
	form.Name = strings.TrimSpace(r.FormValue("name"))
	descInput := strings.TrimSpace(r.FormValue("description"))
	if descInput == "" {
		descInput = strings.TrimSpace(r.FormValue("statement"))
	}
	form.Description = descInput
	form.Points = strings.TrimSpace(r.FormValue("points"))
	if form.Points == "" {
		form.Points = "100"
	}
	form.SampleYAML = strings.TrimSpace(r.FormValue("sample_tests"))
	form.HiddenYAML = strings.TrimSpace(r.FormValue("hidden_tests"))
	form.PublishNow = r.FormValue("is_public") == "on"
	data.Form = form

	if form.Name == "" {
		data.Error = "Please enter a challenge name."
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}
	if form.Description == "" {
		data.Error = "Please provide a challenge description."
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}

	points := 100
	if p, err := strconv.Atoi(form.Points); err != nil || p <= 0 {
		data.Error = "Points must be a positive integer."
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	} else {
		points = p
	}

	samples, err := parseChallengeTestsYAML(form.SampleYAML)
	if err != nil {
		data.Error = "Failed to parse sample tests YAML: " + err.Error()
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}
	if len(samples) == 0 {
		data.Error = "Provide at least one sample test case."
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}
	samples = sanitizeChallengeTests(samples)
	if len(samples) == 0 {
		data.Error = "Provide at least one sample test case."
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}

	hidden, err := parseChallengeTestsYAML(form.HiddenYAML)
	if err != nil {
		data.Error = "Failed to parse hidden tests YAML: " + err.Error()
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}
	hidden = sanitizeChallengeTests(hidden)

	if existingID, exists, err := challengeExists(form.Name); err != nil {
		data.Error = "Failed to verify existing challenges."
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	} else if exists {
		data.Error = "A challenge with that name already exists."
		data.CreatedID = existingID
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}

	publishNow := form.PublishNow
	newChallengeID, err := createChallengeRecord(user.ID, form.Name, form.Description, points, publishNow, samples, hidden)
	if err != nil {
		switch {
		case errors.Is(err, errDuplicateChallenge):
			data.Error = "A challenge with that name already exists."
			if existingID, exists, lookupErr := challengeExists(form.Name); lookupErr == nil && exists {
				data.CreatedID = existingID
			}
		case errors.Is(err, errSampleCaseInsert):
			data.Error = "Failed to add sample test case."
		case errors.Is(err, errHiddenCaseInsert):
			data.Error = "Failed to add hidden test case."
		default:
			data.Error = "Failed to finalize the challenge."
		}
		templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
		return
	}

	log.Printf("Writer %s created challenge %s", user.Username, form.Name)

	data.Success = true
	data.CreatedName = form.Name
	data.CreatedID = newChallengeID
	data.CreatedPublic = publishNow
	data.Form = writerChallengeForm{Points: "100"}

	responsePayload := struct {
		ChallengeID   int    `json:"challenge_id"`
		Name          string `json:"name"`
		IsPublic      bool   `json:"is_public"`
		DetailURL     string `json:"detail_url"`
		StatusMessage string `json:"status"`
	}{
		ChallengeID:   newChallengeID,
		Name:          form.Name,
		IsPublic:      publishNow,
		DetailURL:     fmt.Sprintf("/challenges/%d", newChallengeID),
		StatusMessage: "created",
	}

	w.Header().Set("X-Challenge-ID", strconv.Itoa(newChallengeID))
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responsePayload)
		return
	}

	templates.ExecuteTemplate(w, "writer_new_challenge.html", data)
	return
}

// apiChallengeHandler provides JSON endpoints to query and create challenges.
func apiChallengeHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		if name == "" {
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		id, exists, err := challengeExists(name)
		if err != nil {
			log.Printf("apiChallengeHandler lookup failed: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to lookup challenge")
			return
		}
		if !exists {
			writeJSONError(w, http.StatusNotFound, "challenge not found")
			return
		}
		detail, err := getChallengeForEdit(id)
		if err != nil {
			log.Printf("apiChallengeHandler detail failed: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to load challenge")
			return
		}
		resp := apiChallengeResponse{
			ChallengeID: id,
			Name:        detail.Name,
			IsPublic:    detail.IsPublic,
			DetailURL:   fmt.Sprintf("/challenges/%d", id),
			Status:      "existing",
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Challenge-ID", strconv.Itoa(id))
		json.NewEncoder(w).Encode(resp)
		return
	case http.MethodPost:
		if !user.IsAdmin && !user.IsWriter {
			writeJSONError(w, http.StatusForbidden, "forbidden")
			return
		}
		var req apiChallengeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Description = strings.TrimSpace(req.Description)
		points := req.Points
		if points <= 0 {
			points = 100
		}
		samples := sanitizeChallengeTests(req.SampleTests)
		hidden := sanitizeChallengeTests(req.HiddenTests)

		if req.Name == "" {
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		if req.Description == "" {
			writeJSONError(w, http.StatusBadRequest, "description is required")
			return
		}
		if len(samples) == 0 {
			writeJSONError(w, http.StatusBadRequest, "sample_tests must contain at least one case")
			return
		}
		if existingID, exists, err := challengeExists(req.Name); err != nil {
			log.Printf("apiChallengeHandler duplicate lookup failed: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to verify existing challenge")
			return
		} else if exists {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Challenge-ID", strconv.Itoa(existingID))
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]any{
				"error":        "duplicate",
				"challenge_id": existingID,
			})
			return
		}

		newID, err := createChallengeRecord(user.ID, req.Name, req.Description, points, req.IsPublic, samples, hidden)
		if err != nil {
			switch {
			case errors.Is(err, errDuplicateChallenge):
				existingID, exists, lookupErr := challengeExists(req.Name)
				if lookupErr != nil {
					log.Printf("apiChallengeHandler duplicate resolution failed: %v", lookupErr)
					writeJSONError(w, http.StatusInternalServerError, "failed to resolve duplicate challenge")
					return
				}
				if exists {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Challenge-ID", strconv.Itoa(existingID))
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]any{
						"error":        "duplicate",
						"challenge_id": existingID,
					})
					return
				}
				writeJSONError(w, http.StatusConflict, "duplicate challenge")
				return
			case errors.Is(err, errSampleCaseInsert):
				writeJSONError(w, http.StatusInternalServerError, "failed to add sample test case")
				return
			case errors.Is(err, errHiddenCaseInsert):
				writeJSONError(w, http.StatusInternalServerError, "failed to add hidden test case")
				return
			default:
				log.Printf("apiChallengeHandler create failed: %v", err)
				writeJSONError(w, http.StatusInternalServerError, "failed to create challenge")
				return
			}
		}

		log.Printf("Writer %s created challenge %s via API", user.Username, req.Name)
		resp := apiChallengeResponse{
			ChallengeID: newID,
			Name:        req.Name,
			IsPublic:    req.IsPublic,
			DetailURL:   fmt.Sprintf("/challenges/%d", newID),
			Status:      "created",
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Challenge-ID", strconv.Itoa(newID))
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
		return
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
}

// adminUsersHandler lists users for admin overview
func adminUsersHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil || !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	users, err := getAllUsersForAdmin()
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}
	data := struct {
		BasePageData
		Users []User
	}{
		BasePageData: newBasePageData(user),
		Users:        users,
	}
	templates.ExecuteTemplate(w, "admin_users.html", data)
}

// adminUserDetailHandler shows and updates a specific user's flags
func adminUserDetailHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil || !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	base := newBasePageData(user)
	idStr := strings.TrimPrefix(r.URL.Path, "/admin/users/")
	if idStr == "" {
		renderNotFound(w, r, base)
		return
	}
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		renderNotFound(w, r, base)
		return
	}

	target, err := getUserByID(userID)
	if err != nil {
		renderNotFound(w, r, base)
		return
	}

	data := struct {
		BasePageData
		Target  User
		Error   string
		Success bool
	}{
		BasePageData: base,
		Target:       *target,
	}

	switch r.Method {
	case http.MethodGet:
		templates.ExecuteTemplate(w, "admin_user_detail.html", data)
		return
	case http.MethodPost:
		wantWriter := r.FormValue("is_writer") == "on"
		if err := setUserWriterFlag(userID, wantWriter); err != nil {
			data.Error = "Failed to update permissions."
			templates.ExecuteTemplate(w, "admin_user_detail.html", data)
			return
		}
		target.IsWriter = wantWriter
		data.Target = *target
		data.Success = true
		templates.ExecuteTemplate(w, "admin_user_detail.html", data)
		return
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

// scoreboardHandler displays the aggregated scoreboard
func scoreboardHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	entries, err := getScoreboard()
	if err != nil {
		http.Error(w, "Failed to load scoreboard", http.StatusInternalServerError)
		return
	}
	data := struct {
		BasePageData
		Entries []ScoreEntry
	}{
		BasePageData: newBasePageData(user),
		Entries:      entries,
	}
	templates.ExecuteTemplate(w, "scoreboard.html", data)
}

// loginHandler handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "login.html", getBasePageData(r))
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	// Fetch user from DB
	user, err := getUserByUsername(username)
	if err != nil || user.Password != password {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := setSession(w, user.Username); err != nil {
		log.Printf("failed to set session after login: %v", err)
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// logoutHandler logs out the current user
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// indexHandler displays the list of challenges
func indexHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user != nil && user.IsWriter && !user.IsAdmin {
		http.Redirect(w, r, "/writer", http.StatusFound)
		return
	}
	base := newBasePageData(user)
	page := 1
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	isSearching := query != ""

	const perPage = 9
	var (
		challenges []ChallengeSummary
		total      int
		err        error
	)
	var totalPages int
	if isSearching {
		challenges, err = searchChallenges(query, 60)
		if err != nil {
			log.Printf("Failed to search challenges: %v", err)
			http.Error(w, "Failed to search challenges", http.StatusInternalServerError)
			return
		}
		total = len(challenges)
		totalPages = 1
		page = 1
	} else {
		challenges, total, err = getChallengeSummaries(page, perPage)
		if err != nil {
			log.Printf("Failed to load challenges: %v", err)
			http.Error(w, "Failed to load challenges", http.StatusInternalServerError)
			return
		}
		totalPages = 1
		if total == 0 {
			page = 1
		} else {
			totalPages = (total + perPage - 1) / perPage
			if totalPages < 1 {
				totalPages = 1
			}
			if page > totalPages {
				http.Redirect(w, r, "/?page="+strconv.Itoa(totalPages), http.StatusFound)
				return
			}
		}
	}
	if len(challenges) > 0 {
		populateChallengePreviews(challenges)
	}
	start := 0
	end := 0
	if !isSearching && total > 0 {
		start = (page-1)*perPage + 1
		end = start + len(challenges) - 1
	}
	hasPrev := !isSearching && page > 1
	hasNext := !isSearching && total > 0 && page < totalPages
	prevPage := page - 1
	if !hasPrev {
		prevPage = 1
	}
	nextPage := page + 1
	if !hasNext {
		nextPage = totalPages
	}
	visibleStart := page - 2
	if visibleStart < 1 {
		visibleStart = 1
	}
	visibleEnd := page + 2
	if visibleEnd > totalPages {
		visibleEnd = totalPages
	}
	pageNumbers := make([]int, 0, visibleEnd-visibleStart+1)
	if !isSearching {
		for i := visibleStart; i <= visibleEnd; i++ {
			pageNumbers = append(pageNumbers, i)
		}
	}
	showFirst := !isSearching && visibleStart > 1
	showLast := !isSearching && visibleEnd < totalPages
	data := struct {
		BasePageData
		Challenges  []ChallengeSummary
		Page        int
		TotalPages  int
		Total       int
		HasPrev     bool
		HasNext     bool
		PrevPage    int
		NextPage    int
		Start       int
		End         int
		PageNumbers []int
		ShowFirst   bool
		ShowLast    bool
		FirstPage   int
		LastPage    int
		IsSearching bool
		Query       string
		SearchTotal int
	}{
		BasePageData: base,
		Challenges:   challenges,
		Page:         page,
		TotalPages:   totalPages,
		Total:        total,
		HasPrev:      hasPrev,
		HasNext:      hasNext,
		PrevPage:     prevPage,
		NextPage:     nextPage,
		Start:        start,
		End:          end,
		PageNumbers:  pageNumbers,
		ShowFirst:    showFirst,
		ShowLast:     showLast,
		FirstPage:    1,
		LastPage:     totalPages,
		IsSearching:  isSearching,
		Query:        query,
		SearchTotal:  len(challenges),
	}
	templates.ExecuteTemplate(w, "index.html", data)
}

// usersHandler shows the community directory with optional search
func usersHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	base := newBasePageData(user)

	page := 1
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	isSearching := query != ""

	const perPage = 18
	var (
		users []User
		total int
		err   error
	)
	var totalPages int
	if isSearching {
		users, err = searchUsers(query, 100)
		if err != nil {
			log.Printf("Failed to search users: %v", err)
			http.Error(w, "Failed to search users", http.StatusInternalServerError)
			return
		}
		total = len(users)
		totalPages = 1
		page = 1
	} else {
		users, total, err = getUsersPaginated(page, perPage)
		if err != nil {
			log.Printf("Failed to load users: %v", err)
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}
		if total == 0 {
			totalPages = 1
			page = 1
		} else {
			totalPages = (total + perPage - 1) / perPage
			if totalPages < 1 {
				totalPages = 1
			}
			if page > totalPages {
				http.Redirect(w, r, "/users?page="+strconv.Itoa(totalPages), http.StatusFound)
				return
			}
		}
	}

	start := 0
	end := 0
	if !isSearching && total > 0 {
		start = (page-1)*perPage + 1
		end = start + len(users) - 1
	}

	hasPrev := !isSearching && page > 1
	hasNext := !isSearching && total > 0 && page < totalPages
	prevPage := page - 1
	if !hasPrev {
		prevPage = 1
	}
	nextPage := page + 1
	if !hasNext {
		nextPage = totalPages
	}
	visibleStart := page - 2
	if visibleStart < 1 {
		visibleStart = 1
	}
	visibleEnd := page + 2
	if visibleEnd > totalPages {
		visibleEnd = totalPages
	}
	pageNumbers := make([]int, 0, visibleEnd-visibleStart+1)
	if !isSearching {
		for i := visibleStart; i <= visibleEnd; i++ {
			pageNumbers = append(pageNumbers, i)
		}
	}
	showFirst := !isSearching && visibleStart > 1
	showLast := !isSearching && visibleEnd < totalPages

	data := struct {
		BasePageData
		Users       []User
		Page        int
		TotalPages  int
		Total       int
		HasPrev     bool
		HasNext     bool
		PrevPage    int
		NextPage    int
		Start       int
		End         int
		PageNumbers []int
		ShowFirst   bool
		ShowLast    bool
		FirstPage   int
		LastPage    int
		IsSearching bool
		Query       string
		SearchTotal int
	}{
		BasePageData: base,
		Users:        users,
		Page:         page,
		TotalPages:   totalPages,
		Total:        total,
		HasPrev:      hasPrev,
		HasNext:      hasNext,
		PrevPage:     prevPage,
		NextPage:     nextPage,
		Start:        start,
		End:          end,
		PageNumbers:  pageNumbers,
		ShowFirst:    showFirst,
		ShowLast:     showLast,
		FirstPage:    1,
		LastPage:     totalPages,
		IsSearching:  isSearching,
		Query:        query,
		SearchTotal:  len(users),
	}

	templates.ExecuteTemplate(w, "users.html", data)
}

// challengeHandler renders and manages challenge view/edit flows
func challengeHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	base := newBasePageData(user)
	path := strings.TrimPrefix(r.URL.Path, "/challenges/")
	path = strings.Trim(path, "/")
	if path == "" {
		renderNotFound(w, r, base)
		return
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		renderNotFound(w, r, base)
		return
	}
	idStr := parts[0]
	chalID, err := strconv.Atoi(idStr)
	if err != nil {
		renderNotFound(w, r, base)
		return
	}
	var action string
	if len(parts) > 1 {
		action = parts[1]
	}
	if len(parts) > 2 {
		renderNotFound(w, r, base)
		return
	}

	detail, err := getChallengeForEdit(chalID)
	if err != nil {
		renderNotFound(w, r, base)
		return
	}
	name := detail.Name
	meta, err := fetchChallengeMeta(name)
	if err != nil {
		renderNotFound(w, r, base)
		return
	}

	if action != "" && action != "update" && action != "publish" {
		renderNotFound(w, r, base)
		return
	}

	isOwner := detail.CreatedBy != nil && *detail.CreatedBy == user.ID
	canEdit := user.IsAdmin || isOwner
	canView := detail.IsPublic || canEdit
	if !canView {
		renderNotFound(w, r, base)
		return
	}
	canSubmit := !(user.IsWriter && !user.IsAdmin)

	type submissionRow struct {
		ID        int
		Username  string
		Language  string
		Result    string
		CreatedAt string
	}

	loadSubs := func() []submissionRow {
		var (
			records []struct {
				ID        int
				Username  string
				Language  string
				Result    string
				CreatedAt string
			}
			err error
		)
		records, err = getSubmissionsByChallenge(name)
		if err != nil {
			log.Printf("Failed to load submissions for challenge %s: %v", name, err)
			return nil
		}
		out := make([]submissionRow, 0, len(records))
		for _, rec := range records {
			out = append(out, submissionRow{
				ID:        rec.ID,
				Username:  rec.Username,
				Language:  rec.Language,
				Result:    rec.Result,
				CreatedAt: rec.CreatedAt,
			})
		}
		return out
	}

	sampleCases, err := getSampleCases(name)
	if err != nil {
		log.Printf("Failed to load sample cases for %s: %v", name, err)
		sampleCases = nil
	}
	type challengeEditForm struct {
		Description string
		Points      string
		SampleYAML  string
		HiddenYAML  string
	}

	defaultForm := challengeEditForm{
		Description: meta.Description,
		Points:      strconv.Itoa(detail.Points),
		SampleYAML:  challengeTestsToYAML(sampleCases),
	}

	render := func(form challengeEditForm, errMsg, successMsg string, previewTests []TestCase) {
		if form.Description == "" {
			form.Description = meta.Description
		}
		if form.Points == "" {
			form.Points = strconv.Itoa(detail.Points)
		}
		if form.SampleYAML == "" {
			form.SampleYAML = challengeTestsToYAML(sampleCases)
		}
		preview := TestCase{Description: form.Description}
		useTests := previewTests
		if len(useTests) == 0 {
			useTests = sampleCases
		}
		if len(useTests) > 0 {
			preview.Input = useTests[0].Input
			preview.Output = useTests[0].Output
		}
		data := struct {
			BasePageData
			ID          int
			Name        string
			IsPublic    bool
			TestCase    TestCase
			Submissions []submissionRow
			CanEdit     bool
			CanSubmit   bool
			EditForm    challengeEditForm
			Error       string
			Success     string
		}{
			BasePageData: base,
			ID:           detail.ID,
			Name:         name,
			IsPublic:     detail.IsPublic,
			TestCase:     preview,
			Submissions:  loadSubs(),
			CanEdit:      canEdit,
			CanSubmit:    canSubmit,
			EditForm:     form,
			Error:        errMsg,
			Success:      successMsg,
		}
		templates.ExecuteTemplate(w, "challenge.html", data)
	}

	switch {
	case r.Method == http.MethodGet && action == "":
		query := r.URL.Query()
		successMsg := ""
		if query.Get("published") == "1" {
			successMsg = "Challenge published."
		} else if query.Get("updated") == "1" {
			successMsg = "Challenge updated."
		}
		render(defaultForm, "", successMsg, sampleCases)
		return
	case r.Method == http.MethodPost && action == "update":
		if !canEdit {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		descInput := strings.TrimSpace(r.FormValue("description"))
		if descInput == "" {
			descInput = strings.TrimSpace(r.FormValue("statement"))
		}
		form := challengeEditForm{
			Description: descInput,
			Points:      strings.TrimSpace(r.FormValue("points")),
			SampleYAML:  strings.TrimSpace(r.FormValue("sample_tests")),
			HiddenYAML:  strings.TrimSpace(r.FormValue("hidden_tests")),
		}
		if form.Points == "" {
			form.Points = strconv.Itoa(detail.Points)
		}
		if form.Description == "" {
			render(form, "Please provide a challenge description.", "", nil)
			return
		}
		points, err := strconv.Atoi(form.Points)
		if err != nil || points <= 0 {
			render(form, "Points must be a positive integer.", "", nil)
			return
		}
		sampleSpecs, err := parseChallengeTestsYAML(form.SampleYAML)
		if err != nil {
			render(form, "Failed to parse sample tests YAML: "+err.Error(), "", nil)
			return
		}
		if len(sampleSpecs) == 0 {
			render(form, "Provide at least one sample test case.", "", nil)
			return
		}
		sampleTests := make([]TestCase, len(sampleSpecs))
		for i, t := range sampleSpecs {
			sampleTests[i] = TestCase{Input: t.Input, Output: t.Output, IsSample: true, Index: i}
		}
		hiddenProvided := form.HiddenYAML != ""
		var judgeTests []TestCase
		if hiddenProvided {
			hiddenSpecs, err := parseChallengeTestsYAML(form.HiddenYAML)
			if err != nil {
				render(form, "Failed to parse hidden tests YAML: "+err.Error(), "", sampleTests)
				return
			}
			judgeTests = make([]TestCase, len(hiddenSpecs))
			for i, t := range hiddenSpecs {
				judgeTests[i] = TestCase{Input: t.Input, Output: t.Output, Index: i}
			}
		}
		if err := updateChallengeWithTests(name, form.Description, points, sampleTests, judgeTests); err != nil {
			log.Printf("Failed to update challenge %s: %v", name, err)
			render(form, "Failed to update the challenge.", "", sampleTests)
			return
		}
		http.Redirect(w, r, "/challenges/"+strconv.Itoa(detail.ID)+"?updated=1", http.StatusSeeOther)
		return
	case r.Method == http.MethodPost && action == "publish":
		if !canEdit {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if detail.IsPublic {
			http.Redirect(w, r, "/challenges/"+strconv.Itoa(detail.ID), http.StatusSeeOther)
			return
		}
		if err := setChallengeVisibility(detail.ID, true); err != nil {
			log.Printf("Failed to publish challenge %s: %v", name, err)
			render(defaultForm, "Failed to publish the challenge.", "", sampleCases)
			return
		}
		http.Redirect(w, r, "/challenges/"+strconv.Itoa(detail.ID)+"?published=1", http.StatusSeeOther)
		return

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

// apiTestHandler runs sample tests without navigation and returns JSON
func apiTestHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if user.IsWriter && !user.IsAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	challengeIDStr := r.FormValue("challenge_id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	detail, err := getChallengeForEdit(challengeID)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	isOwner := detail.CreatedBy != nil && *detail.CreatedBy == user.ID
	if !detail.IsPublic && !user.IsAdmin && !isOwner {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	challenge := detail.Name
	language, ok := normalizeLanguage(r.FormValue("language"))
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"error": "unsupported language"})
		return
	}
	code := normalizeLineEndings(r.FormValue("code"))
	// Code size limit
	maxBytes := 131072
	if v := os.Getenv("MAX_CODE_BYTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxBytes = n
		}
	}
	if len([]byte(code)) > maxBytes {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(map[string]any{"error": "code too large"})
		return
	}

	proof, err := powProofFromForm(r)
	if err != nil {
		status, msg := powErrorToHTTP(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]any{"error": msg})
		return
	}
	if err := powMgr.Verify(user.ID, challenge, proof, powPurposeTest); err != nil {
		status, msg := powErrorToHTTP(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]any{"error": msg})
		return
	}

	id := time.Now().UnixNano()
	result, durationMs, failIdx, got, want := executeSample(id, challenge, language, code)
	cacheKey := makeSampleResultCacheKey(user.ID, challengeID, language, code)
	storeSampleResultCache(cacheKey, cachedSampleResult{
		Result:     result,
		DurationMs: durationMs,
		FailIdx:    failIdx,
		Output:     got,
		Expect:     want,
	})
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"result":       result,
		"duration_ms":  durationMs,
		"failed_index": failIdx,
	}
	if result != "Success" && strings.TrimSpace(got) != "" {
		resp["output"] = got
	}
	if failIdx >= 0 && result == "Wrong Answer" {
		resp["expected"] = want
	}
	json.NewEncoder(w).Encode(resp)
}

// submitHandler processes code submissions
func submitHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if user.IsWriter && !user.IsAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	challengeIDStr := r.FormValue("challenge_id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	detail, err := getChallengeForEdit(challengeID)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	isOwner := detail.CreatedBy != nil && *detail.CreatedBy == user.ID
	if !detail.IsPublic && !user.IsAdmin && !isOwner {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	challenge := detail.Name
	language, ok := normalizeLanguage(r.FormValue("language"))
	if !ok {
		http.Error(w, "Unsupported language", http.StatusBadRequest)
		return
	}
	code := normalizeLineEndings(r.FormValue("code"))
	// Code size limit
	maxBytes := 131072
	if v := os.Getenv("MAX_CODE_BYTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxBytes = n
		}
	}
	if len([]byte(code)) > maxBytes {
		http.Error(w, "Code too large", http.StatusRequestEntityTooLarge)
		return
	}

	proof, err := powProofFromForm(r)
	if err != nil {
		status, msg := powErrorToHTTP(err)
		http.Error(w, msg, status)
		return
	}
	if err := powMgr.Verify(user.ID, challenge, proof, powPurposeSubmission); err != nil {
		status, msg := powErrorToHTTP(err)
		http.Error(w, msg, status)
		return
	}

	// Enqueue: record as Pending (FIFO queue is DB-backed worker)
	sub := Submission{
		UserID:      user.ID,
		Challenge:   challenge,
		ChallengeID: challengeID,
		Language:    language,
		Code:        code,
		Result:      "Pending",
		DurationMs:  0,
		FailCaseIdx: -1,
		LastOutput:  "",
		ExpectedOut: "",
		CreatedAt:   time.Now(),
	}
	submissionID, err := createSubmission(sub)
	if err != nil {
		log.Printf("Failed to enqueue submission: %v", err)
		http.Error(w, "Failed to enqueue", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Submission-ID", strconv.Itoa(submissionID))

	sub.ID = submissionID
	responsePayload := buildSubmissionPayload(sub)
	responsePayload.PollingURL = fmt.Sprintf("/api/submissions/%d", submissionID)

	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(responsePayload)
		return
	}

	http.Redirect(w, r, "/submissions", http.StatusFound)
}

// testHandler allows users to verify code against sample tests without recording
func testHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if user.IsWriter && !user.IsAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	challengeIDStr := r.FormValue("challenge_id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	detail, err := getChallengeForEdit(challengeID)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	isOwner := detail.CreatedBy != nil && *detail.CreatedBy == user.ID
	if !detail.IsPublic && !user.IsAdmin && !isOwner {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	challenge := detail.Name
	language, ok := normalizeLanguage(r.FormValue("language"))
	if !ok {
		http.Error(w, "Unsupported language", http.StatusBadRequest)
		return
	}
	code := normalizeLineEndings(r.FormValue("code"))
	// Code size limit
	maxBytes := 131072
	if v := os.Getenv("MAX_CODE_BYTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxBytes = n
		}
	}
	if len([]byte(code)) > maxBytes {
		http.Error(w, "Code too large", http.StatusRequestEntityTooLarge)
		return
	}

	proof, err := powProofFromForm(r)
	if err != nil {
		status, msg := powErrorToHTTP(err)
		http.Error(w, msg, status)
		return
	}
	if err := powMgr.Verify(user.ID, challenge, proof, powPurposeTest); err != nil {
		status, msg := powErrorToHTTP(err)
		http.Error(w, msg, status)
		return
	}

	cacheKey := makeSampleResultCacheKey(user.ID, challengeID, language, code)
	if entry, ok := lookupSampleResultCache(cacheKey); ok {
		data := struct {
			Result      string
			Challenge   string
			ChallengeID int
			FailedIndex int
			Output      string
			Got         string
			Want        string
			DurationMs  int
		}{
			Result:      entry.Result,
			Challenge:   challenge,
			ChallengeID: challengeID,
			FailedIndex: entry.FailIdx,
			Output:      entry.Output,
			Got:         entry.Output,
			Want:        entry.Expect,
			DurationMs:  entry.DurationMs,
		}
		templates.ExecuteTemplate(w, "test.html", data)
		return
	}

	// Use timestamp for sandbox ID
	id := time.Now().UnixNano()

	// For test, run only sample tests and show details
	result, durationMs, failIdx, output, want := executeSample(id, challenge, language, code)
	data := struct {
		Result      string
		Challenge   string
		ChallengeID int
		FailedIndex int
		Output      string
		Got         string
		Want        string
		DurationMs  int
	}{
		Result:      result,
		Challenge:   challenge,
		ChallengeID: challengeID,
		FailedIndex: failIdx,
		Output:      output,
		Got:         output,
		Want:        want,
		DurationMs:  durationMs,
	}
	storeSampleResultCache(cacheKey, cachedSampleResult{
		Result:     result,
		DurationMs: durationMs,
		FailIdx:    failIdx,
		Output:     output,
		Expect:     want,
	})
	templates.ExecuteTemplate(w, "test.html", data)
}

// submissionsHandler shows past submissions for a user
func submissionsHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// Load submissions from DB
	subs, err := getSubmissionsByUser(user.ID)
	if err != nil {
		http.Error(w, "Failed to load submissions", http.StatusInternalServerError)
		return
	}
	var userSubs []struct {
		ID          int
		Challenge   string
		ChallengeID int
		Language    string
		Result      string
		Code        template.HTML
		CreatedAt   string
	}
	for _, s := range subs {
		userSubs = append(userSubs, struct {
			ID          int
			Challenge   string
			ChallengeID int
			Language    string
			Result      string
			Code        template.HTML
			CreatedAt   string
		}{
			ID:          s.ID,
			Challenge:   s.Challenge,
			ChallengeID: s.ChallengeID,
			Language:    s.Language,
			Result:      s.Result,
			Code:        template.HTML(s.Code),
			CreatedAt:   s.CreatedAt.Format(time.RFC1123),
		})
	}
	data := struct {
		BasePageData
		Submissions []struct {
			ID          int
			Challenge   string
			ChallengeID int
			Language    string
			Result      string
			Code        template.HTML
			CreatedAt   string
		}
	}{
		BasePageData: newBasePageData(user),
		Submissions:  userSubs,
	}
	templates.ExecuteTemplate(w, "submissions.html", data)
}

// submissionDetailHandler shows details of a specific submission
func submissionDetailHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	base := newBasePageData(user)
	// parse submission ID from URL
	idStr := r.URL.Path[len("/submission/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		renderNotFound(w, r, base)
		return
	}
	// fetch submission detail
	detail, err := getSubmissionDetail(id)
	if err != nil {
		renderNotFound(w, r, base)
		return
	}
	data := struct {
		BasePageData
		ID          int
		Username    string
		Challenge   string
		ChallengeID int
		Language    string
		Code        string
		Result      string
		CreatedAt   string
		DurationMs  int
		FailedCase  int
		Got         string
		Want        string
	}{
		BasePageData: base,
		ID:           detail.ID,
		Username:     detail.Username,
		Challenge:    detail.Challenge,
		ChallengeID:  detail.ChallengeID,
		Language:     detail.Language,
		Code:         detail.Code,
		Result:       detail.Result,
		CreatedAt:    detail.CreatedAt,
		DurationMs:   detail.DurationMs,
		FailedCase:   detail.FailedCase,
		Got:          detail.Got,
		Want:         detail.Want,
	}
	templates.ExecuteTemplate(w, "submission.html", data)
}

const (
	defaultAPISubmissionWait  = 60 * time.Second
	apiSubmissionPollInterval = 500 * time.Millisecond
)

type submissionStatusPayload struct {
	SubmissionID   int    `json:"submission_id"`
	Challenge      string `json:"challenge"`
	ChallengeID    int    `json:"challenge_id,omitempty"`
	Language       string `json:"language"`
	Result         string `json:"result"`
	DurationMs     int    `json:"duration_ms"`
	FailCaseIndex  int    `json:"fail_case_index"`
	Output         string `json:"output,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	CreatedAt      string `json:"created_at"`
	PollingURL     string `json:"polling_url,omitempty"`
	Mode           string `json:"mode"`
	TimedOut       bool   `json:"timed_out,omitempty"`
}

type apiPowRequest struct {
	ChallengeID *int   `json:"challenge_id,omitempty"`
	Challenge   string `json:"challenge,omitempty"`
	Purpose     string `json:"purpose,omitempty"`
}

type apiSubmissionRequest struct {
	ChallengeID *int     `json:"challenge_id,omitempty"`
	Challenge   string   `json:"challenge,omitempty"`
	Language    string   `json:"language"`
	Code        string   `json:"code"`
	Mode        string   `json:"mode,omitempty"`
	WaitMs      *int     `json:"wait_ms,omitempty"`
	Pow         powProof `json:"pow"`
}

type apiAdminDebugRequest struct {
	Language      string   `json:"language"`
	SandboxMode   string   `json:"sandbox_mode,omitempty"`
	Code          string   `json:"code"`
	Input         string   `json:"input,omitempty"`
	AdminPassword string   `json:"admin_password"`
	Pow           powProof `json:"pow"`
}

func resolveChallengeForUser(user *User, idPtr *int, name string) (*ChallengeDetail, int, string) {
	var challengeID int
	if idPtr != nil {
		challengeID = *idPtr
	} else {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, http.StatusBadRequest, "challenge_id or challenge is required"
		}
		id, err := getChallengeIDByName(name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, http.StatusNotFound, "challenge not found"
			}
			log.Printf("getChallengeIDByName failed: %v", err)
			return nil, http.StatusInternalServerError, "failed to resolve challenge"
		}
		challengeID = id
	}

	detail, err := getChallengeForEdit(challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.StatusNotFound, "challenge not found"
		}
		log.Printf("getChallengeForEdit failed: %v", err)
		return nil, http.StatusInternalServerError, "failed to load challenge"
	}
	isOwner := detail.CreatedBy != nil && *detail.CreatedBy == user.ID
	if !detail.IsPublic && !user.IsAdmin && !isOwner {
		return nil, http.StatusForbidden, "challenge is not accessible"
	}
	return detail, http.StatusOK, ""
}

func apiPowChallengeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user := getUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if user.IsWriter && !user.IsAdmin {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req apiPowRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	purpose := strings.TrimSpace(req.Purpose)
	if purpose == "" {
		purpose = powPurposeSubmission
	}
	switch purpose {
	case powPurposeSubmission, powPurposeTest, powPurposeAdmin:
	default:
		writeJSONError(w, http.StatusBadRequest, "unsupported purpose")
		return
	}

	var (
		challengeName string
		challengeID   int
	)

	if purpose == powPurposeAdmin {
		if !user.IsAdmin {
			writeJSONError(w, http.StatusForbidden, "forbidden")
			return
		}
		challengeName = strings.TrimSpace(req.Challenge)
		if challengeName == "" {
			challengeName = "admin_debug_runner"
		}
	} else {
		detail, status, message := resolveChallengeForUser(user, req.ChallengeID, req.Challenge)
		if status != http.StatusOK {
			writeJSONError(w, status, message)
			return
		}
		challengeID = detail.ID
		challengeName = detail.Name
	}

	challenge, err := powMgr.Issue(user.ID, challengeName, purpose)
	if err != nil {
		log.Printf("pow issue failed: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to issue proof-of-work")
		return
	}

	response := map[string]any{
		"pow":           challenge,
		"challenge":     challengeName,
		"difficulty":    challenge.Difficulty,
		"expires_at":    challenge.ExpiresAt,
		"purpose":       challenge.Purpose,
		"pow_signature": challenge.Signature,
	}
	if challengeID > 0 {
		response["challenge_id"] = challengeID
	}
	writeJSON(w, http.StatusOK, response)
}

func apiAdminDebugHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user := getUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !user.IsAdmin {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req apiAdminDebugRequest
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	code := normalizeLineEndings(req.Code)
	if strings.TrimSpace(code) == "" {
		writeJSONError(w, http.StatusBadRequest, "code is required")
		return
	}
	maxBytes := 131072
	if v := os.Getenv("MAX_CODE_BYTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxBytes = n
		}
	}
	if len([]byte(code)) > maxBytes {
		writeJSONError(w, http.StatusBadRequest, "code size limit exceeded")
		return
	}

	mode := strings.TrimSpace(req.SandboxMode)
	if mode == "" {
		mode = "runner"
	}
	var sandbox string
	switch mode {
	case "runner":
		sandbox = "default"
	case "nsjail":
		sandbox = "nsjail_only"
	default:
		writeJSONError(w, http.StatusBadRequest, "unsupported sandbox mode")
		return
	}

	lang := strings.TrimSpace(req.Language)
	normalized, ok := normalizeLanguage(lang)
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unsupported language")
		return
	}
	adminPassword := strings.TrimSpace(req.AdminPassword)
	if adminPassword == "" {
		writeJSONError(w, http.StatusBadRequest, "admin password is required")
		return
	}
	if adminPassword != user.Password {
		writeJSONError(w, http.StatusForbidden, "invalid admin password")
		return
	}

	challengeName := "admin_debug_runner"
	if mode == "nsjail" {
		challengeName = "admin_debug_nsjail"
	}

	if err := powMgr.Verify(user.ID, challengeName, req.Pow, powPurposeAdmin); err != nil {
		status, msg := powErrorToHTTP(err)
		writeJSONError(w, status, msg)
		return
	}

	res, err := executeDebugRun(normalized, code, req.Input, sandbox)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("runner request failed: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"result": res,
	})
}

// apiSubmissionCreateHandler enqueues a submission and optionally waits for completion.
func apiSubmissionCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user := getUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if user.IsWriter && !user.IsAdmin {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req apiSubmissionRequest
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	detail, status, message := resolveChallengeForUser(user, req.ChallengeID, req.Challenge)
	if status != http.StatusOK {
		writeJSONError(w, status, message)
		return
	}
	challengeID := detail.ID
	challengeName := detail.Name

	language, ok := normalizeLanguage(req.Language)
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "unsupported language")
		return
	}
	code := req.Code
	if strings.TrimSpace(code) == "" {
		writeJSONError(w, http.StatusBadRequest, "code is required")
		return
	}

	maxBytes := 131072
	if v := os.Getenv("MAX_CODE_BYTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxBytes = n
		}
	}
	if len([]byte(code)) > maxBytes {
		writeJSONError(w, http.StatusRequestEntityTooLarge, "code too large")
		return
	}

	if err := powMgr.Verify(user.ID, challengeName, req.Pow, powPurposeSubmission); err != nil {
		status, msg := powErrorToHTTP(err)
		writeJSONError(w, status, msg)
		return
	}

	now := time.Now()
	submission := Submission{
		UserID:      user.ID,
		Challenge:   challengeName,
		ChallengeID: challengeID,
		Language:    language,
		Code:        code,
		Result:      "Pending",
		DurationMs:  0,
		FailCaseIdx: -1,
		LastOutput:  "",
		ExpectedOut: "",
		CreatedAt:   now,
	}
	subID, err := createSubmission(submission)
	if err != nil {
		log.Printf("Failed to create submission: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to create submission")
		return
	}
	submission.ID = subID

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	shouldWait := false
	switch mode {
	case "":
		if req.WaitMs != nil && *req.WaitMs > 0 {
			shouldWait = true
			mode = "sync"
		} else {
			mode = "async"
		}
	case "sync":
		shouldWait = true
	case "async":
	default:
		writeJSONError(w, http.StatusBadRequest, "mode must be \"sync\" or \"async\"")
		return
	}

	waitDuration := defaultAPISubmissionWait
	if req.WaitMs != nil && *req.WaitMs > 0 {
		waitDuration = time.Duration(*req.WaitMs) * time.Millisecond
	}

	pollURL := fmt.Sprintf("/api/submissions/%d", subID)
	payload := buildSubmissionPayload(submission)
	payload.Mode = mode
	payload.PollingURL = pollURL

	statusCode := http.StatusAccepted
	if shouldWait {
		subStatus, completed, waitErr := waitForSubmissionResult(r.Context(), subID, apiSubmissionPollInterval, waitDuration)
		if waitErr != nil {
			if errors.Is(waitErr, context.Canceled) || errors.Is(waitErr, context.DeadlineExceeded) {
				writeJSONError(w, http.StatusRequestTimeout, "request canceled")
				return
			}
			log.Printf("waitForSubmissionResult failed: %v", waitErr)
			writeJSONError(w, http.StatusInternalServerError, "failed to retrieve submission result")
			return
		}
		payload = buildSubmissionPayload(subStatus)
		payload.Mode = "sync"
		payload.PollingURL = pollURL
		payload.TimedOut = !completed
		if completed {
			statusCode = http.StatusOK
		}
	}

	writeJSON(w, statusCode, payload)
}

// apiSubmissionDetailHandler returns the status of a specific submission.
func apiSubmissionDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user := getUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/submissions/")
	idStr = strings.Trim(idStr, "/")
	if idStr == "" {
		writeJSONError(w, http.StatusNotFound, "submission not found")
		return
	}
	subID, err := strconv.Atoi(idStr)
	if err != nil || subID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid submission id")
		return
	}

	subStatus, err := getSubmissionStatusByID(subID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "submission not found")
		} else {
			log.Printf("getSubmissionStatusByID failed: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to load submission")
		}
		return
	}

	if subStatus.UserID != user.ID && !user.IsAdmin {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	payload := buildSubmissionPayload(subStatus)
	payload.Mode = "async"
	payload.PollingURL = fmt.Sprintf("/api/submissions/%d", subStatus.ID)
	writeJSON(w, http.StatusOK, payload)
}

func powProofFromForm(r *http.Request) (powProof, error) {
	target := strings.TrimSpace(r.FormValue("pow_target"))
	nonce := strings.TrimSpace(r.FormValue("pow_nonce"))
	sig := strings.TrimSpace(r.FormValue("pow_signature"))
	purpose := strings.TrimSpace(r.FormValue("pow_purpose"))
	diffStr := strings.TrimSpace(r.FormValue("pow_difficulty"))
	expStr := strings.TrimSpace(r.FormValue("pow_expires_at"))

	if target == "" || nonce == "" || sig == "" || purpose == "" || diffStr == "" || expStr == "" {
		return powProof{}, errPowInvalid
	}

	diff, err := strconv.Atoi(diffStr)
	if err != nil {
		return powProof{}, errPowInvalid
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return powProof{}, errPowInvalid
	}

	return powProof{
		Target:     target,
		Difficulty: diff,
		ExpiresAt:  exp,
		Purpose:    purpose,
		Signature:  sig,
		Nonce:      nonce,
	}, nil
}

func powErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, errPowInvalid), errors.Is(err, errPowMismatch):
		return http.StatusBadRequest, "invalid proof-of-work"
	case errors.Is(err, errPowExpired):
		return http.StatusGone, "proof-of-work expired"
	case errors.Is(err, errPowDifficulty):
		return http.StatusBadRequest, "insufficient proof-of-work"
	case errors.Is(err, errPowReuse):
		return http.StatusTooManyRequests, "proof-of-work replay detected"
	case errors.Is(err, errPowConfiguration):
		return http.StatusInternalServerError, "proof-of-work temporarily unavailable"
	default:
		return http.StatusBadRequest, "invalid proof-of-work"
	}
}

func waitForSubmissionResult(ctx context.Context, submissionID int, pollInterval, maxWait time.Duration) (Submission, bool, error) {
	if pollInterval <= 0 {
		pollInterval = apiSubmissionPollInterval
	}
	sub, err := getSubmissionStatusByID(submissionID)
	if err != nil {
		return sub, false, err
	}
	if sub.Result != "Pending" && sub.Result != "Running" {
		return sub, true, nil
	}
	if maxWait <= 0 {
		return sub, false, nil
	}

	deadline := time.Now().Add(maxWait)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return sub, false, ctx.Err()
		case <-ticker.C:
		}
		sub, err = getSubmissionStatusByID(submissionID)
		if err != nil {
			return sub, false, err
		}
		if sub.Result != "Pending" && sub.Result != "Running" {
			return sub, true, nil
		}
		if time.Now().After(deadline) {
			return sub, false, nil
		}
	}
}

func buildSubmissionPayload(sub Submission) submissionStatusPayload {
	payload := submissionStatusPayload{
		SubmissionID:  sub.ID,
		Challenge:     sub.Challenge,
		Language:      sub.Language,
		Result:        sub.Result,
		DurationMs:    sub.DurationMs,
		FailCaseIndex: sub.FailCaseIdx,
		CreatedAt:     sub.CreatedAt.Format(time.RFC3339),
	}
	if sub.ChallengeID != 0 {
		payload.ChallengeID = sub.ChallengeID
	}
	if sub.LastOutput != "" {
		payload.Output = sub.LastOutput
	}
	if sub.ExpectedOut != "" {
		payload.ExpectedOutput = sub.ExpectedOut
	}
	return payload
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("writeJSON encode failed: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
