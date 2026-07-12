// Command docgen builds a lightweight static documentation site from
// Markdown files. It walks a content directory, converts every .md file to
// HTML, and writes the result into an output directory alongside a small
// navigation sidebar. No external services or databases are required.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// page holds the data passed to the HTML template for a single document.
type page struct {
	Title    string
	Content  template.HTML
	Nav      []navItem
	Active   string
	IsHome   bool
	Year     int
	SiteName string
}

// navItem is a single entry in the navigation sidebar.
type navItem struct {
	Title string
	Href  string
}

// pageMeta is gathered while scanning the content tree.
type pageMeta struct {
	title string
	href  string
}

func main() {
	contentDir := flag.String("content", "doc/content", "directory of Markdown sources")
	outDir := flag.String("out", "doc/public", "output directory for generated HTML")
	siteName := flag.String("site", "light_wikipedia_cli docs", "site name shown in the header")
	skillPath := flag.String("skill", "", "optional path to a SKILL.md copied verbatim into the output")
	llmsPath := flag.String("llms", "", "optional path to an llms.txt copied verbatim into the output")
	flag.Parse()

	if err := run(*contentDir, *outDir, *siteName, *skillPath, *llmsPath); err != nil {
		fmt.Fprintf(os.Stderr, "docgen: %v\n", err)
		os.Exit(1)
	}
}

func run(contentDir, outDir, siteName, skillPath, llmsPath string) error {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	// Explicit navigation order: Home, Quickstart, Usage, then the rest.
	order := map[string]int{
		"index":           0,
		"getting-started": 1,
		"cli":             2,
	}
	orderKey := func(href string) int {
		base := strings.TrimPrefix(strings.TrimSuffix(href, ".html"), "/")
		if k, ok := order[base]; ok {
			return k
		}
		return 100
	}

	type rendered struct {
		title, body, href, outRel string
	}
	var pages []rendered
	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, err := filepath.Rel(contentDir, path)
		if err != nil {
			return err
		}
		outRel := strings.TrimSuffix(rel, ".md") + ".html"
		title, htmlBody, err := render(md, path)
		if err != nil {
			return fmt.Errorf("render %s: %w", path, err)
		}
		href := "/" + filepath.ToSlash(outRel)
		pages = append(pages, rendered{title: title, body: htmlBody, href: href, outRel: outRel})
		return nil
	})
	if err != nil {
		return err
	}

	// Sort into the desired navigation order (stable within each tier).
	sort.SliceStable(pages, func(i, j int) bool {
		return orderKey(pages[i].href) < orderKey(pages[j].href)
	})

	metas := make([]pageMeta, 0, len(pages))
	for _, p := range pages {
		metas = append(metas, pageMeta{title: p.title, href: p.href})
	}

	for _, p := range pages {
		if err := writePage(md, outDir, p.outRel, p.title, p.body, metas, p.href, p.href == "/index.html", siteName); err != nil {
			return err
		}
	}

	if err := writeAgentsTxt(outDir, siteName, metas); err != nil {
		return err
	}

	if skillPath != "" {
		if err := copySkill(outDir, skillPath); err != nil {
			return err
		}
	}

	if llmsPath != "" {
		if err := copyLlmsTxt(outDir, llmsPath); err != nil {
			return err
		}
	}

	fmt.Printf("docgen: generated %d pages into %s\n", len(metas), outDir)
	return nil
}

// copySkill copies a SKILL.md verbatim into the output so agents can read and
// import it directly from the published site.
func copySkill(outDir, skillPath string) error {
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("read skill %s: %w", skillPath, err)
	}
	dst := filepath.Join(outDir, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("docgen: copied %s -> %s\n", skillPath, dst)
	return nil
}

// copyLlmsTxt copies an llms.txt verbatim into the output so AI agents
// visiting the published site can discover the project documentation.
func copyLlmsTxt(outDir, llmsPath string) error {
	data, err := os.ReadFile(llmsPath)
	if err != nil {
		return fmt.Errorf("read llms %s: %w", llmsPath, err)
	}
	dst := filepath.Join(outDir, "llms.txt")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("docgen: copied %s -> %s\n", llmsPath, dst)
	return nil
}

