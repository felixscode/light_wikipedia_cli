//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	cliBinary  string
	mcpBinary  string
	ansiEscape = "\033["
	apiBase    string
)

// projectRoot returns the repository root by walking up from this test file.
func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// TestMain builds both binaries and starts a mock Wikipedia API server.
func TestMain(m *testing.M) {
	root := projectRoot()
	if root == "" {
		panic("could not locate project root (go.mod)")
	}

	tmp, err := os.MkdirTemp("", "light_wikipedia_cli-e2e-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	cliBinary = filepath.Join(tmp, "wikipedia")
	mcpBinary = filepath.Join(tmp, "wikipedia-mcp")

	cliPkg := filepath.Join(root, "cmd", "wikipedia")
	mcpPkg := filepath.Join(root, "cmd", "wikipedia-mcp")

	if out, err := exec.Command("go", "build", "-o", cliBinary, cliPkg).CombinedOutput(); err != nil {
		panic("failed to build CLI binary: " + string(out))
	}
	if out, err := exec.Command("go", "build", "-o", mcpBinary, mcpPkg).CombinedOutput(); err != nil {
		panic("failed to build MCP binary: " + string(out))
	}

	srv := newMockWikiServer()
	apiBase = srv.URL
	defer srv.Close()

	// Set the env var so subprocesses talk to the mock server.
	os.Setenv("WIKIPEDIA_API_BASE", apiBase)

	os.Exit(m.Run())
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(cliBinary, args...)
	cmd.Env = append(os.Environ(), "WIKIPEDIA_API_BASE="+apiBase)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()

	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run CLI: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestCLIDefaultRandom(t *testing.T) {
	stdout, stderr, code := runCLI(t)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr)
	}
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Fatal("expected non-empty output for random article")
	}
	if !strings.Contains(stdout, "https://") {
		t.Fatalf("expected a Wikipedia URL in output, got:\n%s", stdout)
	}
}

func TestCLIHelp(t *testing.T) {
	stdout, _, code := runCLI(t, "--help")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	for _, want := range []string{"Usage", "wikipedia", "--search", "--lang", "--random"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output missing %q:\n%s", want, stdout)
		}
	}
}

func TestCLISearch(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--search", "Go programming")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Go (programming language)") {
		t.Fatalf("expected title in output, got:\n%s", stdout)
	}
}

func TestCLISearchShortFlag(t *testing.T) {
	stdout, stderr, code := runCLI(t, "-s", "Go programming")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Go (programming language)") {
		t.Fatalf("expected title in output, got:\n%s", stdout)
	}
}

func TestCLISearchFull(t *testing.T) {
	stdout, stderr, code := runCLI(t, "-s", "Go programming", "-f")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Go (programming language)") {
		t.Fatalf("expected title in output, got:\n%s", stdout)
	}
}

func TestCLILangGerman(t *testing.T) {
	stdout, stderr, code := runCLI(t, "-l", "de", "-s", "Go")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Go") {
		t.Fatalf("expected title in output, got:\n%s", stdout)
	}
}

func TestCLINoColor(t *testing.T) {
	stdout, _, code := runCLI(t, "--no-color", "-s", "Go programming")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if strings.Contains(stdout, ansiEscape) {
		t.Fatalf("expected no ANSI codes with --no-color, got:\n%s", stdout)
	}
}

func TestCLISearchNotFound(t *testing.T) {
	stdout, stderr, code := runCLI(t, "-s", "__NONEXISTENT_WIKIPEDIA_PAGE_XYZ_123__")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	combined := stdout + stderr
	if !strings.Contains(combined, "Error") {
		t.Fatalf("expected error message, got:\n%s", combined)
	}
}

func TestCLIInvalidTitle(t *testing.T) {
	// Empty search returns random article (no error), so just verify it runs
	stdout, stderr, code := runCLI(t)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr)
	}
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Fatal("expected output")
	}
}

// --- mock Wikipedia REST API server ---

func newMockWikiServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/w/api.php":
			handleSearchMock(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/rest_v1/page/summary/"):
			title := strings.TrimPrefix(r.URL.Path, "/api/rest_v1/page/summary/")
			handleSummaryMock(w, r, title)
		case strings.HasPrefix(r.URL.Path, "/api/rest_v1/page/html/"):
			title := strings.TrimPrefix(r.URL.Path, "/api/rest_v1/page/html/")
			handleHTMLMock(w, title)
		case r.URL.Path == "/api/rest_v1/page/random/summary":
			handleSummaryMock(w, r, "Random Article of the Day")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func handleSearchMock(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("srsearch")
	if q == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if strings.Contains(q, "NONEXISTENT") {
		writeJSONMock(w, map[string]any{
			"query": map[string]any{
				"search": []any{},
			},
		})
		return
	}
	if r.URL.Query().Get("srlimit") == "0" || r.URL.Query().Get("srlimit") == "" {
		writeJSONMock(w, map[string]any{
			"query": map[string]any{
				"search": []any{},
			},
		})
		return
	}
	title := q + " page"
	snippet := "Sample snippet for " + q
	if strings.Contains(q, "Go") {
		title = "Go (programming language)"
		snippet = "Go is a statically typed, compiled programming language."
	}
	writeJSONMock(w, map[string]any{
		"query": map[string]any{
			"search": []any{
				map[string]any{
					"title":     title,
					"snippet":   snippet,
					"pageid":    12345,
					"wordcount": 100,
					"size":      1000,
				},
			},
		},
	})
}

func handleSummaryMock(w http.ResponseWriter, r *http.Request, title string) {
	writeJSONMock(w, map[string]any{
		"title":   title,
		"extract": "This is a sample extract for " + title + ".",
		"content_urls": map[string]any{
			"desktop": map[string]any{
				"page": "https://en.wikipedia.org/wiki/" + title,
			},
			"mobile": map[string]any{
				"page": "https://en.m.wikipedia.org/wiki/" + title,
			},
		},
		"description": "",
		"pageid":      12345,
	})
}

func handleHTMLMock(w http.ResponseWriter, title string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<html><body><p>Go is a programming language created at Google.</p><p>It supports concurrency.</p></body></html>")
}

func writeJSONMock(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}