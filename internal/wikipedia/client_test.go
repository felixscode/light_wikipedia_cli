package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		limit   int
		results []SearchResult
		wantLen int
		wantErr bool
	}{
		{
			name:  "success",
			query: "Go programming",
			limit: 2,
			results: []SearchResult{
				{Title: "Go (programming language)", Snippet: "Go is a statically typed..."},
				{Title: "Go (game)", Snippet: "Go is an abstract strategy..."},
			},
			wantLen: 2,
		},
		{
			name:    "empty query",
			query:   "",
			limit:   10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(routeMap{
				"/w/api.php": func(w http.ResponseWriter, r *http.Request) {
					srch := r.URL.Query().Get("srsearch")
					if srch == "" {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					writeJSON(w, searchResultsJSON(tt.results))
				},
			})
			defer srv.Close()

			c := testClient(srv.URL)

			results, err := c.Search(context.Background(), tt.query, tt.limit)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != tt.wantLen {
				t.Fatalf("got %d results, want %d", len(results), tt.wantLen)
			}
		})
	}
}

func TestSearchLimitClamping(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		wantBody string
	}{
		{"default when zero", 0, "srlimit=10"},
		{"default when negative", -5, "srlimit=10"},
		{"default when over 50", 100, "srlimit=10"},
		{"pass through valid", 25, "srlimit=25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seenQuery string
			srv := newTestServer(routeMap{
				"/w/api.php": func(w http.ResponseWriter, r *http.Request) {
					seenQuery = r.URL.RawQuery
					writeJSON(w, searchResultsJSON(nil))
				},
			})
			defer srv.Close()

			c := testClient(srv.URL)
			_, _ = c.Search(context.Background(), "test", tt.limit)
			if !strings.Contains(seenQuery, tt.wantBody) {
				t.Errorf("query %q does not contain %q", seenQuery, tt.wantBody)
			}
		})
	}
}

