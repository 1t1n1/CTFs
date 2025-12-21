package main

import (
	"database/sql"
	_ "embed"
	"errors"
	"gopkg.in/yaml.v3"
	"log"
	"strings"
)

//go:embed challenges.yaml
var embeddedChallengeData []byte

type seedTest struct {
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
	Sample bool   `yaml:"sample"`
}

type seedChallenge struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Points      int        `yaml:"points"`
	Tests       []seedTest `yaml:"tests"`
}

func parseSeedChallenges(data []byte) ([]seedChallenge, error) {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil, errors.New("seed data is empty")
	}
	var items []seedChallenge
	if err := yaml.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func seedInitialChallenges() {
	items, err := parseSeedChallenges(embeddedChallengeData)
	if err != nil {
		log.Fatalf("runner: failed to load embedded challenges: %v", err)
	}
	tx, err := rdb.Begin()
	if err != nil {
		log.Fatalf("runner: begin seed transaction failed: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, ch := range items {
		if err := seedOneChallenge(tx, ch); err != nil {
			log.Fatalf("runner: seed challenge %s failed: %v", ch.Name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatalf("runner: commit seed transaction failed: %v", err)
	}
	log.Printf("runner: loaded %d built-in challenges", len(items))
}

func seedOneChallenge(tx *sql.Tx, ch seedChallenge) error {
	name := strings.TrimSpace(ch.Name)
	if name == "" {
		return errors.New("challenge name is empty")
	}
	desc := strings.TrimSpace(ch.Description)
	if desc == "" {
		return errors.New("challenge description is empty")
	}
	points := ch.Points
	if points <= 0 {
		points = 100
	}
	const upsertChallenge = `
INSERT INTO challenges(name, description, input, output, points, is_public)
VALUES($1,$2,'','',$3,TRUE)
ON CONFLICT(name) DO UPDATE SET
  description=EXCLUDED.description,
  points=EXCLUDED.points,
  is_public=EXCLUDED.is_public;`
	if _, err := tx.Exec(upsertChallenge, name, desc, points); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM sample_cases WHERE challenge=$1`, name); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM judge_cases WHERE challenge=$1`, name); err != nil {
		return err
	}
	var samples []seedTest
	var judges []seedTest
	for _, t := range ch.Tests {
		in := strings.TrimRight(t.Input, "\r\n")
		out := strings.TrimRight(t.Output, "\r\n")
		if strings.TrimSpace(in) == "" && strings.TrimSpace(out) == "" {
			continue
		}
		entry := seedTest{Input: in, Output: out, Sample: t.Sample}
		if t.Sample {
			samples = append(samples, entry)
		} else {
			judges = append(judges, entry)
		}
	}
	if len(samples) == 0 && len(ch.Tests) > 0 {
		first := ch.Tests[0]
		samples = append(samples, seedTest{
			Input:  strings.TrimRight(first.Input, "\r\n"),
			Output: strings.TrimRight(first.Output, "\r\n"),
			Sample: true,
		})
	}
	if len(judges) == 0 {
		for _, t := range ch.Tests {
			judges = append(judges, seedTest{
				Input:  strings.TrimRight(t.Input, "\r\n"),
				Output: strings.TrimRight(t.Output, "\r\n"),
			})
		}
	}
	for idx, t := range samples {
		if _, err := tx.Exec(`INSERT INTO sample_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, name, idx, t.Input, t.Output); err != nil {
			return err
		}
	}
	for idx, t := range judges {
		if _, err := tx.Exec(`INSERT INTO judge_cases(challenge, idx, input, output) VALUES($1,$2,$3,$4)`, name, idx, t.Input, t.Output); err != nil {
			return err
		}
	}
	return nil
}
