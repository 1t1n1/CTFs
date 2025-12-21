package main

import (
	"html/template"
	"time"
)

// User represents a user account
type User struct {
	ID       int
	Username string
	Password string
	IsAdmin  bool
	IsWriter bool
}

// Submission holds a code submission
type Submission struct {
	ID          int
	UserID      int
	Challenge   string
	ChallengeID int
	Language    string
	Code        string
	Result      string
	DurationMs  int
	FailCaseIdx int
	LastOutput  string
	ExpectedOut string
	CreatedAt   time.Time
}

// TestCase holds input and output for a challenge
type TestCase struct {
	Description string
	Input       string
	Output      string
	IsSample    bool
	Index       int
}

// ChallengeDetail captures metadata and tests for a challenge when editing
type ChallengeDetail struct {
	ID        int
	Name      string
	Points    int
	CreatedBy *int
	IsPublic  bool
}

// ChallengeSummary is used when listing challenges
type ChallengeSummary struct {
	ID      int
	Name    string
	Points  int
	Preview string
}

// BasePageData carries the minimal user context required by layout fragments
type BasePageData struct {
	Username string
	IsAdmin  bool
	IsWriter bool
}

func newBasePageData(user *User) BasePageData {
	if user == nil {
		return BasePageData{}
	}
	return BasePageData{
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		IsWriter: user.IsAdmin || user.IsWriter,
	}
}

// templates holds parsed HTML templates
var templates *template.Template
