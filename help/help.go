package help

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/knipferrc/teacup/help"
)

// Bubble represents the properties of the UI.
type Bubble struct {
	help help.Bubble
}

// New create a new instance of the UI.
func New() Bubble {
	helpModel := help.New(
		true,
		true,
		"Help",
		help.TitleColor{
			Background: lipgloss.AdaptiveColor{Light: "62", Dark: "62"},
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffffs"},
		},
		lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"},
		[]help.Entry{
			{Key: "ctrl+c", Description: "Exit FM"},
			{Key: "j/up", Description: "Move up"},
			{Key: "k/down", Description: "Move down"},
			{Key: "h/left", Description: "Go back a directory"},
			{Key: "l/right", Description: "Read file or enter directory"},
			{Key: "p", Description: "Preview directory"},
			{Key: "G", Description: "Jump to bottom"},
			{Key: "~", Description: "Go to home directory"},
			{Key: ".", Description: "Toggle hidden files"},
			{Key: "y", Description: "Copy file path to clipboard"},
			{Key: "Z", Description: "Zip currently selected tree item"},
			{Key: "U", Description: "Unzip currently selected tree item"},
			{Key: "n", Description: "Create new file"},
			{Key: "N", Description: "Create new directory"},
			{Key: "ctrl+d", Description: "Delete currently selected tree item"},
			{Key: "M", Description: "Move currently selected tree item"},
			{Key: "enter", Description: "Process command"},
			{Key: "E", Description: "Edit currently selected tree item"},
			{Key: "C", Description: "Copy currently selected tree item"},
			{Key: "esc", Description: "Reset FM to initial state"},
			{Key: "tab", Description: "Toggle between boxes"},
		},
	)

	return Bubble{
		help: helpModel,
	}
}

// Init initializes the application.
func (b Bubble) Init() tea.Cmd {
	return nil
}

// Update handles all UI interactions.
func (b Bubble) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.help.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)
		}
	}

	b.help, cmd = b.help.Update(msg)
	cmds = append(cmds, cmd)

	return b, tea.Batch(cmds...)
}

// View renders the UI.
func (b Bubble) View() string {
	return b.help.View()
}

func main() {
	b := New()
	p := tea.NewProgram(b, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
