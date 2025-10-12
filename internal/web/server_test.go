package web

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir, err := ioutil.TempDir("", "test-logs")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some dummy log files
	logFile1 := filepath.Join(tmpDir, "script1.log")
	if err := ioutil.WriteFile(logFile1, []byte("log content 1"), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}
	logFile2 := filepath.Join(tmpDir, "script2.log")
	if err := ioutil.WriteFile(logFile2, []byte("log content 2"), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	// Set the log directory for the handlers
	logDir = tmpDir

	// Create a new server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handleList(w, r)
		} else {
			handleView(w, r)
		}
	}))
	defer server.Close()

	// Test the list handler
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			resp.StatusCode, http.StatusOK)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedList := "<li><a href=\"/logs/script1.log\">script1.log</a></li>"
	if !strings.Contains(string(body), expectedList) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			string(body), expectedList)
	}

	// Test the view handler
	resp, err = http.Get(server.URL + "/logs/script1.log")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			resp.StatusCode, http.StatusOK)
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedView := "<pre>log content 1</pre>"
	if !strings.Contains(string(body), expectedView) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			string(body), expectedView)
	}
}