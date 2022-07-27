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
const listHeight = 35

// Colors: https://www.ditig.com/256-colors-cheat-sheet
var (
	titleStyle                       = lipgloss.NewStyle().MarginLeft(2)
	itemStyle                        = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle                = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("11")) // Xterm Yellow (SYSTEM)
	doesNotExistItemStyle            = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("9"))  // Xterm Red (SYSTEM)
	selectedAndDoesNotExistItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("9"))
	duplicateItemStyle               = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("14")) // Xterm Aqua (SYSTEM)
	selectedAndDuplicateItemStyle    = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("14"))
	quitTextStyle                    = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

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

type model struct {
	keys     HelpKeyMap
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

func additionalKeys() []key.Binding {
	return []key.Binding{
		keys.NewPath,
		keys.DeletePath,
		keys.SaveShellSource,
	}
}

func initialModel() model {
	ti := setupTextInput()

	items := createPaths()
	duplicatePaths = findDuplicatePaths(items)

	const defaultWidth = 60

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "PATHOS - CLI Manager of the PATH env variable"
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	// l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	l.AdditionalFullHelpKeys = additionalKeys
	l.AdditionalShortHelpKeys = additionalKeys

	m := model{
		keys:           keys,
		help:           help.New(),
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
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
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

func (m model) View() string {
	switch m.state {
	case inputView:
		return m.textInput.View()
	default:

		// helpView := m.list.Help.ShortHelpView(keys.ShortHelp())
		// fmt.Println("^^^", m.list.Help.ShowAll)
		// if m.list.Help.ShowAll {
		// 	helpView = m.list.Help.FullHelpView(keys.FullHelp())
		// }
		return m.list.View() // + m.list.Help.View()
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

// // KeyMap defines a set of keybindings. To work for help it must satisfy
// // key.Map. It could also very easily be a map[string]key.Binding.
type HelpKeyMap struct {
	// Up              key.Binding
	// Down            key.Binding
	Help            key.Binding
	Quit            key.Binding
	NewPath         key.Binding
	DeletePath      key.Binding
	SaveShellSource key.Binding
	Enter           key.Binding
}

var keys = HelpKeyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit new path"),
	),
	NewPath: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "new"),
	),
	DeletePath: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "delete"),
	),
	SaveShellSource: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "save paths"),
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
