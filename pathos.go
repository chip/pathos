package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	// "github.com/kr/pretty"
)

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

// TODO Highlight duplicates in blue
// TODO Show color legend
// TODO Auto-removal for duplicates, non-existent paths, or both? (could be configurable)
// TODO Insert new path at specific location
// TODO Update path
// TODO Make listHeight configurable
const listHeight = 40

var (
	titleStyle                       = lipgloss.NewStyle().MarginLeft(2)
	itemStyle                        = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle                = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")) // Xterm Orchid
	doesNotExistItemStyle            = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("160")) // Xterm Red3
	selectedAndDoesNotExistItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("160")) // Xterm Orchid
	paginationStyle                  = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle                        = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
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
	}

	if index == m.Index() {
		fn = func(s string) string {
			if directoryExists(s) {
				return selectedItemStyle.Render("> " + s)
			} else {
				return selectedAndDoesNotExistItemStyle.Render("> " + s)
			}
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	items    []item
	choice   string
	quitting bool

	textInput      textinput.Model
	msg            tea.Msg
	err            error
	state          sessionState
	showPagination bool
}

type errMsg error

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
		m.list.RemoveItem(id)
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
		// log.Printf("element: %s", path)
		if path != "" {
			s = append(s, path)
		}
	}
	data := "export PATH=" + strings.Join(s, ":")
	// log.Println(data)
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
		return m, nil

	case savePathMsg:
		log.Println("savePathMsg recd:", msg, " path:", item(msg.path))
		// TODO insert at what index???
		m.list.InsertItem(msg.cursor, item(msg.path))
		return m, nil

	case deletePathMsg:
		log.Println("deletePathMsg recd", msg)
		m.list.RemoveItem(int(msg))
		return m, nil

	case saveShellSourceMsg:
		log.Println("saveShellSourceMsg recd", msg)
		saveShellSource(m)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.state == inputView {
				text := strings.TrimSpace(m.textInput.Value())
				if text != "" {
					cursor := m.list.Cursor()
					value := m.textInput.Value()
					cmds = append(cmds, savePathCmd(cursor, value))
					m.state = listView
				}

			} else {
				i, ok := m.list.SelectedItem().(item)
				log.Printf("case enter else %v, ok %v", i, ok)
				if ok {
					m.choice = string(i)
				}
			}

		case "N":
			m.state = inputView
			return m, nil

		case "D":
			log.Println("case d")
			if m.state == listView {
				i := m.list.Index()
				cmds = append(cmds, deletePathCmd(m, i))
			}

		case "S":
			log.Println("case S")
			cmds = append(cmds, saveShellSourceCmd(m))
			// return m, tea.Quit

		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}
	// Update different view states
	switch m.state {
	case inputView:
		log.Println("inputView")
		m.textInput, cmd = m.textInput.Update(msg)
	case listView:
		log.Println("listView")
		m.list, cmd = m.list.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.state {
	case inputView:
		// log.Println("inputView")
		return m.textInput.View()
	default:
		// log.Println("listView")
		return m.list.View()
	}
}

func createPaths() []list.Item {
	PATH := os.Getenv("PATH")
	paths := strings.Split(PATH, ":")
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

func initialModel() model {
	ti := setupTextInput()

	items := createPaths()
	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "PATHOS - CLI Manager of the PATH env variable"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = helpStyle

	m := model{
		list:           l,
		textInput:      ti,
		err:            nil,
		state:          listView,
		showPagination: false,
	}

	return m
}

func main() {
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
