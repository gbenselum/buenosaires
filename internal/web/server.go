// Package web provides a simple HTTP server for viewing script execution logs.
// It serves two main endpoints:
//   - / : Lists all available log files
//   - /logs/{filename} : Displays the contents of a specific log file
//
// The PatternFly stylesheet is embedded into the binary so the web interface
// works regardless of the current working directory or deployment artifact.
package web

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed assets/patternfly.min.css
var assetsFS embed.FS

// Server serves the web interface for a given log directory and repository.
type Server struct {
	logDir     string
	repoDir    string
	assetsDir  string
	httpServer *http.Server
}

// NewServer creates a Server that reads logs from logDir and asset metadata
// from the shell plugin's assets directory inside repoDir.
func NewServer(logDir, repoDir string) *Server {
	return &Server{
		logDir:    logDir,
		repoDir:   repoDir,
		assetsDir: filepath.Join(repoDir, "plugins", "shell", "assets"),
	}
}

// HTML templates for rendering the web interface.
const (
	listTemplate = `
<!DOCTYPE html>
<html class="pf-v5-c-page">
<head>
    <title>Logs</title>
    <link rel="stylesheet" href="/static/patternfly.min.css">
</head>
<body class="pf-v5-c-page__body">
    <div class="pf-v5-c-page__main" tabindex="-1">
        <section class="pf-v5-c-page__main-section pf-m-light">
            <div class="pf-v5-c-content">
                <h1>Log Files</h1>
            </div>
        </section>
        <section class="pf-v5-c-page__main-section">
            <div class="pf-v5-l-gallery pf-m-gutter">
                {{range .}}
                <div class="pf-v5-l-gallery__item">
                    <div class="pf-v5-c-card">
                        <div class="pf-v5-c-card__body">
                            <a href="/logs/{{.}}">{{.}}</a>
                        </div>
                    </div>
                </div>
                {{end}}
            </div>
        </section>
    </div>
</body>
</html>`

	viewTemplate = `
<!DOCTYPE html>
<html class="pf-v5-c-page">
<head>
    <title>View Log</title>
    <link rel="stylesheet" href="/static/patternfly.min.css">
</head>
<body class="pf-v5-c-page__body">
    <div class="pf-v5-c-page__main" tabindex="-1">
        <section class="pf-v5-c-page__main-section pf-m-light">
            <div class="pf-v5-c-content">
                <h1>Log: {{.Title}}</h1>
            </div>
        </section>
        <section class="pf-v5-c-page__main-section">
            <div class="pf-v5-c-card">
                <div class="pf-v5-c-card__body">
                    <pre>{{.Content}}</pre>
                </div>
            </div>
            <div class="pf-v5-c-accordion">
                <div class="pf-v5-c-accordion__toggle">
                    <span class="pf-v5-c-accordion__toggle-text">Asset JSON</span>
                </div>
                <div class="pf-v5-c-accordion__expanded-content">
                    <div class="pf-v5-c-accordion__expanded-content-body">
                        <pre id="asset-json"></pre>
                    </div>
                </div>
            </div>
        </section>
    </div>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            const assetUrl = '/plugins/shell/assets/' + encodeURIComponent('{{.ScriptBase}}') + '.json';
            fetch(assetUrl)
                .then(response => response.json())
                .then(data => {
                    document.getElementById('asset-json').textContent = JSON.stringify(data, null, 2);
                })
                .catch(error => {
                    console.error('Error fetching asset JSON:', error);
                    document.getElementById('asset-json').textContent = 'No asset metadata available.';
                });
        });
    </script>
</body>
</html>`
)

// StartServer starts the HTTP server on the specified address in a background
// goroutine. Errors are logged instead of terminating the process so a web
// server failure never kills the monitor.
// Parameters:
//   - addr: The address to listen on (e.g., ":8080")
//   - logDir: The directory containing log files
func StartServer(addr, logDir string) {
	srv := NewServer(logDir, ".")
	go func() {
		log.Printf("Starting web server on %s", addr)
		if err := srv.ListenAndServe(addr); err != nil && err != http.ErrServerClosed {
			log.Printf("Web server error: %v", err)
		}
	}()
}

