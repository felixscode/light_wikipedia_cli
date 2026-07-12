package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/felixschelling/light_wikipedia_cli/internal/wikipedia"
)

type page int

const (
	pageMenu page = iota
	pageRandom
	pageSearch
	pageSearchActions
	pageSummary
	pageRead
	pageMCPSettings
)

type (
	errMsg     struct{ error }
	randomMsg  struct{ summary *wikipedia.Summary }
	searchMsg  struct{ results []wikipedia.SearchResult }
	summaryMsg struct{ summary *wikipedia.Summary }
	pageMsg    struct{ page *wikipedia.Page }
)

type model struct {
	client  *wikipedia.Client
	ctx     context.Context
	cancel  context.CancelFunc
	current page
	prev    page
	loading bool
	err     error
	width   int
	height  int

	random        *wikipedia.Summary
	randomLoading bool

	menuCursor int

	searchInput   textinput.Model
	searchResults []wikipedia.SearchResult
	searchCursor  int
	searchFocused bool
	searchLoading bool
	searchQuery   string

	actionCursor int

	summary        *wikipedia.Summary
	summaryTitle   string
	summaryLoading bool

	readTitle    string
	readContent  string
	readViewport viewport.Model
	readURL      string

	mcpLang    string
	mcpCursor  int
	mcpRunning bool

	quitting bool
}

func New(client *wikipedia.Client) *model {
	ti := textinput.New()
	ti.Placeholder = "Search Wikipedia..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	ctx, cancel := context.WithCancel(context.Background())

	return &model{
		client:       client,
		ctx:          ctx,
		cancel:       cancel,
		current:      pageMenu,
		searchInput:  ti,
		mcpLang:      "en",
		readViewport: viewport.New(80, 20),
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = max(20, msg.Width-10)
		m.readViewport.Width = msg.Width - 4
		m.readViewport.Height = msg.Height - 8
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case errMsg:
		m.loading = false
		m.randomLoading = false
		m.searchLoading = false
		m.summaryLoading = false
		m.err = msg.error
		return m, nil

	case randomMsg:
		m.loading = false
		m.randomLoading = false
		m.random = msg.summary
		return m, nil

	case searchMsg:
		m.searchLoading = false
		m.searchResults = msg.results
		m.searchCursor = 0
		return m, nil

	case summaryMsg:
		m.summaryLoading = false
		m.summary = msg.summary
		m.current = pageSummary
		return m, nil

	case pageMsg:
		m.loading = false
		m.readTitle = msg.page.Title
		content := msg.page.Content
		if content == "" {
			content = "(Full content not available)"
		}
		rendered := renderHTML(content)
		m.readContent = rendered
		m.readViewport.SetContent(rendered)
		m.readViewport.GotoTop()
		m.readURL = msg.page.URL
		m.current = pageRead
		return m, nil

	default:
		return m, nil
	}
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		m.quitting = true
		m.cancel()
		return m, tea.Quit
	}

	switch m.current {
	case pageMenu:
		return m.handleMenuKey(msg)
	case pageRandom:
		return m.handleRandomKey(msg)
	case pageSearch:
		return m.handleSearchKey(msg)
	case pageSearchActions:
		return m.handleSearchActionsKey(msg)
	case pageSummary:
		return m.handleSummaryKey(msg)
	case pageRead:
		return m.handleReadKey(msg)
	case pageMCPSettings:
		return m.handleMCPSettingsKey(msg)
	}
	return m, nil
}

