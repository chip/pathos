package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var duplicatePaths map[string]struct{}

// this is an enum for Go
type sessionState uint

const (
	listView sessionState = iota
	inputView
)

type savePathMsg struct {
	path   string
	cursor int
}
type deletePathMsg int
type saveShellSourceMsg struct {
	m model
}

type errMsg error

// TODO Show color legend
const listHeight = 20

var (
	titleStyle                       = lipgloss.NewStyle().MarginLeft(2)
	itemStyle                        = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle                = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")) // Xterm Orchid
	doesNotExistItemStyle            = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("160")) // Xterm Red3
	selectedAndDoesNotExistItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("160"))
	duplicateItemStyle               = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("123")) // Xterm DarkSlateGray1
	selectedAndDuplicateItemStyle    = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("123"))
	paginationStyle                  = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle                        = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle                    = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct {
	keys *keys.KeyMap
}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {

	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := string(i)

	fn := itemStyle.Render

	if !directoryExists(str) {
		fn = doesNotExistItemStyle.Render
	} else if duplicatePath(str) {
		fn = duplicateItemStyle.Render
	}

	if index == m.Index() {
		fn = func(s string) string {
			if directoryExists(s) {
				return selectedItemStyle.Render("> " + s)
			} else if duplicatePath(str) {
				return selectedAndDuplicateItemStyle.Render("> " + s)
			} else {
				return selectedAndDoesNotExistItemStyle.Render("> " + s)
			}
		}
	}

	fmt.Fprint(w, fn(str))
}

type Styles struct {
	Ellipsis lipgloss.Style

	// Styling for the short help
	ShortKey       lipgloss.Style
	ShortDesc      lipgloss.Style
	ShortSeparator lipgloss.Style

	// Styling for the full help
	FullKey       lipgloss.Style
	FullDesc      lipgloss.Style
	FullSeparator lipgloss.Style
}

func helpScreen() help.Model {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#DDDADA",
		Dark:  "#3C3C3C",
	})

	return help.Model{
		ShortSeparator: " • ",
		FullSeparator:  "    ",
		Ellipsis:       "…",
		Styles: help.Styles{
			ShortKey:       keyStyle,
			ShortDesc:      descStyle,
			ShortSeparator: sepStyle,
			Ellipsis:       sepStyle.Copy(),
			FullKey:        keyStyle.Copy(),
			FullDesc:       descStyle.Copy(),
			FullSeparator:  sepStyle.Copy(),
		},
	}
}

type model struct {
	keys KeyMap
	// help string
	help     help.Model
	list     list.Model
	items    []item
	quitting bool

	textInput      textinput.Model
	msg            tea.Msg
	err            error
	state          sessionState
	showPagination bool
}

func initialModel() model {
	ti := setupTextInput()

	items := createPaths()
	duplicatePaths = findDuplicatePaths(items)

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "PATHOS - CLI Manager of the PATH env variable"
	l.SetShowHelp(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	// l.Styles.HelpStyle = helpStyle

	m := model{
		keys: keys,
		help: helpScreen(),
		// help:           "H_E_L_P",
		list:           l,
		textInput:      ti,
		err:            nil,
		state:          listView,
		showPagination: false,
	}
	return m
}

func directoryExists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}
	return true
}

func savePathCmd(cursor int, path string) tea.Cmd {
	return func() tea.Msg {
		return savePathMsg{path: path, cursor: cursor}
	}
}

func deletePathCmd(m model, id int) tea.Cmd {
	return func() tea.Msg {
		return deletePathMsg(id)
	}
}

func saveShellSourceCmd(m model) tea.Cmd {
	return func() tea.Msg {
		return saveShellSourceMsg{m: m}
	}
}

