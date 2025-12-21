package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	loadDotEnv()
	if err := initAuth(); err != nil {
		log.Fatal(err)
	}
	initPowManager()
	os.MkdirAll("sandbox", 0755)
	templates = template.Must(template.New("").Option("missingkey=zero").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseGlob("templates/*.html"))
	// Initialize DB and load challenge data
	initDB()
	// Start FIFO submission workers (DB-backed)
	startSubmissionWorkersFromEnv()
	// Serve static assets
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Route handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/challenges/", challengeHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/test", testHandler)
	http.HandleFunc("/api/test", apiTestHandler)
	http.HandleFunc("/api/pow", apiPowChallengeHandler)
	http.HandleFunc("/api/admin/debug", apiAdminDebugHandler)
	http.HandleFunc("/api/challenges", apiChallengeHandler)
	http.HandleFunc("/api/submissions", apiSubmissionCreateHandler)
	http.HandleFunc("/api/submissions/", apiSubmissionDetailHandler)
	http.HandleFunc("/submissions", submissionsHandler)
	http.HandleFunc("/scoreboard", scoreboardHandler)
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/writer", writerDashboardHandler)
	http.HandleFunc("/writer/challenges/new", writerNewChallengeHandler)
	http.HandleFunc("/admin/users", adminUsersHandler)
	http.HandleFunc("/admin/users/", adminUserDetailHandler)
	http.HandleFunc("/admin/debug", adminDebugHandler)

	// Admin upload is disabled for now (hidden tests live only in runner)

	// Submission detail route
	http.HandleFunc("/submission/", submissionDetailHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
