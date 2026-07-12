package tfmt

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/felixschelling/light_wikipedia_cli/internal/wikipedia"
	"golang.org/x/term"
)

const (
	cyan   = "\033[36;1m"
	white  = "\033[37;1m"
	dim    = "\033[2m"
	reset  = "\033[0m"
	yellow = "\033[33;1m"
	red    = "\033[31;1m"
)

type Formatter struct {
	width   int
	noColor bool
}

func New(noColor bool) *Formatter {
	f := &Formatter{noColor: noColor}
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		f.width = w
	} else {
		f.width = 80
	}
	if f.width > 120 {
		f.width = 120
	}
	return f
}

func (f *Formatter) color(code string) string {
	if f.noColor {
		return ""
	}
	return code
}

func (f *Formatter) resetStyle() string {
	if f.noColor {
		return ""
	}
	return reset
}

func (f *Formatter) Summary(w io.Writer, s *wikipedia.Summary) {
	fmt.Fprintf(w, "%s%s%s\n\n", f.color(cyan), s.Title, f.resetStyle())
	fmt.Fprintln(w, strings.Repeat("─", f.width))
	fmt.Fprintln(w)

	text := f.wrap(s.Extract)
	fmt.Fprintln(w, text)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s%s%s\n", f.color(dim), s.URL, f.resetStyle())
}

func (f *Formatter) Page(w io.Writer, p *wikipedia.Page) {
	fmt.Fprintf(w, "%s%s%s\n\n", f.color(cyan), p.Title, f.resetStyle())
	fmt.Fprintln(w, strings.Repeat("─", f.width))
	fmt.Fprintln(w)

	if p.Content != "" {
		text := f.wrap(wikipedia.StripHTML(p.Content))
		fmt.Fprintln(w, text)
	} else {
		fmt.Fprintf(w, "%s(Full content not available in plain text. Visit the URL below.)%s\n", f.color(dim), f.resetStyle())
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s%s%s\n", f.color(dim), p.URL, f.resetStyle())
}

func (f *Formatter) SearchResults(w io.Writer, results []wikipedia.SearchResult) {
	fmt.Fprintf(w, "%sSearch Results%s\n\n", f.color(cyan), f.resetStyle())
	fmt.Fprintln(w, strings.Repeat("─", f.width))
	fmt.Fprintln(w)

	for i, r := range results {
		fmt.Fprintf(w, "%s%d.%s %s\n", f.color(yellow), i+1, f.resetStyle(), r.Title)
		if r.Snippet != "" {
			fmt.Fprintf(w, "   %s%s%s\n", f.color(dim), f.wrap(r.Snippet), f.resetStyle())
		}
		fmt.Fprintln(w)
	}
}

func (f *Formatter) Error(w io.Writer, msg string) {
	fmt.Fprintf(w, "%sError: %s%s\n", f.color(red), msg, f.resetStyle())
}

func (f *Formatter) wrap(s string) string {
	if f.width <= 0 {
		return s
	}

	var out strings.Builder
	for _, line := range strings.Split(s, "\n") {
		words := strings.Fields(line)
		if len(words) == 0 {
			out.WriteString("\n")
			continue
		}

		current := 0
		for _, word := range words {
			if current > 0 && current+len(word)+1 > f.width {
				out.WriteString("\n")
				current = 0
			}
			if current > 0 {
				out.WriteString(" ")
				current++
			}
			out.WriteString(word)
			current += len(word)
		}
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n")
}