func (m *model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < 3 {
			m.menuCursor++
		}
	case "enter":
		switch m.menuCursor {
		case 0:
			m.current = pageRandom
			m.randomLoading = true
			m.err = nil
			return m, fetchRandom(m.client, m.ctx)
		case 1:
			m.current = pageSearch
			m.searchFocused = false
			m.searchResults = nil
			m.searchInput.SetValue("")
			return m, nil
		case 2:
			m.current = pageMCPSettings
			return m, nil
		case 3:
			m.quitting = true
			m.cancel()
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) handleRandomKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.random != nil {
			m.loading = true
			return m, fetchPage(m.client, m.ctx, m.random.Title)
		}
	case "up":
		m.current = pageMenu
	case "down":
		m.randomLoading = true
		m.err = nil
		return m, fetchRandom(m.client, m.ctx)
	}
	return m, nil
}

func (m *model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = pageMenu
		m.searchResults = nil
		m.searchInput.SetValue("")
		return m, nil
	case "enter":
		if m.searchFocused && m.searchInput.Value() != "" {
			m.searchLoading = true
			m.searchQuery = m.searchInput.Value()
			m.err = nil
			return m, performSearch(m.client, m.ctx, m.searchQuery, 10)
		}
		if !m.searchFocused && len(m.searchResults) > 0 {
			m.prev = m.current
			m.current = pageSearchActions
			m.actionCursor = 0
		}
	case "up", "k":
		if !m.searchFocused && m.searchCursor > 0 {
			m.searchCursor--
		} else if !m.searchFocused && m.searchCursor == 0 {
			m.searchFocused = true
			m.searchInput.Focus()
		}
	case "down", "j":
		if m.searchFocused {
			m.searchFocused = false
			m.searchInput.Blur()
			m.searchCursor = 0
		} else if m.searchCursor < len(m.searchResults)-1 {
			m.searchCursor++
		}
	}

	if m.searchFocused {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) handleSearchActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = m.prev
	case "up", "k":
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case "down", "j":
		if m.actionCursor < 1 {
			m.actionCursor++
		}
	case "enter":
		title := m.searchResults[m.searchCursor].Title
		switch m.actionCursor {
		case 0:
			m.summaryLoading = true
			m.err = nil
			m.summaryTitle = title
			return m, fetchSummary(m.client, m.ctx, title)
		case 1:
			m.loading = true
			return m, fetchPage(m.client, m.ctx, title)
		}
	}
	return m, nil
}

func (m *model) handleSummaryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.summary != nil {
			m.loading = true
			return m, fetchPage(m.client, m.ctx, m.summary.Title)
		}
	case "esc":
		m.current = pageSearch
	}
	return m, nil
}

func (m *model) handleReadKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.current = pageMenu
	default:
		var cmd tea.Cmd
		m.readViewport, cmd = m.readViewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) handleMCPSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.mcpCursor > 0 {
			m.mcpCursor--
		}
	case "down", "j":
		if m.mcpCursor < 2 {
			m.mcpCursor++
		}
	case "enter":
		switch m.mcpCursor {
		case 0:
			switch m.mcpLang {
			case "en":
				m.mcpLang = "de"
			case "de":
				m.mcpLang = "fr"
			case "fr":
				m.mcpLang = "es"
			case "es":
				m.mcpLang = "ja"
			case "ja":
				m.mcpLang = "en"
			}
		case 2:
			m.current = pageMenu
		}
	case "esc":
		m.current = pageMenu
	}
	return m, nil
}

func (m *model) View() string {
	if m.quitting {
		return ""
	}

	switch m.current {
	case pageMenu:
		return m.menuView()
	case pageRandom:
		return m.randomView()
	case pageSearch:
		return m.searchView()
	case pageSearchActions:
		return m.searchActionsView()
	case pageSummary:
		return m.summaryView()
	case pageRead:
		return m.readView()
	case pageMCPSettings:
		return m.mcpSettingsView()
	}
	return ""
}