func TestSearchAPIError(t *testing.T) {
	srv := newTestServer(routeMap{
		"/w/api.php": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"title":"too many requests","detail":"rate limited"}`)
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	_, err := c.Search(context.Background(), "test", 10)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("error %q does not contain detail message", err)
	}
}

func TestSearchNonJSONError(t *testing.T) {
	srv := newTestServer(routeMap{
		"/w/api.php": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal error"))
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	_, err := c.Search(context.Background(), "test", 10)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("error %q does not contain status code", err)
	}
}

func TestGetSummary(t *testing.T) {
	srv := newTestServer(routeMap{
		"/api/rest_v1/page/summary/Go (programming language)": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, summaryJSON(
				"Go (programming language)",
				"Go is a statically typed compiled programming language...",
				"https://en.wikipedia.org/wiki/Go_(programming_language)",
			))
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)

	summary, err := c.GetSummary(context.Background(), "Go (programming language)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Title != "Go (programming language)" {
		t.Fatalf("got title %q, want %q", summary.Title, "Go (programming language)")
	}
	if summary.URL != "https://en.wikipedia.org/wiki/Go_(programming_language)" {
		t.Fatalf("got URL %q, want %q", summary.URL, "https://en.wikipedia.org/wiki/Go_(programming_language)")
	}
}

func TestGetSummaryEmptyTitle(t *testing.T) {
	c := New()
	_, err := c.GetSummary(context.Background(), "")
	if err != ErrInvalidTitle {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestGetSummaryNotFound(t *testing.T) {
	srv := newTestServer(routeMap{
		"/api/rest_v1/page/summary/NotReal": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	_, err := c.GetSummary(context.Background(), "NotReal")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetPage(t *testing.T) {
	htmlContent := "<html><body><p>Go is a programming language.</p></body></html>"
	srv := newTestServer(routeMap{
		"/api/rest_v1/page/summary/Go": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, summaryJSON("Go", "Go is a programming language", "https://en.wikipedia.org/wiki/Go"))
		},
		"/api/rest_v1/page/html/Go": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(htmlContent))
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	page, err := c.GetPage(context.Background(), "Go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Title != "Go" {
		t.Fatalf("got title %q, want %q", page.Title, "Go")
	}
	if page.URL != "https://en.wikipedia.org/wiki/Go" {
		t.Fatalf("got URL %q, want %q", page.URL, "https://en.wikipedia.org/wiki/Go")
	}
	if !strings.Contains(page.Content, "Go is a programming language") {
		t.Fatalf("content %q does not contain expected text", page.Content)
	}
}

func TestGetPageSummaryFails(t *testing.T) {
	srv := newTestServer(routeMap{
		"/api/rest_v1/page/summary/Go": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	_, err := c.GetPage(context.Background(), "Go")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetPageHTMLEndpointFails(t *testing.T) {
	srv := newTestServer(routeMap{
		"/api/rest_v1/page/summary/Go": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, summaryJSON("Go", "Go is a programming language", "https://en.wikipedia.org/wiki/Go"))
		},
		"/api/rest_v1/page/html/Go": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	page, err := c.GetPage(context.Background(), "Go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Title != "Go" {
		t.Fatalf("got title %q, want %q", page.Title, "Go")
	}
	if page.Content != "" {
		t.Fatalf("expected empty content on HTML error, got %q", page.Content)
	}
}

func TestGetPageEmptyTitle(t *testing.T) {
	c := New()
	_, err := c.GetPage(context.Background(), "")
	if err != ErrInvalidTitle {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestRandom(t *testing.T) {
	srv := newTestServer(routeMap{
		"/api/rest_v1/page/random/summary": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, summaryJSON("Random Article", "This is a random article...", "https://en.wikipedia.org/wiki/Random_Article"))
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	summary, err := c.Random(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Title != "Random Article" {
		t.Fatalf("got title %q, want %q", summary.Title, "Random Article")
	}
}

func TestNewDefaults(t *testing.T) {
	c := New()
	if c.lang != "en" {
		t.Errorf("default lang = %q, want %q", c.lang, "en")
	}
	if c.baseURL != "https://en.wikipedia.org" {
		t.Errorf("default baseURL = %q, want %q", c.baseURL, "https://en.wikipedia.org")
	}
	if c.userAgent != "light_wikipedia_cli/1.0" {
		t.Errorf("default userAgent = %q, want %q", c.userAgent, "light_wikipedia_cli/1.0")
	}
}

func TestWithUserAgent(t *testing.T) {
	c := New(WithUserAgent("custom-agent/1.0"))
	if c.userAgent != "custom-agent/1.0" {
		t.Fatalf("got userAgent %q, want %q", c.userAgent, "custom-agent/1.0")
	}
}

func TestSetLang(t *testing.T) {
	c := New()
	if c.lang != "en" {
		t.Fatalf("default lang: got %q, want %q", c.lang, "en")
	}
	c.SetLang("de")
	if c.lang != "de" {
		t.Fatalf("after SetLang: got %q, want %q", c.lang, "de")
	}
	if c.baseURL != "https://de.wikipedia.org" {
		t.Fatalf("baseURL after SetLang: got %q, want %q", c.baseURL, "https://de.wikipedia.org")
	}
}

func TestSetLangWithOverride(t *testing.T) {
	c := New(WithBaseURL("http://mock:8080"))
	c.SetLang("fr")
	if c.lang != "fr" {
		t.Fatalf("lang: got %q, want %q", c.lang, "fr")
	}
	if c.baseURL != "http://mock:8080" {
		t.Fatalf("baseURL should not change with override, got %q", c.baseURL)
	}
}

func TestWithLang(t *testing.T) {
	c := New(WithLang("fr"))
	if c.lang != "fr" {
		t.Fatalf("got lang %q, want %q", c.lang, "fr")
	}
	if c.baseURL != "https://fr.wikipedia.org" {
		t.Fatalf("baseURL: got %q, want %q", c.baseURL, "https://fr.wikipedia.org")
	}
}

func TestWithBaseURL(t *testing.T) {
	c := New(WithBaseURL("http://localhost:9999"))
	if c.baseURL != "http://localhost:9999" {
		t.Fatalf("got baseURL %q, want %q", c.baseURL, "http://localhost:9999")
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<p>Hello World</p>", "Hello World"},
		{"Hello &amp; goodbye", "Hello & goodbye"},
		{"&lt;tag&gt;", "<tag>"},
		{"&quot;quoted&quot;", "\"quoted\""},
		{"&apos;single&apos;", "'single'"},
		{"foo &nbsp; bar", "foo   bar"},
		{"<div><b>Bold</b> text</div>", "Bold text"},
		{"No HTML here", "No HTML here"},
		{"Multi\nline\ntext", "Multi\nline\ntext"},
		{"&#65;&#66;&#67;", "ABC"},
		{"", ""},
	}

	for _, tt := range tests {
		got := StripHTML(tt.input)
		if got != tt.want {
			t.Errorf("StripHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestContextCancelled(t *testing.T) {
	srv := newTestServer(routeMap{
		"/w/api.php": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, searchResultsJSON(nil))
		},
	})
	defer srv.Close()

	c := testClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Search(ctx, "test", 10)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// --- test helpers ---

type routeMap map[string]http.HandlerFunc

func newTestServer(routes routeMap) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for route, handler := range routes {
			if r.URL.Path == route {
				handler(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no route for %s", r.URL.Path)
	}))
}

func testClient(srvURL string) *Client {
	return New(
		WithHTTPClient(httptest.NewServer(nil).Client()),
		WithUserAgent("test"),
		WithBaseURL(srvURL),
	)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func searchResultsJSON(results []SearchResult) any {
	search := make([]map[string]any, len(results))
	for i, r := range results {
		search[i] = map[string]any{
			"title":     r.Title,
			"snippet":   r.Snippet,
			"pageid":    i + 1,
			"wordcount": 100,
			"size":      1000,
		}
	}
	return map[string]any{
		"query": map[string]any{
			"search": search,
		},
	}
}

func summaryJSON(title, extract, pageURL string) any {
	return map[string]any{
		"title":   title,
		"extract": extract,
		"content_urls": map[string]any{
			"desktop": map[string]any{"page": pageURL},
			"mobile":  map[string]any{"page": pageURL},
		},
		"description": "",
		"pageid":      12345,
	}
}

var _ ClientOption = func(c *Client) {}
