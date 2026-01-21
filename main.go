package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Bookmark represents a saved directory path
type Bookmark struct {
	Path  string `json:"path"`
	Name  string `json:"name,omitempty"`
	Count int    `json:"count"`
}

// Config holds all bookmarks
type Config struct {
	Bookmarks []Bookmark `json:"bookmarks"`
}

// Model represents the TUI state
type model struct {
	bookmarks    []Bookmark
	filtered     []int // indices into bookmarks
	cursor       int
	filter       string
	editing      bool
	editValue    string
	selectedPath string // path to cd to after quit
}

var (
	normalStyle = lipgloss.NewStyle().
			Padding(0, 1)
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	filterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
)

func renderSelected(s string) string {
	return "\x1b[7m" + s + "\x1b[0m"
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "bk", "bookmarks.json")
}

func loadConfig() Config {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{Bookmarks: []Bookmark{}}
	}
	var config Config
	json.Unmarshal(data, &config)
	return config
}

func saveConfig(config Config) error {
	configPath := getConfigPath()
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func sortBookmarks(bookmarks []Bookmark) {
	sort.Slice(bookmarks, func(i, j int) bool {
		return bookmarks[i].Count > bookmarks[j].Count
	})
}

func filterBookmarks(bookmarks []Bookmark, filter string) []int {
	if filter == "" {
		indices := make([]int, len(bookmarks))
		for i := range bookmarks {
			indices[i] = i
		}
		return indices
	}
	filter = strings.ToLower(filter)
	var indices []int
	for i, b := range bookmarks {
		name := strings.ToLower(b.Name)
		path := strings.ToLower(b.Path)
		if strings.Contains(name, filter) || strings.Contains(path, filter) {
			indices = append(indices, i)
		}
	}
	return indices
}

func initialModel() model {
	config := loadConfig()
	sortBookmarks(config.Bookmarks)
	filtered := filterBookmarks(config.Bookmarks, "")
	return model{
		bookmarks: config.Bookmarks,
		filtered:  filtered,
		cursor:    0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			switch msg.Type {
			case tea.KeyEnter:
				if len(m.filtered) > 0 {
					idx := m.filtered[m.cursor]
					m.bookmarks[idx].Name = m.editValue
					config := Config{Bookmarks: m.bookmarks}
					saveConfig(config)
				}
				m.editing = false
				m.editValue = ""
			case tea.KeyEsc:
				m.editing = false
				m.editValue = ""
			case tea.KeyBackspace:
				if len(m.editValue) > 0 {
					m.editValue = m.editValue[:len(m.editValue)-1]
				}
			case tea.KeySpace:
				m.editValue += " "
			default:
				if msg.Type == tea.KeyRunes {
					m.editValue += string(msg.Runes)
				}
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.filter != "" {
				m.filter = ""
				m.filtered = filterBookmarks(m.bookmarks, "")
				m.cursor = 0
			} else {
				return m, tea.Quit
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				idx := m.filtered[m.cursor]
				m.bookmarks[idx].Count++
				config := Config{Bookmarks: m.bookmarks}
				saveConfig(config)
				m.selectedPath = m.bookmarks[idx].Path
				return m, tea.Quit
			}
		case tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.filtered = filterBookmarks(m.bookmarks, m.filter)
				m.cursor = 0
			}
		case tea.KeyRunes:
			char := string(msg.Runes)
			// Handle special single-key commands only when not filtering
			if m.filter == "" {
				switch char {
				case "q":
					return m, tea.Quit
				case "e":
					if len(m.filtered) > 0 {
						m.editing = true
						idx := m.filtered[m.cursor]
						m.editValue = m.bookmarks[idx].Name
					}
					return m, nil
				case "d":
					if len(m.filtered) > 0 {
						idx := m.filtered[m.cursor]
						m.bookmarks = append(m.bookmarks[:idx], m.bookmarks[idx+1:]...)
						config := Config{Bookmarks: m.bookmarks}
						saveConfig(config)
						m.filtered = filterBookmarks(m.bookmarks, m.filter)
						if m.cursor >= len(m.filtered) && m.cursor > 0 {
							m.cursor--
						}
					}
					return m, nil
				case "j":
					if m.cursor < len(m.filtered)-1 {
						m.cursor++
					}
					return m, nil
				case "k":
					if m.cursor > 0 {
						m.cursor--
					}
					return m, nil
				}
			}
			// Otherwise, add to filter
			m.filter += char
			m.filtered = filterBookmarks(m.bookmarks, m.filter)
			m.cursor = 0
		}
	}
	return m, nil
}