func saveShellSource(m model) (int, error) {
	s := []string{}
	for _, listItem := range m.list.Items() {
		i, _ := listItem.(item)
		path := string(i)
		if path != "" {
			s = append(s, path)
		}
	}
	data := "export PATH=" + strings.Join(s, ":")
	filename := "pathos.sh"

	file, err := os.Create(filename)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	return file.WriteString(data)
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		// m.help.Width = msg.Width
		return m, nil

	case savePathMsg:
		m.list.InsertItem(msg.cursor, item(msg.path))
		duplicatePaths = findDuplicatePaths(m.list.Items())
		return m, nil

	case deletePathMsg:
		m.list.RemoveItem(int(msg))
		duplicatePaths = findDuplicatePaths(m.list.Items())
		return m, nil

	case saveShellSourceMsg:
		saveShellSource(m)
		return m, nil

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, keys.Enter):

			if m.state == inputView {
				text := strings.TrimSpace(m.textInput.Value())
				if text != "" {
					cursor := m.list.Cursor()
					value := m.textInput.Value()
					cmds = append(cmds, savePathCmd(cursor, value))
					m.state = listView
				}
			}

		case key.Matches(msg, keys.NewPath):
			m.state = inputView
			return m, nil

		case key.Matches(msg, keys.DeletePath):
			if m.state == listView {
				i := m.list.Index()
				cmds = append(cmds, deletePathCmd(m, i))
			}

		case key.Matches(msg, keys.SaveShellSource):
			cmds = append(cmds, saveShellSourceCmd(m))

		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}
	// Update different view states
	switch m.state {
	case inputView:
		m.textInput, cmd = m.textInput.Update(msg)
	case listView:
		m.list, cmd = m.list.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// 	helpView := m.help.View(m.keys)
// 	height := 8 - strings.Count(status, "\n") - strings.Count(helpView, "\n")
//
// 	return "\n" + status + strings.Repeat("\n", height) + helpView
// }
func (m model) View() string {
	helpView := m.help.View(m.keys)
	// helpView := m.list.FullHelp()

	switch m.state {
	case inputView:
		return m.textInput.View()
	default:
		// return m.list.View() + helpView
		// v := m.list.View()
		// fmt.Printf("View(): %+v", v)
		// _, err := tea.LogToFile("%+v", v)
		// if err != nil {
		// 	log.Fatal("Couldn't LogToFile")
		// }
		return m.list.View() + helpView
	}
}

func getPaths() []string {
	PATH := os.Getenv("PATH")
	return strings.Split(PATH, ":")
}

func createPaths() []list.Item {
	paths := getPaths()
	items := make([]list.Item, len(paths))
	for i, path := range paths {
		items[i] = item(path)
	}
	return items
}

func setupTextInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = "Enter directory: "
	ti.Placeholder = "/"
	ti.SetValue("")
	ti.Blink()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return ti
}

func duplicatePath(path string) bool {
	_, isPresent := duplicatePaths[path]
	return isPresent
}

func findDuplicatePaths(items []list.Item) map[string]struct{} {
	pathMap := make(map[string]int)
	duplicates := make(map[string]struct{})

	for _, listItem := range items {
		i, ok := listItem.(item)
		if ok {
			path := string(i)
			if value, ok := pathMap[path]; ok {
				pathMap[path] = value + 1
			} else {
				pathMap[path] = 0
			}
		}
	}
	for path, count := range pathMap {
		if count > 1 {
			duplicates[path] = struct{}{}
		}
	}
	return duplicates
}

// KeyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type KeyMap struct {
	Up              key.Binding
	Down            key.Binding
	Help            key.Binding
	Quit            key.Binding
	NewPath         key.Binding
	DeletePath      key.Binding
	SaveShellSource key.Binding
	Enter           key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	// fmt.Println("inside ShortHelp")
	return []key.Binding{k.Help, k.Quit, k.NewPath}
}

func (k KeyMap) AdditionalShortHelpKeys() []key.Binding {
	fmt.Println("inside AdditionalShortHelpKeys")
	return []key.Binding{k.NewPath, k.DeletePath, k.SaveShellSource}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.NewPath, k.DeletePath, k.SaveShellSource}, // first column
		{k.Help, k.Quit}, // second column
	}
}

var keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	NewPath: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "Add new path"),
	),
	DeletePath: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "Delete path"),
	),
	SaveShellSource: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "Save paths to shell script"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
}

func main() {
	if os.Getenv("HELP_DEBUG") != "" {
		if f, err := tea.LogToFile("debug.log", "help"); err != nil {
			fmt.Println("Couldn't open a file for logging:", err)
			os.Exit(1)
		} else {
			defer f.Close()
		}
	}

	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Printf("Could not start program :(\n%v\n", err)
		os.Exit(1)
	}
}