func (m *model) menuView() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("Wikipedia Terminal")

	b.WriteString(title)
	b.WriteString("\n\n")

	items := []string{
		"Random Article",
		"Search Wikipedia",
		"MCP Server Settings",
		"Quit",
	}

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	itemStyle := lipgloss.NewStyle()

	for i, item := range items {
		prefix := "  "
		if i == m.menuCursor {
			prefix = cursorStyle.Render("▸ ")
		}
		b.WriteString(fmt.Sprintf("  %s%s\n", prefix, itemStyle.Render(item)))
	}

	b.WriteString("\n  ─────────────────────\n")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render("  ↑/↓ Navigate  Enter Select  Esc Quit"))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m *model) randomView() string {
	if m.randomLoading {
		return lipgloss.NewStyle().Padding(1, 2).Render("Loading random article...")
	}
	if m.err != nil {
		return lipgloss.NewStyle().Padding(1, 2).
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}
	if m.random == nil {
		return lipgloss.NewStyle().Padding(1, 2).Render("Press [n] to fetch a random article")
	}

	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render(m.random.Title)
	b.WriteString(title)
	b.WriteString("\n\n")

	width := m.width - 6
	if width <= 0 {
		width = 80
	}
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n\n")

	extract := wrapText(m.random.Extract, width)
	b.WriteString(extract)
	b.WriteString("\n\n")

	url := lipgloss.NewStyle().Faint(true).Render(m.random.URL)
	b.WriteString(url)
	b.WriteString("\n\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	help := lipgloss.NewStyle().Faint(true).Render("  [↑] Back to menu  [↓] New random  [Enter] Read full article")
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m *model) searchView() string {
	var b strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("Search")
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.searchFocused {
		b.WriteString(m.searchInput.View())
	} else {
		val := m.searchInput.Value()
		if val == "" {
			b.WriteString(lipgloss.NewStyle().Faint(true).Render("/  Search Wikipedia..."))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(val))
		}
	}
	b.WriteString("\n\n")
	help := lipgloss.NewStyle().Faint(true).Render("  [↑/↓] Navigate  [Enter] Search/Open  [Esc] Back")
	b.WriteString(help)
	b.WriteString("\n\n")

	if m.searchLoading {
		b.WriteString("Searching...")
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	if m.err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Error: %v", m.err)))
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	if len(m.searchResults) == 0 {
		if m.searchInput.Value() != "" {
			b.WriteString(lipgloss.NewStyle().Faint(true).Render("No results found"))
		}
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	width := m.width - 6
	if width <= 0 {
		width = 80
	}
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n\n")

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("227"))
	dimStyle := lipgloss.NewStyle().Faint(true)
	titleStyle := lipgloss.NewStyle().Bold(true)

	for i, r := range m.searchResults {
		prefix := "  "
		if i == m.searchCursor && !m.searchFocused {
			prefix = cursorStyle.Render("▸ ")
		}
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, titleStyle.Render(r.Title)))
		if r.Snippet != "" {
			snippet := truncateText(r.Snippet, 100)
			b.WriteString(fmt.Sprintf("   %s\n", dimStyle.Render(snippet)))
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m *model) searchActionsView() string {
	searchView := m.searchView()

	actions := []string{"View Summary", "Read Full Article"}

	var d strings.Builder
	d.WriteString("\n\n")
	d.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("227")).Render("Select action:"))
	d.WriteString("\n\n")

	for i, a := range actions {
		prefix := "  "
		if i == m.actionCursor {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("▸ ")
		}
		d.WriteString(fmt.Sprintf("%s%s\n", prefix, a))
	}

	d.WriteString("\n")
	d.WriteString(lipgloss.NewStyle().Faint(true).Render("  [↑/↓] Navigate  [Enter] Select  [Esc] Back"))

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Render(d.String())

	dialogWidth := 40
	if m.width > 40 {
		dialogWidth = m.width / 3
		if dialogWidth < 40 {
			dialogWidth = 40
		}
	}

	dialog = lipgloss.NewStyle().
		Width(m.width - 4).
		Align(lipgloss.Center).
		Render(dialog)

	return lipgloss.JoinVertical(lipgloss.Top, searchView, dialog)
}

