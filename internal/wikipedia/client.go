package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultLang = "en"

type Client struct {
	httpClient      *http.Client
	baseURL         string
	baseURLOverride bool
	lang            string
	userAgent       string
}

type ClientOption func(*Client)

func WithLang(lang string) ClientOption {
	return func(c *Client) {
		c.lang = lang
	}
}

func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
		c.baseURLOverride = true
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func New(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: http.DefaultClient,
		lang:       defaultLang,
		userAgent:  "light_wikipedia_cli/1.0",
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.baseURL == "" {
		c.baseURL = fmt.Sprintf("https://%s.wikipedia.org", c.lang)
	}
	return c
}

func (c *Client) restURL(parts ...string) string {
	return c.baseURL + "/api/rest_v1/" + strings.Join(parts, "/")
}

func (c *Client) actionURL(params url.Values) string {
	return c.baseURL + "/w/api.php?" + params.Encode()
}

func (c *Client) doRequest(ctx context.Context, reqURL, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	return resp, nil
}

func (c *Client) decodeJSON(resp *http.Response, target any) error {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Title  string `json:"title"`
			Detail string `json:"detail"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Detail != "" {
			return fmt.Errorf("%w: %s", ErrAPI, apiErr.Detail)
		}
		return fmt.Errorf("%w: status %d", ErrAPI, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if query == "" {
		return nil, ErrInvalidTitle
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	params := url.Values{}
	params.Set("action", "query")
	params.Set("list", "search")
	params.Set("srsearch", query)
	params.Set("srlimit", fmt.Sprintf("%d", limit))
	params.Set("format", "json")

	resp, err := c.doRequest(ctx, c.actionURL(params), "application/json")
	if err != nil {
		return nil, err
	}

	var ar actionSearchResponse
	if err := c.decodeJSON(resp, &ar); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(ar.Query.Search))
	for _, s := range ar.Query.Search {
		results = append(results, SearchResult{
			Title:   s.Title,
			Snippet: StripHTML(s.Snippet),
		})
	}
	return results, nil
}

func (c *Client) GetPage(ctx context.Context, title string) (*Page, error) {
	if title == "" {
		return nil, ErrInvalidTitle
	}

	summary, err := c.GetSummary(ctx, title)
	if err != nil {
		return nil, err
	}

	page := &Page{
		Title: summary.Title,
		URL:   summary.URL,
	}

	htmlURL := c.restURL("page", "html", url.PathEscape(title))
	resp, err := c.doRequest(ctx, htmlURL, "text/html")
	if err != nil {
		return page, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return page, nil
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return page, nil
	}

	page.Content = string(htmlBytes)
	return page, nil
}

func (c *Client) GetSummary(ctx context.Context, title string) (*Summary, error) {
	if title == "" {
		return nil, ErrInvalidTitle
	}

	base := c.restURL("page", "summary", url.PathEscape(title))
	resp, err := c.doRequest(ctx, base, "application/json")
	if err != nil {
		return nil, err
	}

	var sr summaryResponse
	if err := c.decodeJSON(resp, &sr); err != nil {
		return nil, err
	}
	return &Summary{
		Title:   sr.Title,
		Extract: sr.Extract,
		URL:     sr.URL,
	}, nil
}

func (c *Client) Random(ctx context.Context) (*Summary, error) {
	base := c.restURL("page", "random", "summary")
	resp, err := c.doRequest(ctx, base, "application/json")
	if err != nil {
		return nil, err
	}

	var rs summaryResponse
	if err := c.decodeJSON(resp, &rs); err != nil {
		return nil, err
	}
	return &Summary{
		Title:   rs.Title,
		Extract: rs.Extract,
		URL:     rs.URL,
	}, nil
}

func (c *Client) SetLang(lang string) {
	c.lang = lang
	if !c.baseURLOverride {
		c.baseURL = fmt.Sprintf("https://%s.wikipedia.org", lang)
	}
}

// StripHTML performs basic HTML tag stripping and entity decoding,
// skipping the content of <style> and <script> tags entirely.
func StripHTML(s string) string {
	var out strings.Builder
	inTag := false
	inEntity := false
	skipContent := false
	var entity strings.Builder
	var tagName strings.Builder

	tagNameStart := func() string {
		name := strings.ToLower(strings.TrimSpace(tagName.String()))
		if idx := strings.IndexAny(name, " \t\n\r>"); idx >= 0 {
			name = name[:idx]
		}
		return name
	}

	for _, r := range s {
		switch {
		case skipContent:
			if r == '<' {
				inTag = true
				tagName.Reset()
			} else if r == '>' && inTag {
				name := tagNameStart()
				if name == "/style" || name == "/script" {
					skipContent = false
				}
				inTag = false
				tagName.Reset()
			} else if inTag {
				tagName.WriteRune(r)
			}
		case inTag:
			if r == '>' {
				name := tagNameStart()
				if name == "style" || name == "script" {
					skipContent = true
				}
				inTag = false
				tagName.Reset()
			} else {
				tagName.WriteRune(r)
			}
		case r == '<':
			inTag = true
			tagName.Reset()
		case r == '&':
			inEntity = true
			entity.Reset()
		case inEntity:
			if r == ';' {
				inEntity = false
				switch entity.String() {
				case "amp":
					out.WriteRune('&')
				case "lt":
					out.WriteRune('<')
				case "gt":
					out.WriteRune('>')
				case "quot":
					out.WriteRune('"')
				case "apos":
					out.WriteRune('\'')
				case "nbsp":
					out.WriteRune(' ')
				default:
					if strings.HasPrefix(entity.String(), "#") {
						var code int
						if _, err := fmt.Sscanf(entity.String()[1:], "%d", &code); err == nil {
							out.WriteRune(rune(code))
						}
					}
				}
			} else {
				entity.WriteRune(r)
			}
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}
