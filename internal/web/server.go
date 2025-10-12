package web

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	logDir string
)

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

// StartServer starts the web server on the given address.
func StartServer(addr, lDir string) {
	logDir = lDir
	http.HandleFunc("/", handleList)
	http.HandleFunc("/logs/", handleView)

	log.Printf("Starting web server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func handleList(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		http.Error(w, "Failed to read log directory", http.StatusInternalServerError)
		return
	}

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

func handleView(w http.ResponseWriter, r *http.Request) {
	logName := strings.TrimPrefix(r.URL.Path, "/logs/")
	logPath := filepath.Join(logDir, logName)

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	content, err := ioutil.ReadFile(logPath)
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