// writeAgentsTxt emits an agents.txt at the output root so AI agents visiting
// the published site can discover the documentation and available skills.
func writeAgentsTxt(outDir, siteName string, metas []pageMeta) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# agents.txt\n")
	fmt.Fprintf(&b, "# Machine-readable instructions for AI agents browsing this site.\n\n")
	fmt.Fprintf(&b, "[site]\n")
	fmt.Fprintf(&b, "name: %s\n", siteName)
	fmt.Fprintf(&b, "description: Lightweight documentation for the light_wikipedia_cli CLI and MCP server.\n")
	fmt.Fprintf(&b, "generator: docgen (https://github.com/felixschelling/light_wikipedia_cli)\n")
	fmt.Fprintf(&b, "llms: /llms.txt\n")
	fmt.Fprintf(&b, "skill: /SKILL.md\n\n")

	fmt.Fprintf(&b, "[pages]\n")
	for _, m := range metas {
		fmt.Fprintf(&b, "- %s: %s\n", m.title, m.href)
	}
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "[skills]\n")
	fmt.Fprintf(&b, "name: setup-wiki\n")
	fmt.Fprintf(&b, "description: Scaffold and build the static documentation site for this project from Markdown sources.\n")
	fmt.Fprintf(&b, "trigger: When the user asks to set up, build, or regenerate the project wiki/docs site under /doc.\n")
	fmt.Fprintf(&b, "steps:\n")
	fmt.Fprintf(&b, "  1. Ensure Go 1.26+ is installed.\n")
	fmt.Fprintf(&b, "  2. Add Markdown sources under doc/content/ (one .md file per page; the first H1 is the page title).\n")
	fmt.Fprintf(&b, "  3. Run `make docs` to invoke cmd/docgen, which converts each .md to HTML in doc/public/ and writes llms.txt + agents.txt.\n")
	fmt.Fprintf(&b, "  4. Serve doc/public/ with any static file server (e.g. `make serve` or `python3 -m http.server -d doc/public`).\n")
	fmt.Fprintf(&b, "  5. The generated doc/public/ directory is gitignored; only doc/content/ is committed.\n")
	fmt.Fprintf(&b, "notes: The site is fully static (no database or runtime server). Navigation is auto-generated from all pages.\n")

	dst := filepath.Join(outDir, "agents.txt")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(b.String()), 0o644)
}

// render reads a Markdown file and returns its title (first H1) and HTML body.
func render(md goldmark.Markdown, path string) (string, string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	title := deriveTitle(path, string(src))
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", "", err
	}
	return title, buf.String(), nil
}

// deriveTitle returns the first top-level heading, falling back to the file name.
func deriveTitle(path, src string) string {
	for _, line := range strings.Split(src, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# ") {
			return strings.TrimSpace(t[2:])
		}
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// writePage renders one page to disk using the site template.
func writePage(md goldmark.Markdown, outDir, outRel, title, body string, metas []pageMeta, active string, isHome bool, siteName string) error {
	nav := make([]navItem, 0, len(metas)+1)
	nav = append(nav, navItem{Title: "Home", Href: "/index.html"})
	for _, m := range metas {
		if m.href == "/index.html" {
			continue
		}
		nav = append(nav, navItem{Title: m.title, Href: m.href})
	}
	p := page{
		Title:    title,
		Content:  template.HTML(body),
		Nav:      nav,
		Active:   "/" + filepath.ToSlash(outRel),
		IsHome:   isHome,
		Year:     time.Now().Year(),
		SiteName: siteName,
	}
	tmpl, err := template.New("page").Parse(pageTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p); err != nil {
		return err
	}
	dst := filepath.Join(outDir, filepath.FromSlash(outRel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, buf.Bytes(), 0o644)
}

const pageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{ .Title }} — {{ .SiteName }}</title>
<style>
:root { --fg:#1a1a1a; --muted:#6b7280; --bg:#ffffff; --accent:#0ea5e9; --border:#e5e7eb; }
* { box-sizing: border-box; }
body { margin:0; font:16px/1.6 system-ui,-apple-system,Segoe UI,Roboto,sans-serif; color:var(--fg); background:var(--bg); }
header { border-bottom:1px solid var(--border); padding:1rem 1.5rem; }
header a { color:var(--accent); text-decoration:none; font-weight:600; }
.layout { display:flex; align-items:flex-start; }
nav { width:240px; flex:0 0 240px; padding:1.5rem 1rem; border-right:1px solid var(--border); min-height:calc(100vh - 57px); }
nav ul { list-style:none; padding:0; margin:0; }
nav li { margin:.25rem 0; }
nav a { color:var(--muted); text-decoration:none; }
nav a:hover, nav a.active { color:var(--accent); }
main { flex:1; padding:1.5rem 2rem; max-width:820px; }
main h1 { margin-top:0; }
main code { background:#f3f4f6; padding:.15em .35em; border-radius:4px; }
main pre { background:#0f172a; color:#e2e8f0; padding:1rem; border-radius:8px; overflow:auto; }
main pre code { background:none; padding:0; }
main a { color:var(--accent); }
footer { color:var(--muted); font-size:.85rem; padding:1.5rem 2rem; border-top:1px solid var(--border); }
@media (max-width:720px){ .layout{ flex-direction:column; } nav{ width:auto; border-right:none; border-bottom:1px solid var(--border); min-height:auto; } }
</style>
</head>
<body>
<header><a href="/index.html">{{ .SiteName }}</a></header>
<div class="layout">
<nav><ul>
{{ range .Nav }}<li><a href="{{ .Href }}"{{ if eq .Href $.Active }} class="active"{{ end }}>{{ .Title }}</a></li>
{{ end }}</ul></nav>
<main>
{{ if not .IsHome }}<p><a href="/index.html">← Back to main</a></p>
{{ end }}{{ .Content }}
</main>
</div>
<footer>© {{ .Year }} {{ .SiteName }} — generated by docgen</footer>
</body>
</html>`