func (m model) View() string {
	if len(m.bookmarks) == 0 {
		return "\n  No bookmarks yet. Use 'bk add' to add the current directory.\n\n  Press q to quit.\n"
	}

	s := "\n"

	// Show filter input
	if m.filter != "" {
		s += fmt.Sprintf("  %s %s\n\n", filterStyle.Render("filter:"), m.filter)
	}

	if m.editing {
		s += fmt.Sprintf("  Rename bookmark: %s\n", m.editValue)
		s += "  (Enter to save, Esc to cancel)\n\n"
	}

	if len(m.filtered) == 0 {
		s += dimStyle.Render("  No matches") + "\n"
	} else {
		for i, idx := range m.filtered {
			b := m.bookmarks[idx]
			displayName := b.Path
			if b.Name != "" {
				displayName = b.Name
			}

			if i == m.cursor {
				if b.Name != "" {
					s += fmt.Sprintf("  > %s", renderSelected(displayName))
					s += dimStyle.Render(fmt.Sprintf(" %s", b.Path))
				} else {
					s += fmt.Sprintf("  > %s", renderSelected(displayName))
				}
			} else {
				if b.Name != "" {
					s += fmt.Sprintf("    %s", displayName)
					s += dimStyle.Render(fmt.Sprintf(" %s", b.Path))
				} else {
					s += fmt.Sprintf("    %s", displayName)
				}
			}
			s += "\n"
		}
	}

	s += "\n  ↑/↓ navigate • enter select • e rename • d delete • esc clear • q quit\n"

	return s
}

func addBookmark() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	config := loadConfig()

	// Check if bookmark already exists
	for _, b := range config.Bookmarks {
		if b.Path == cwd {
			fmt.Printf("Bookmark already exists: %s\n", cwd)
			return
		}
	}

	// Prompt for alias
	fmt.Printf("Adding: %s\n", cwd)
	fmt.Print("Alias (enter to skip): ")
	reader := bufio.NewReader(os.Stdin)
	alias, _ := reader.ReadString('\n')
	alias = strings.TrimSpace(alias)

	config.Bookmarks = append(config.Bookmarks, Bookmark{
		Path:  cwd,
		Name:  alias,
		Count: 0,
	})

	if err := saveConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	if alias != "" {
		fmt.Printf("Added bookmark: %s (%s)\n", alias, cwd)
	} else {
		fmt.Printf("Added bookmark: %s\n", cwd)
	}
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "add":
			addBookmark()
			return
		case "help", "--help", "-h":
			fmt.Println("bk - directory bookmarks")
			fmt.Println("")
			fmt.Println("Usage:")
			fmt.Println("  bk        Open bookmark selector")
			fmt.Println("  bk add    Add current directory to bookmarks")
			fmt.Println("")
			fmt.Println("Keys:")
			fmt.Println("  ↑/↓, j/k  Navigate")
			fmt.Println("  Enter     Go to selected directory")
			fmt.Println("  e         Edit bookmark name")
			fmt.Println("  d         Delete bookmark")
			fmt.Println("  q         Quit")
			return
		}
	}

	// Open /dev/tty for TUI so it works even when stdout is captured
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	p := tea.NewProgram(initialModel(), tea.WithInput(tty), tea.WithOutput(tty))
	m, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print selected path to stdout (captured by shell function)
	if finalModel, ok := m.(model); ok && finalModel.selectedPath != "" {
		fmt.Println(finalModel.selectedPath)
	}
}
