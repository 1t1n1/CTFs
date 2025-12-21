package main

import (
	"database/sql"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

// startSubmissionWorkersFromEnv starts FIFO workers based on env var WORKER_CONCURRENCY
func startSubmissionWorkersFromEnv() {
	n := runtime.NumCPU()
	if v := os.Getenv("WORKER_CONCURRENCY"); v != "" {
		if x, e := strconv.Atoi(v); e == nil && x > 0 {
			n = x
		}
	} else {
		// default light parallelism
		if n > 2 {
			n = 2
		}
	}
	for i := 0; i < n; i++ {
		go submissionWorkerLoop(i)
	}
	log.Printf("Started %d submission worker(s)", n)
}

// submissionWorkerLoop claims oldest Pending submission and processes it via runner
func submissionWorkerLoop(workerID int) {
	for {
		job, ok, err := claimNextPending()
		if err != nil {
			log.Printf("[worker %d] claim error: %v", workerID, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if !ok {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		// Execute via runner
		id := time.Now().UnixNano()
		result, durationMs, failIdx, lastOut, expect := executeSubmission(id, job.Challenge, job.Language, job.Code)
		// Update DB
		if err := updateSubmissionAfterRun(job.ID, result, durationMs, failIdx, lastOut, expect); err != nil {
			log.Printf("[worker %d] update submission %d failed: %v", workerID, job.ID, err)
		}
		if result == "Success" {
			// best-effort solve record
			if err := ensureSolve(job.UserID, job.Challenge, time.Now()); err != nil {
				log.Printf("[worker %d] ensureSolve failed: %v", workerID, err)
			}
		}
	}
}

type pendingJob struct {
	ID        int
	UserID    int
	Challenge string
	Language  string
	Code      string
}

// claimNextPending atomically selects the oldest Pending submission and marks it Running
func claimNextPending() (pendingJob, bool, error) {
	var job pendingJob
	tx, err := db.Begin()
	if err != nil {
		return job, false, err
	}
	defer func() {
		// ensure rollback on early returns
		_ = tx.Rollback()
	}()

	row := tx.QueryRow(`
        SELECT id, user_id, challenge, language, code
        FROM submissions
        WHERE result = 'Pending'
        ORDER BY created_at ASC
        LIMIT 1
        FOR UPDATE SKIP LOCKED`)
	var id, uid int
	var chal, lang, code string
	if err := row.Scan(&id, &uid, &chal, &lang, &code); err != nil {
		if err == sql.ErrNoRows {
			return job, false, nil
		}
		return job, false, err
	}
	if _, err := tx.Exec(`UPDATE submissions SET result = 'Running' WHERE id = $1 AND result = 'Pending'`, id); err != nil {
		return job, false, err
	}
	if err := tx.Commit(); err != nil {
		return job, false, err
	}
	job = pendingJob{ID: id, UserID: uid, Challenge: chal, Language: lang, Code: code}
	return job, true, nil
}

// updateSubmissionAfterRun writes the final result and details
func updateSubmissionAfterRun(id int, result string, durationMs, failIdx int, lastOut, expect string) error {
	_, err := db.Exec(`
        UPDATE submissions
        SET result = $1,
            execution_time_ms = $2,
            fail_case_index = $3,
            last_output = $4,
            expected_output = $5
        WHERE id = $6
    `, result, durationMs, failIdx, lastOut, expect, id)
	return err
}
