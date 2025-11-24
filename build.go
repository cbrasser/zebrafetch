package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

type project struct {
	name string
	path string
}

type model struct {
	projects []project
	cursor   int
	selected int
	status   string
	quitting bool
}

func initialModel() model {
	return model{
		projects: []project{
			{name: "zebracal", path: "zebracal"},
			{name: "zebratube", path: "zebratube"},
			{name: "zebrafetch", path: "zebrafetch"},
		},
		cursor:   0,
		selected: -1,
		status:   "",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.selected == -1 {
				m.selected = m.cursor
				m.status = "Building..."
				return m, m.buildProject(m.cursor)
			}

		case "esc":
			if m.selected != -1 {
				m.selected = -1
				m.status = ""
			}
		}

	case buildResultMsg:
		m.status = msg.message
		m.selected = -1 // Reset selection after build
		return m, nil
	}

	return m, nil
}

func (m model) buildProject(selectedIdx int) tea.Cmd {
	return func() tea.Msg {
		if selectedIdx < 0 || selectedIdx >= len(m.projects) {
			return buildResultMsg{success: false, message: "Invalid project selection"}
		}

		proj := m.projects[selectedIdx]

		// Build the project
		buildCmd := exec.Command("go", "build", "-o", proj.name, ".")
		buildCmd.Dir = proj.path
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			return buildResultMsg{
				success: false,
				message: fmt.Sprintf("Build failed: %v", err),
			}
		}

		// Move to Applications directory
		sourcePath := filepath.Join(proj.path, proj.name)
		destPath := filepath.Join("/Users/claudiobrasser/Applications", proj.name)

		// Remove old binary if it exists
		if _, err := os.Stat(destPath); err == nil {
			if err := os.Remove(destPath); err != nil {
				return buildResultMsg{
					success: false,
					message: fmt.Sprintf("Failed to remove old binary: %v", err),
				}
			}
		}

		// Move the new binary
		if err := os.Rename(sourcePath, destPath); err != nil {
			return buildResultMsg{
				success: false,
				message: fmt.Sprintf("Failed to move binary: %v", err),
			}
		}

		return buildResultMsg{
			success: true,
			message: fmt.Sprintf("Successfully built and moved %s to Applications", proj.name),
		}
	}
}

type buildResultMsg struct {
	success bool
	message string
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Select project to build\n\n"))

	for i, proj := range m.projects {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		style := itemStyle
		if m.cursor == i {
			style = selectedStyle
		}

		b.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(proj.name)))
	}

	b.WriteString("\n")

	if m.status != "" {
		if strings.Contains(m.status, "Successfully") {
			b.WriteString(successStyle.Render(m.status))
		} else if strings.Contains(m.status, "failed") || strings.Contains(m.status, "Failed") {
			b.WriteString(errorStyle.Render(m.status))
		} else {
			b.WriteString(m.status)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("↑/↓: navigate • enter: build • q: quit\n")

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