// ListenAndServe starts the HTTP server synchronously.
func (s *Server) ListenAndServe(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s.httpServer.ListenAndServe()
}

// routes registers all HTTP handlers.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleList)
	mux.HandleFunc("GET /logs/{name}", s.handleView)
	mux.HandleFunc("GET /static/patternfly.min.css", s.handlePatternflyCSS)
	mux.HandleFunc("GET /plugins/shell/assets/{file}", s.handleAsset)
	return withSecurityHeaders(mux)
}

// withSecurityHeaders applies basic hardening headers to every response.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// handlePatternflyCSS serves the embedded PatternFly stylesheet.
func (s *Server) handlePatternflyCSS(w http.ResponseWriter, r *http.Request) {
	data, err := assetsFS.ReadFile("assets/patternfly.min.css")
	if err != nil {
		http.Error(w, "CSS asset unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(data)
}

// handleList handles requests to the root path and displays a list of all log files.
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	var logFiles []string
	files, err := os.ReadDir(s.logDir)
	if err != nil {
		// A missing log directory simply means nothing has run yet.
		if !os.IsNotExist(err) {
			http.Error(w, "Failed to read log directory", http.StatusInternalServerError)
			return
		}
	} else {
		// Filter for .log files only
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
				logFiles = append(logFiles, file.Name())
			}
		}
	}
	s.executeTemplate(w, listTemplate, logFiles)
}

// handleView handles requests to view a specific log file.
// It extracts the filename from the URL path and displays its contents.
func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	logName := r.PathValue("name")

	// Validation to prevent directory traversal: the name must be a single
	// path segment with no traversal components.
	if logName == "" || strings.Contains(logName, "..") || strings.Contains(logName, "/") || strings.Contains(logName, "\\") {
		http.Error(w, "Invalid log file name", http.StatusBadRequest)
		return
	}

	logPath := filepath.Join(s.logDir, logName)

	// Defense in depth: verify the resolved path stays inside the log
	// directory even though logName was already restricted to a single
	// path segment with no traversal components.
	logDirClean := filepath.Clean(s.logDir)
	if !strings.HasPrefix(logPath, logDirClean) {
		http.Error(w, "Invalid log file path", http.StatusBadRequest)
		return
	}

	// Check if the log file exists
	// #nosec G703 G304 -- path is a single sanitized segment joined onto the
	// configured log directory and verified to stay within it above.
	if _, err := os.Stat(logPath); err != nil {
		http.NotFound(w, r)
		return
	}

	// Read the log file contents
	// #nosec G703 G304 -- path is a single sanitized segment joined onto the
	// configured log directory and verified to stay within it above.
	content, err := os.ReadFile(logPath)
	if err != nil {
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title      string
		ScriptBase string
		Content    string
	}{
		Title:      logName,
		ScriptBase: strings.TrimSuffix(logName, ".log"),
		Content:    string(content),
	}
	s.executeTemplate(w, viewTemplate, data)
}

// handleAsset serves the asset JSON metadata for a script. The log filename
// (e.g. "deploy.sh.log") does not preserve the folder prefix that the asset
// path contains ("shell/deploy.sh.json"), so the asset directory is searched
// for a file whose base name matches the requested name.
func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if file == "" || strings.Contains(file, "..") || strings.Contains(file, "/") || strings.Contains(file, "\\") {
		http.Error(w, "Invalid asset name", http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(file, ".json") {
		http.Error(w, "Invalid asset file", http.StatusBadRequest)
		return
	}

	content, err := s.findAsset(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(content)
}

// findAsset locates an asset JSON file by base name in the assets directory,
// searching subdirectories to account for folder-prefixed script names.
func (s *Server) findAsset(name string) ([]byte, error) {
	var match string
	err := filepath.WalkDir(s.assetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == name {
			match = path
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if match == "" {
		return nil, os.ErrNotExist
	}
	// #nosec G304 -- match is produced by a directory walk of the plugin's
	// own assets folder and the requested name was sanitized above.
	return os.ReadFile(match)
}

// executeTemplate parses and executes the given template with the provided data.
func (s *Server) executeTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.New("page").Parse(tmpl)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}
