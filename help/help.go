package help

import (
	"fmt"
	"strings"

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
	fmt.Println("help.go New()")
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
			{Key: "N", Description: "Create new directory"},
			{Key: "D", Description: "Delete directory"},
			{Key: "S", Description: "Save PATH"},
			{Key: "j/up", Description: "Move up"},
			{Key: "k/down", Description: "Move down"},
			{Key: "d", Description: "Move down to next page"},
			{Key: "u", Description: "Move up to previous page"},
			{Key: "g", Description: "Jump to top"},
			{Key: "G", Description: "Jump to bottom"},
			{Key: "?", Description: "Toggle Help"},
			{Key: "ctrl+c, esc, q", Description: "Quit"},
		},
	)
	fmt.Println("help.go return Bubble{help: % +v}", helpModel)

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
	fmt.Printf("help Update\n")
	// fmt.Printf("b.help.Active before: %t\n", b.help.Active)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		fmt.Println("tea.KeyMsg msg", msg)
	default:
		fmt.Println("default msg", msg)

	}

	b.help.SetIsActive(!b.help.Active)
	fmt.Printf("b.help.Active active: %t\n", b.help.Active)
	return b, nil
	// var (
	// 	cmd  tea.Cmd
	// 	cmds []tea.Cmd
	// )
	//
	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	// 	b.help.SetSize(msg.Width, msg.Height)
	// case tea.KeyMsg:
	// 	switch msg.String() {
	// 	case "ctrl+c", "esc", "q":
	// 		cmds = append(cmds, tea.Quit)
	// 	}
	// }
	//
	// b.help, cmd = b.help.Update(msg)
	// cmds = append(cmds, cmd)
	//
	// return b, tea.Batch(cmds...)
}

// View renders the UI.
func (b Bubble) View() string {
	// fmt.Println("help.go View()")
	// fmt.Printf("b.help: %+v", b.help)
	fmt.Printf("b.help.Active: %t\n", b.help.Active)
	var sb strings.Builder
	for _, value := range b.help.Entries {
		// fmt.Printf("key: %d, value: %s", key, value)
		fmt.Fprintf(&sb, "%s => %s\n", value.Key, value.Description)
	}
	return sb.String()
	// return b.help.Viewport.View()
}

// func main() {
// 	b := New()
// 	p := tea.NewProgram(b, tea.WithAltScreen())
//
// 	if err := p.Start(); err != nil {
// 		log.Fatal(err)
// 	}
// }