func (m *model) summaryView() string {
	if m.summaryLoading {
		return lipgloss.NewStyle().Padding(1, 2).Render("Loading summary...")
	}
	if m.err != nil {
		return lipgloss.NewStyle().Padding(1, 2).
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}
	if m.summary == nil {
		return lipgloss.NewStyle().Padding(1, 2).Render("No summary available")
	}

	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render(m.summary.Title)
	b.WriteString(title)
	b.WriteString("\n\n")

	width := m.width - 6
	if width <= 0 {
		width = 80
	}
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n\n")

	extract := wrapText(m.summary.Extract, width)
	b.WriteString(extract)
	b.WriteString("\n\n")

	url := lipgloss.NewStyle().Faint(true).Render(m.summary.URL)
	b.WriteString(url)
	b.WriteString("\n\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	help := lipgloss.NewStyle().Faint(true).Render("  [Enter] Read full article  [Esc] Back")
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m *model) readView() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render(m.readTitle)
	b.WriteString(title)
	b.WriteString("\n\n")

	content := m.readViewport.View()
	b.WriteString(content)
	b.WriteString("\n\n")

	help := lipgloss.NewStyle().Faint(true).
		Render("  [↑/↓] Scroll  [Esc] Back to menu")
	b.WriteString(help)

	full := lipgloss.NewStyle().Padding(1, 2).Render(b.String())

	if m.width > 0 {
		full = lipgloss.NewStyle().Width(m.width - 4).Render(full)
	}

	return full
}

func (m *model) mcpSettingsView() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("MCP Server Settings")
	b.WriteString(title)
	b.WriteString("\n\n")

	width := m.width - 6
	if width <= 0 {
		width = 80
	}
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n\n")

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	prefix := "  "
	if m.mcpCursor == 0 {
		prefix = cursorStyle.Render("▸ ")
	}
	langLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("227")).Render("Language:")
	langVal := m.mcpLang
	b.WriteString(fmt.Sprintf("  %s%s  %s\n\n", prefix, langLabel, langVal))

	prefix = "  "
	if m.mcpCursor == 1 {
		prefix = cursorStyle.Render("▸ ")
	}
	statusLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("227")).Render("Status:")
	statusVal := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Stopped")
	if m.mcpRunning {
		statusVal = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("Running")
	}
	b.WriteString(fmt.Sprintf("  %s%s  %s\n\n", prefix, statusLabel, statusVal))

	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n\n")

	prefix = "  "
	if m.mcpCursor == 2 {
		prefix = cursorStyle.Render("▸ ")
	}
	b.WriteString(fmt.Sprintf("  %sBack\n\n", prefix))

	help := lipgloss.NewStyle().Faint(true).
		Render("  [↑/↓] Navigate  [Enter] Toggle/Select  [Esc] Back")
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func fetchRandom(client *wikipedia.Client, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		summary, err := client.Random(ctx)
		if err != nil {
			return errMsg{err}
		}
		return randomMsg{summary}
	}
}

func performSearch(client *wikipedia.Client, ctx context.Context, query string, limit int) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Search(ctx, query, limit)
		if err != nil {
			return errMsg{err}
		}
		return searchMsg{results}
	}
}

func fetchSummary(client *wikipedia.Client, ctx context.Context, title string) tea.Cmd {
	return func() tea.Msg {
		summary, err := client.GetSummary(ctx, title)
		if err != nil {
			return errMsg{err}
		}
		return summaryMsg{summary}
	}
}

func fetchPage(client *wikipedia.Client, ctx context.Context, title string) tea.Cmd {
	return func() tea.Msg {
		page, err := client.GetPage(ctx, title)
		if err != nil {
			return errMsg{err}
		}
		return pageMsg{page}
	}
}

func wrapText(s string, width int) string {
	if width <= 0 {
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
			if current > 0 && current+len(word)+1 > width {
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

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func Run(client *wikipedia.Client) error {
	m := New(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	return nil
}
