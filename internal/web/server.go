// Package web provides a simple HTTP server for viewing script execution logs.
// It serves two main endpoints:
//   - / : Lists all available log files
//   - /logs/{filename} : Displays the contents of a specific log file
package web

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// logDir stores the directory path where log files are located.
var (
	logDir string
)

// HTML templates for rendering the web interface.
const (
	listTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Logs</title>
</head>
<body>
    <h1>Log Files</h1>
    <ul>
        {{range .}}
        <li><a href="/logs/{{.}}">{{.}}</a></li>
        {{end}}
    </ul>
</body>
</html>`

	viewTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>View Log</title>
</head>
<body>
    <h1>Log: {{.Title}}</h1>
    <pre>{{.Content}}</pre>
</body>
</html>`
)

// StartServer starts the HTTP server on the specified address.
// It serves the log listing page and individual log file viewers.
// Parameters:
//   - addr: The address to listen on (e.g., ":8080")
//   - lDir: The directory containing log files
func StartServer(addr, lDir string) {
	logDir = lDir
	// Register HTTP handlers
	http.HandleFunc("/", handleList)
	http.HandleFunc("/logs/", handleView)

	log.Printf("Starting web server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

// handleList handles requests to the root path and displays a list of all log files.
func handleList(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		http.Error(w, "Failed to read log directory", http.StatusInternalServerError)
		return
	}

	// Filter for .log files only
	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
			logFiles = append(logFiles, file.Name())
		}
	}

	tmpl, err := template.New("list").Parse(listTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, logFiles); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}

// handleView handles requests to view a specific log file.
// It extracts the filename from the URL path and displays its contents.
func handleView(w http.ResponseWriter, r *http.Request) {
	logName := strings.TrimPrefix(r.URL.Path, "/logs/")
	
	// Additional validation to prevent directory traversal
	if strings.Contains(logName, "..") || strings.Contains(logName, "/") || strings.Contains(logName, "\\") {
		http.Error(w, "Invalid log file name", http.StatusBadRequest)
		return
	}
	
	logPath := filepath.Clean(filepath.Join(logDir, logName))

	// Sanitize the file path to prevent directory traversal attacks
	if !strings.HasPrefix(logPath, filepath.Clean(logDir)) {
		http.Error(w, "Invalid log file path", http.StatusBadRequest)
		return
	}

	// Check if the log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Read the log file contents
	content, err := os.ReadFile(logPath)
	if err != nil {
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title   string
		Content string
	}{
		Title:   logName,
		Content: string(content),
	}

	tmpl, err := template.New("view").Parse(viewTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}