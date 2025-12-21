package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB
var ErrUserExists = errors.New("user already exists")

// initDB connects to PostgreSQL and loads seed data required by the web app
func initDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "app_web"
	}
	flagPath := strings.TrimSpace(os.Getenv("DB_PASSWORD_FLAG_PATH"))
	if flagPath == "" {
		flagPath = "/flag1"
	}
	password, err := hashFileSHA256(flagPath)
	if err != nil {
		log.Fatalf("Failed to derive DB password from %s: %v", flagPath, err)
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "postgres"
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	// Connection pool configuration
	maxOpen := 20
	if v := os.Getenv("DB_MAX_OPEN_CONNS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxOpen = n
		}
	}
	maxIdle := 10
	if v := os.Getenv("DB_MAX_IDLE_CONNS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n >= 0 {
			maxIdle = n
		}
	}
	lifeMin := 15
	if v := os.Getenv("DB_CONN_MAX_LIFETIME_MINUTES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			lifeMin = n
		}
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(lifeMin) * time.Minute)
	// Wait for DB to be ready (up to ~10s)
	var pingErr error
	for i := 0; i < 10; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		log.Printf("Waiting for DB to be ready (%d/10): %v", i+1, pingErr)
		time.Sleep(1 * time.Second)
	}
	if pingErr != nil {
		log.Fatalf("DB ping failed after retries: %v", pingErr)
	}
}

func hashFileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:]), nil
}

