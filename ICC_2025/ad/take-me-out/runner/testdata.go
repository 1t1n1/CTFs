package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"time"
)

type runnerTest struct {
	Input, Output string
	IsSample      bool
}

type challengeMeta struct {
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Samples     []struct{ Input, Output string } `json:"samples"`
}

var rdb *sql.DB

// init read-only DB connection for runner
func initRunnerDB() {
	host := getenv("RUNNER_DB_HOST", getenv("DB_HOST", "db"))
	port := getenv("RUNNER_DB_PORT", getenv("DB_PORT", "5432"))
	user := getenv("RUNNER_DB_USER", getenv("DB_USER", "app_runner"))
	pass := deriveRunnerDBPassword()
	name := getenv("RUNNER_DB_NAME", getenv("DB_NAME", "postgres"))
	ssl := getenv("DB_SSLMODE", "disable")
	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + pass + " dbname=" + name + " sslmode=" + ssl
	var err error
	rdb, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("runner DB open failed: %v", err)
	}
	// wait until DB is ready (similar to web)
	var pingErr error
	for i := 0; i < 10; i++ {
		pingErr = rdb.Ping()
		if pingErr == nil {
			break
		}
		log.Printf("runner: waiting for DB to be ready (%d/10): %v", i+1, pingErr)
		time.Sleep(1 * time.Second)
	}
	if pingErr != nil {
		log.Fatalf("runner DB ping failed after retries: %v", pingErr)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func deriveRunnerDBPassword() string {
	flagPath := strings.TrimSpace(os.Getenv("RUNNER_DB_PASSWORD_FLAG_PATH"))
	if flagPath == "" {
		flagPath = strings.TrimSpace(os.Getenv("DB_PASSWORD_FLAG_PATH"))
	}
	if flagPath == "" {
		flagPath = "/flag2"
	}
	data, err := os.ReadFile(flagPath)
	if err != nil {
		log.Fatalf("runner: failed to read password flag file %s: %v", flagPath, err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:])
}

func getRunnerTests(challenge, mode string) []runnerTest {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "sample" {
		rows, err := rdb.Query(`SELECT idx, input, output FROM sample_cases WHERE challenge=$1 ORDER BY idx ASC`, challenge)
		if err != nil {
			log.Printf("runner: query sample cases failed: %v", err)
			return nil
		}
		defer rows.Close()
		var tests []runnerTest
		for rows.Next() {
			var idx int
			var in, out string
			if err := rows.Scan(&idx, &in, &out); err != nil {
				log.Printf("runner: scan sample case failed: %v", err)
				return nil
			}
			tests = append(tests, runnerTest{Input: in, Output: out, IsSample: true})
		}
		if len(tests) == 0 {
			row := rdb.QueryRow(`SELECT input, output FROM judge_cases WHERE challenge=$1 ORDER BY idx ASC LIMIT 1`, challenge)
			var in, out string
			if err := row.Scan(&in, &out); err == nil {
				tests = append(tests, runnerTest{Input: in, Output: out, IsSample: true})
			}
		}
		return tests
	}

	// judge mode: run both sample (for sanity) and hidden cases
	var tests []runnerTest
	rows, err := rdb.Query(`SELECT idx, input, output FROM sample_cases WHERE challenge=$1 ORDER BY idx ASC`, challenge)
	if err != nil {
		log.Printf("runner: query sample cases failed: %v", err)
	} else {
		for rows.Next() {
			var idx int
			var in, out string
			if err := rows.Scan(&idx, &in, &out); err != nil {
				rows.Close()
				log.Printf("runner: scan sample case failed: %v", err)
				return nil
			}
			tests = append(tests, runnerTest{Input: in, Output: out, IsSample: true})
		}
		rows.Close()
	}
	rows, err = rdb.Query(`SELECT idx, input, output FROM judge_cases WHERE challenge=$1 ORDER BY idx ASC`, challenge)
	if err != nil {
		log.Printf("runner: query judge cases failed: %v", err)
	} else {
		for rows.Next() {
			var idx int
			var in, out string
			if err := rows.Scan(&idx, &in, &out); err != nil {
				rows.Close()
				log.Printf("runner: scan judge case failed: %v", err)
				return nil
			}
			tests = append(tests, runnerTest{Input: in, Output: out})
		}
		rows.Close()
	}
	return tests
}

func getChallengeMeta(name string) (challengeMeta, bool) {
	name = strings.TrimSpace(name)
	var desc string
	if err := rdb.QueryRow(`SELECT description FROM challenges WHERE name=$1`, name).Scan(&desc); err != nil {
		return challengeMeta{}, false
	}
	rows, err := rdb.Query(`SELECT input, output FROM sample_cases WHERE challenge=$1 ORDER BY idx ASC`, name)
	if err != nil {
		log.Printf("runner: query samples failed: %v", err)
		return challengeMeta{}, false
	}
	defer rows.Close()
	var samples []struct{ Input, Output string }
	for rows.Next() {
		var in, out string
		if err := rows.Scan(&in, &out); err != nil {
			return challengeMeta{}, false
		}
		samples = append(samples, struct{ Input, Output string }{in, out})
	}
	if len(samples) == 0 {
		// fallback: first hidden case
		row := rdb.QueryRow(`SELECT input, output FROM judge_cases WHERE challenge=$1 ORDER BY idx ASC LIMIT 1`, name)
		var in, out string
		if err := row.Scan(&in, &out); err == nil {
			samples = append(samples, struct{ Input, Output string }{in, out})
		}
	}
	return challengeMeta{Name: name, Description: desc, Samples: samples}, true
}
