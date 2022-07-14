package main

import (
	"bufio"

	// "github.com/pkg/errors"

	// "gorm.io/gorm"
	// "gorm.io/driver/sqlite"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	// "github.com/spf13/viper"
	// "github.com/kr/pretty"
	// github.com/davecgh/go-spew/spew
)

var db *gorm.DB

type Path struct {
	gorm.Model
	// Index int
	Name string
}

var path Path
var paths []Path

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
type updatePathListMsg int

// TODO Highlight duplicates in blue
// TODO Highlight non-existent paths in red
// TODO Show color legend
// TODO Auto-removal for duplicates, non-existent paths, or both? (could be configurable)
// TODO Insert new path at specific location
// TODO Update path
// TODO Save to colon-delimited file pathos.env (proper name?)
// TODO Make listHeight configurable
const listHeight = 40

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
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
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
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

func savePathCmd(cursor int, path string) tea.Cmd {
	return func() tea.Msg {
		return savePathMsg{path: path, cursor: cursor}
	}
}

func deletePathCmd(id int) tea.Cmd {
	return func() tea.Msg {
		result := db.Delete(&path, id)
		if result.Error != nil {
			return errMsg(result.Error)
		}
		return deletePathMsg(id)
	}
}

func pathsToItems() []list.Item {
	result := db.Find(&paths)
	if result.Error != nil {
		log.Println("Unable to convert paths to items for display")
		log.Fatal(result.Error)
	}
	items := make([]list.Item, len(paths))
	for i, path := range paths {
		items[i] = list.Item(item(path.Name))

	}
	return items
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// log.Printf("Update() msg: %#v\n", msg)
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// tea.KeyMsg.String()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case updatePathListMsg:
		items := pathsToItems()
		m.list.SetItems(items)
		m.state = listView

	case savePathMsg:
		log.Println("SavePathMsg recd", msg)
		// if s != "" {
		// TODO insert at what index???
		m.list.InsertItem(msg.cursor, item(msg.path))
		return m, nil

	case deletePathMsg:
		log.Println("deletePathMsg recd", msg)
		// m.deletePath(int(msg))
		m.list.RemoveItem(int(msg))
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "Z":
			result := db.Find(&paths)
			if result.Error != nil {
				log.Println(result.Error)
			}
			log.Println("# paths:", len(paths))

		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.state == inputView {
				if m.textInput.Value() == "" {
					log.Println("Please enter a pathname")
				} else {
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
				cmds = append(cmds, deletePathCmd(i))
			}
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

func createPaths(db *gorm.DB) {
	PATH := os.Getenv("PATH")
	for _, name := range strings.Split(PATH, ":") {
		path := Path{Name: name}
		result := db.Create(&path)
		if result.Error != nil {
			log.Printf("oops! %+v", result.Error)
		}
	}
}

func initialModel() model {
	// Setup textinput
	ti := textinput.New()
	ti.Prompt = "Enter directory: "
	ti.Placeholder = "/"
	ti.SetValue("")
	ti.Blink()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	items := pathsToItems()
	// TODO Make configurable
	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "PATHOS - CLI Manager of the PATH env variable"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	// l.Styles.PaginationStyle = paginationStyle
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

func NewPathPrompt() string {
	var text string
	fmt.Println("What directory would you like added to your PATH?")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text = scanner.Text()
	return text
}

func createLog() {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
}

func main() {
	// createLog()
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// DB init
	db = openDB()
	createPaths(db)

	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func openDB() *gorm.DB {
	// db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		// panic("failed to connect database")
		log.Fatalf("unable to open in-memory SQLite DB: %v", err)
	}

	// Migrate the schema
	db.AutoMigrate(&Path{})
	return db
}
