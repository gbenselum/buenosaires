package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestServer builds a Server backed by temporary log and repo directories,
// pre-populated with two log files and one asset JSON file.
func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()

	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	logFile1 := filepath.Join(logDir, "script1.log")
	if err := os.WriteFile(logFile1, []byte("log content 1"), 0o644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}
	logFile2 := filepath.Join(logDir, "script2.log")
	if err := os.WriteFile(logFile2, []byte("log content 2"), 0o644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	// Create a nested asset file (assets are stored under the script's folder).
	assetDir := filepath.Join(tmpDir, "plugins", "shell", "assets", "shell")
	if err := os.MkdirAll(assetDir, 0o750); err != nil {
		t.Fatalf("failed to create asset dir: %v", err)
	}
	assetPath := filepath.Join(assetDir, "script1.sh.json")
	assetContent := `{"generation":1,"status":"success"}`
	if err := os.WriteFile(assetPath, []byte(assetContent), 0o644); err != nil {
		t.Fatalf("failed to write asset file: %v", err)
	}

	srv := NewServer(logDir, tmpDir)
	return srv, tmpDir
}

func doRequest(t *testing.T, srv *Server, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "http://example.com"+path, nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	return rec
}

func TestServer_ListLogs(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(t, srv, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("list handler returned wrong status code: got %v want %v", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, expected := range []string{`href="/logs/script1.log"`, `href="/logs/script2.log"`} {
		if !strings.Contains(body, expected) {
			t.Errorf("list handler body missing %q: got %s", expected, body)
		}
	}
}

func TestServer_ViewLog(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(t, srv, http.MethodGet, "/logs/script1.log")
	if rec.Code != http.StatusOK {
		t.Fatalf("view handler returned wrong status code: got %v want %v", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<pre>log content 1</pre>") {
		t.Errorf("view handler missing log content: got %s", body)
	}
	// The view must reference the asset endpoint and embedded stylesheet.
	if !strings.Contains(body, "/plugins/shell/assets/") || !strings.Contains(body, "encodeURIComponent('script1')") {
		t.Errorf("view handler missing asset URL: got %s", body)
	}
	if !strings.Contains(body, "/static/patternfly.min.css") {
		t.Errorf("view handler missing stylesheet link: got %s", body)
	}
}

func TestServer_ViewLogNotFound(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(t, srv, http.MethodGet, "/logs/missing.log")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("view handler returned wrong status code: got %v want %v", rec.Code, http.StatusNotFound)
	}
}

func TestServer_RejectsTraversal(t *testing.T) {
	srv, _ := newTestServer(t)

	// Directly invoke handleView with a crafted path value, since the router
	// normalizes ".." segments before they reach the handler.
	req := httptest.NewRequest(http.MethodGet, "http://example.com/logs/evil", nil)
	req.SetPathValue("name", "../etc/passwd")
	rec := httptest.NewRecorder()
	srv.handleView(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("traversal attempt should be rejected: got %v want %v", rec.Code, http.StatusBadRequest)
	}
}

func TestServer_ServesAssetJSON(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(t, srv, http.MethodGet, "/plugins/shell/assets/script1.sh.json")
	if rec.Code != http.StatusOK {
		t.Fatalf("asset handler returned wrong status code: got %v want %v", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("asset handler returned wrong content type: got %q", ct)
	}
	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"status":"success"`) {
		t.Errorf("asset handler returned unexpected body: got %s", body)
	}
}

func TestServer_AssetNotFound(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(t, srv, http.MethodGet, "/plugins/shell/assets/nope.json")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("asset handler returned wrong status code: got %v want %v", rec.Code, http.StatusNotFound)
	}
}

func TestServer_ServesPatternflyCSS(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(t, srv, http.MethodGet, "/static/patternfly.min.css")
	if rec.Code != http.StatusOK {
		t.Fatalf("css handler returned wrong status code: got %v want %v", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/css") {
		t.Errorf("css handler returned wrong content type: got %q", ct)
	}
	if rec.Body.Len() < 1000 {
		t.Errorf("css handler returned suspiciously small payload: %d bytes", rec.Body.Len())
	}
}

func TestServer_EmptyLogDir(t *testing.T) {
	// A missing log directory should render an empty list, not a 500.
	srv := NewServer(filepath.Join(t.TempDir(), "does-not-exist"), t.TempDir())

	rec := doRequest(t, srv, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("list handler returned wrong status code for missing log dir: got %v want %v", rec.Code, http.StatusOK)
	}
}

func TestServer_SecurityHeaders(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doRequest(t, srv, http.MethodGet, "/")
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing X-Content-Type-Options header")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("missing X-Frame-Options header")
	}
}