// getChallengeSummaries fetches a paginated list of public challenges
func getChallengeSummaries(page, perPage int) ([]ChallengeSummary, int, error) {
	if perPage <= 0 {
		perPage = 12
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage
	rows, err := db.Query(`SELECT id, name, points
        FROM challenges WHERE is_public=TRUE ORDER BY name LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []ChallengeSummary
	for rows.Next() {
		var item ChallengeSummary
		if err := rows.Scan(&item.ID, &item.Name, &item.Points); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM challenges WHERE is_public=TRUE`).Scan(&total); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// searchChallenges finds public challenges matching the query by name or description
func searchChallenges(query string, limit int) ([]ChallengeSummary, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := db.Query(`SELECT id, name, points, description
        FROM challenges
        WHERE is_public = TRUE AND (name ILIKE $1 OR description ILIKE $1)
        ORDER BY LOWER(name)
        LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ChallengeSummary
	for rows.Next() {
		var item ChallengeSummary
		var desc sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.Points, &desc); err != nil {
			return nil, err
		}
		if desc.Valid {
			item.Preview = summarizeDescription(desc.String)
		}
		results = append(results, item)
	}
	return results, nil
}

// getTestCase retrieves a challenge by name
func getTestCase(name string) (TestCase, error) {
	// Return the first sample case along with the challenge description
	var tc TestCase
	desc, err := getChallengeDescription(name)
	if err != nil {
		return tc, err
	}
	samples, err := getSampleCases(name)
	if err != nil {
		return tc, err
	}
	if len(samples) > 0 {
		tc = samples[0]
	}
	tc.Description = desc
	return tc, nil
}

// getChallengeDescription returns the description for a challenge
func getChallengeDescription(name string) (string, error) {
	var desc string
	row := db.QueryRow(`SELECT description FROM challenges WHERE name=$1`, name)
	err := row.Scan(&desc)
	return desc, err
}

// getSampleCases returns public sample cases for a challenge
func getSampleCases(name string) ([]TestCase, error) {
	rows, err := db.Query(`SELECT idx, input, output FROM sample_cases WHERE challenge=$1 ORDER BY idx ASC`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tcs []TestCase
	for rows.Next() {
		var tc TestCase
		if err := rows.Scan(&tc.Index, &tc.Input, &tc.Output); err != nil {
			return nil, err
		}
		tc.IsSample = true
		tcs = append(tcs, tc)
	}
	return tcs, nil
}

// getChallengeForEdit gathers challenge ownership and numeric metadata by ID (description is fetched via runner)
func getChallengeForEdit(id int) (*ChallengeDetail, error) {
	row := db.QueryRow(`SELECT name, points, created_by, is_public FROM challenges WHERE id=$1`, id)
	var name string
	var points int
	var createdBy sql.NullInt64
	var isPublic bool
	if err := row.Scan(&name, &points, &createdBy, &isPublic); err != nil {
		return nil, err
	}
	var ownerPtr *int
	if createdBy.Valid {
		owner := int(createdBy.Int64)
		ownerPtr = &owner
	}
	return &ChallengeDetail{
		ID:        id,
		Name:      name,
		Points:    points,
		CreatedBy: ownerPtr,
		IsPublic:  isPublic,
	}, nil
}

// getChallengeNameByID returns the canonical name for a challenge ID
func getChallengeNameByID(id int) (string, error) {
	row := db.QueryRow(`SELECT name FROM challenges WHERE id=$1`, id)
	var name string
	if err := row.Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}

// getChallengeIDByName returns the numeric ID for a challenge name
func getChallengeIDByName(name string) (int, error) {
	row := db.QueryRow(`SELECT id FROM challenges WHERE name=$1`, name)
	var id int
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// setChallengeVisibility toggles whether a challenge is visible to players
func setChallengeVisibility(id int, public bool) error {
	_, err := db.Exec(`UPDATE challenges SET is_public=$1 WHERE id=$2`, public, id)
	return err
}

func replaceSampleCasesTx(tx *sql.Tx, name string, tests []TestCase) error {
	if _, err := tx.Exec(`DELETE FROM sample_cases WHERE challenge=$1`, name); err != nil {
		return err
	}
	for i, t := range tests {
		if _, err := tx.Exec(`INSERT INTO sample_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, name, i, t.Input, t.Output); err != nil {
			return err
		}
	}
	return nil
}

func replaceJudgeCasesTx(tx *sql.Tx, name string, tests []TestCase) error {
	if tests == nil {
		return nil
	}
	if _, err := tx.Exec(`SELECT purge_judge_cases($1)`, name); err != nil {
		return err
	}
	for i, t := range tests {
		if _, err := tx.Exec(`INSERT INTO judge_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, name, i, t.Input, t.Output); err != nil {
			return err
		}
	}
	return nil
}

// updateChallengeWithTests atomically updates challenge metadata and replaces its tests
func updateChallengeWithTests(name, description string, points int, sampleTests, judgeTests []TestCase) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE challenges SET description=$1, points=$2 WHERE name=$3`, description, points, name); err != nil {
		return err
	}
	if err := replaceSampleCasesTx(tx, name, sampleTests); err != nil {
		return err
	}
	if err := replaceJudgeCasesTx(tx, name, judgeTests); err != nil {
		return err
	}
	return tx.Commit()
}

// createUser inserts a new user
func createUser(username, password string, isWriter bool) (*User, error) {
	var existingID int
	err := db.QueryRow(`SELECT id FROM users WHERE username=$1`, username).Scan(&existingID)
	if err == nil {
		return nil, ErrUserExists
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	isAdmin := false
	row := db.QueryRow(`INSERT INTO users(username, password, is_admin, is_writer)
        VALUES($1,$2,$3,$4)
        RETURNING id, is_admin, is_writer`, username, password, isAdmin, isWriter)
	var u User
	if err := row.Scan(&u.ID, &u.IsAdmin, &u.IsWriter); err != nil {
		return nil, err
	}
	u.Username = username
	u.Password = password
	return &u, nil
}

// getUserByUsername fetches a user by username
func getUserByUsername(username string) (*User, error) {
	row := db.QueryRow(`SELECT id, username, password, is_admin, is_writer FROM users WHERE username=$1`, username)
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.IsAdmin, &u.IsWriter)
	return &u, err
}

// getUserByID fetches a user by ID (admin view)
func getUserByID(userID int) (*User, error) {
	row := db.QueryRow(`SELECT id, username, password, is_admin, is_writer FROM users WHERE id=$1`, userID)
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.IsAdmin, &u.IsWriter); err != nil {
		return nil, err
	}
	return &u, nil
}

// getAllUsersForAdmin lists users ordered by username (password omitted)
func getAllUsersForAdmin() ([]User, error) {
	rows, err := db.Query(`SELECT id, username, is_admin, is_writer FROM users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsAdmin, &u.IsWriter); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// getUsersPaginated returns a page of users ordered by username
func getUsersPaginated(page, perPage int) ([]User, int, error) {
	if perPage <= 0 {
		perPage = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage
	rows, err := db.Query(`SELECT id, username, is_admin, is_writer
        FROM users
        ORDER BY LOWER(username) ASC
        LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsAdmin, &u.IsWriter); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// searchUsers finds users whose username matches the query
func searchUsers(query string, limit int) ([]User, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := db.Query(`SELECT id, username, is_admin, is_writer
        FROM users
        WHERE username ILIKE $1
        ORDER BY LOWER(username)
        LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsAdmin, &u.IsWriter); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// setUserWriterFlag updates writer flag for a user
func setUserWriterFlag(userID int, isWriter bool) error {
	_, err := db.Exec(`UPDATE users SET is_writer = $1 WHERE id = $2`, isWriter, userID)
	return err
}

// createSubmission records a submission and returns its ID
func createSubmission(sub Submission) (int, error) {
	row := db.QueryRow(
		`INSERT INTO submissions(user_id, challenge, language, code, result, created_at, execution_time_ms, fail_case_index, last_output, expected_output)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING id`,
		sub.UserID, sub.Challenge, sub.Language, sub.Code, sub.Result, sub.CreatedAt, sub.DurationMs, sub.FailCaseIdx, sub.LastOutput, sub.ExpectedOut,
	)
	var id int
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// ensureSolve records a first-time solve for a user and challenge
func ensureSolve(userID int, challenge string, at time.Time) error {
	_, err := db.Exec(`INSERT INTO solves(user_id, challenge, created_at) VALUES($1,$2,$3) ON CONFLICT (user_id, challenge) DO NOTHING`, userID, challenge, at)
	return err
}

// ScoreEntry represents a row in the scoreboard
type ScoreEntry struct {
	Username string
	Total    int
	Solves   int
}

// getScoreboard aggregates scores per user
func getScoreboard() ([]ScoreEntry, error) {
	rows, err := db.Query(`
       SELECT u.username, COALESCE(SUM(c.points),0) AS total, COUNT(s.challenge) AS solves
       FROM users u
       LEFT JOIN solves s ON s.user_id = u.id
       LEFT JOIN challenges c ON c.name = s.challenge
       GROUP BY u.id, u.username
       ORDER BY total DESC, u.username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ScoreEntry
	for rows.Next() {
		var e ScoreEntry
		if err := rows.Scan(&e.Username, &e.Total, &e.Solves); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

// getSubmissionsByUser returns submissions for a user
func getSubmissionsByUser(userID int) ([]Submission, error) {
	rows, err := db.Query(
		`SELECT s.id, s.challenge, c.id, s.language, s.code, s.result, s.created_at
        FROM submissions s
        LEFT JOIN challenges c ON c.name = s.challenge
        WHERE s.user_id=$1 ORDER BY s.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subs []Submission
	for rows.Next() {
		var s Submission
		var chalID sql.NullInt64
		if err := rows.Scan(&s.ID, &s.Challenge, &chalID, &s.Language, &s.Code, &s.Result, &s.CreatedAt); err != nil {
			return nil, err
		}
		if chalID.Valid {
			s.ChallengeID = int(chalID.Int64)
		}
		s.UserID = userID
		subs = append(subs, s)
	}
	return subs, nil
}

// getSubmissionsByChallenge returns submissions for a specific challenge (with user)
func getSubmissionsByChallenge(challenge string) ([]struct {
	ID        int
	Username  string
	Language  string
	Result    string
	CreatedAt string
}, error) {
	rows, err := db.Query(
		`SELECT s.id, u.username, s.language, s.result, s.created_at
        FROM submissions s JOIN users u ON s.user_id = u.id
        WHERE s.challenge=$1 ORDER BY s.created_at DESC`, challenge)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subs []struct {
		ID        int
		Username  string
		Language  string
		Result    string
		CreatedAt string
	}
	for rows.Next() {
		var id int
		var username, language, result string
		var createdAt time.Time
		if err := rows.Scan(&id, &username, &language, &result, &createdAt); err != nil {
			return nil, err
		}
		subs = append(subs, struct {
			ID        int
			Username  string
			Language  string
			Result    string
			CreatedAt string
		}{id, username, language, result, createdAt.Format(time.RFC1123)})
	}
	return subs, nil
}

// getMySubmissionsByChallenge returns submissions for a specific challenge by a specific user
func getMySubmissionsByChallenge(userID int, challenge string) ([]struct {
	ID        int
	Username  string
	Language  string
	Result    string
	CreatedAt string
}, error) {
	rows, err := db.Query(
		`SELECT s.id, u.username, s.language, s.result, s.created_at
        FROM submissions s JOIN users u ON s.user_id = u.id
        WHERE s.challenge=$1 AND s.user_id=$2 ORDER BY s.created_at DESC`, challenge, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subs []struct {
		ID        int
		Username  string
		Language  string
		Result    string
		CreatedAt string
	}
	for rows.Next() {
		var id int
		var username, language, result string
		var createdAt time.Time
		if err := rows.Scan(&id, &username, &language, &result, &createdAt); err != nil {
			return nil, err
		}
		subs = append(subs, struct {
			ID        int
			Username  string
			Language  string
			Result    string
			CreatedAt string
		}{id, username, language, result, createdAt.Format(time.RFC1123)})
	}
	return subs, nil
}

// getSubmissionDetail fetches a single submission with user info

func getSubmissionDetail(subID int) (struct {
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
}, error) {
	var detail struct {
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
	}
	row := db.QueryRow(
		`SELECT s.id, u.username, s.challenge, c.id, s.language, s.code, s.result, s.created_at, s.execution_time_ms, s.fail_case_index, s.last_output, s.expected_output
        FROM submissions s JOIN users u ON s.user_id = u.id
        LEFT JOIN challenges c ON c.name = s.challenge
        WHERE s.id = $1`, subID)
	var createdAt time.Time
	var chalID sql.NullInt64
	if err := row.Scan(&detail.ID, &detail.Username, &detail.Challenge, &chalID, &detail.Language, &detail.Code, &detail.Result, &createdAt, &detail.DurationMs, &detail.FailedCase, &detail.Got, &detail.Want); err != nil {
		return detail, err
	}
	if chalID.Valid {
		detail.ChallengeID = int(chalID.Int64)
	}
	detail.CreatedAt = createdAt.Format(time.RFC1123)
	return detail, nil
}

// getSubmissionStatusByID returns submission fields for API polling
func getSubmissionStatusByID(subID int) (Submission, error) {
	var s Submission
	var createdAt time.Time
	var chalID sql.NullInt64
	row := db.QueryRow(
		`SELECT s.id, s.user_id, s.challenge, c.id, s.language, s.code, s.result, s.execution_time_ms, s.fail_case_index, s.last_output, s.expected_output, s.created_at
        FROM submissions s
        LEFT JOIN challenges c ON c.name = s.challenge
        WHERE s.id = $1`, subID)
	if err := row.Scan(&s.ID, &s.UserID, &s.Challenge, &chalID, &s.Language, &s.Code, &s.Result, &s.DurationMs, &s.FailCaseIdx, &s.LastOutput, &s.ExpectedOut, &createdAt); err != nil {
		return s, err
	}
	if chalID.Valid {
		s.ChallengeID = int(chalID.Int64)
	}
	s.CreatedAt = createdAt
	return s, nil
}
