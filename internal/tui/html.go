package tui

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func renderHTML(doc string) string {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return ""
	}

	n, err := html.Parse(strings.NewReader(doc))
	if err != nil {
		return stripTags(doc)
	}

	var out strings.Builder
	renderNode(n, &out, 0)
	return strings.TrimSpace(out.String())
}

func renderNode(n *html.Node, out *strings.Builder, depth int) {
	switch n.Type {
	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			out.WriteString(text)
			out.WriteRune(' ')
		}

	case html.DocumentNode:
		renderChildren(n, out, depth)

	case html.ElementNode:
		switch n.Data {
		case "h1", "h2", "h3", "h4":
			out.WriteString("\n\n")
			out.WriteString("\033[1;36m")
			renderChildren(n, out, depth)
			out.WriteString("\033[0m")
			out.WriteString("\n")

		case "p":
			out.WriteString("\n\n")
			renderChildren(n, out, depth)
			out.WriteString("\n")

		case "a":
			href := ""
			for _, a := range n.Attr {
				if a.Key == "href" {
					href = a.Val
					break
				}
			}
			renderChildren(n, out, depth)
			if href != "" && !strings.HasPrefix(href, "#") {
				out.WriteString(fmt.Sprintf(" \033[2m[%s]\033[0m", href))
			}

		case "b", "strong":
			out.WriteString("\033[1m")
			renderChildren(n, out, depth)
			out.WriteString("\033[0m")

		case "i", "em":
			out.WriteString("\033[3m")
			renderChildren(n, out, depth)
			out.WriteString("\033[0m")

		case "code", "tt":
			out.WriteString("\033[1;38;5;244m")
			renderChildren(n, out, depth)
			out.WriteString("\033[0m")

		case "ul":
			out.WriteString("\n")
			renderChildren(n, out, depth+1)
			out.WriteString("\n")

		case "ol":
			out.WriteString("\n")
			renderChildren(n, out, depth+1)
			out.WriteString("\n")

		case "li":
			out.WriteString("\n")
			prefix := "  "
			if depth > 1 {
				prefix = prefix + strings.Repeat("  ", depth-1)
			}
			if isOrderedList(n.Parent) {
				idx := siblingIndex(n)
				prefix += fmt.Sprintf("%d. ", idx+1)
			} else {
				prefix += "\033[38;5;39m•\033[0m "
			}
			out.WriteString(prefix)
			renderChildren(n, out, depth)
			out.WriteString("\n")

		case "br":
			out.WriteString("\n")

		case "hr":
			out.WriteString("\n\n────────────────────────────────────\n")

		case "dl":
			out.WriteString("\n")
			renderChildren(n, out, depth)
			out.WriteString("\n")

		case "dt":
			out.WriteString("\n\033[1m")
			renderChildren(n, out, depth)
			out.WriteString("\033[0m\n")

		case "dd":
			out.WriteString("  ")
			renderChildren(n, out, depth)
			out.WriteString("\n")

		case "blockquote":
			out.WriteString("\n\033[2m")
			renderChildren(n, out, depth)
			out.WriteString("\033[0m\n")

		case "pre":
			out.WriteString("\n\033[38;5;244m")
			renderChildren(n, out, depth)
			out.WriteString("\033[0m\n")

		case "table":
			out.WriteString("\n")
			renderChildren(n, out, depth)
			out.WriteString("\n")

		case "tr":
			out.WriteString("\n")
			renderChildren(n, out, depth)
			out.WriteString("\n")

		case "td", "th":
			out.WriteString("  ")
			renderChildren(n, out, depth)

		case "div", "span", "section", "article", "nav", "header", "footer", "main", "figure", "figcaption":
			renderChildren(n, out, depth)

		case "img":
			alt := ""
			for _, a := range n.Attr {
				if a.Key == "alt" {
					alt = a.Val
					break
				}
			}
			if alt != "" {
				out.WriteString(fmt.Sprintf("\033[2m[image: %s]\033[0m", alt))
			}

		case "style", "script", "noscript", "link", "meta", "title", "head":
			return

		default:
			renderChildren(n, out, depth)
		}
	}
}

func renderChildren(n *html.Node, out *strings.Builder, depth int) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderNode(c, out, depth)
	}
}

func isOrderedList(n *html.Node) bool {
	return n != nil && n.Type == html.ElementNode && n.Data == "ol"
}

func siblingIndex(n *html.Node) int {
	idx := 0
	for c := n.Parent.FirstChild; c != nil && c != n; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "li" {
			idx++
		}
	}
	return idx
}

func stripTags(s string) string {
	var out strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			out.WriteRune(r)
		}
	}
	return out.String()
}
