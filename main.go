package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	// "../lib/pathos/tui"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// this is an enum for Go
type sessionState uint

const (
	listView sessionState = iota
	inputView
)

// TODO Create pathos.yaml upon initialization
// TODO Highlight duplicates in blue
// TODO Highlight non-existent paths in red
// TODO Show color legend
// TODO Auto-removal for duplicates, non-existent paths, or both? (could be configurable)
// TODO Insert new path at specific location
// TODO Update path
// TODO Convert pathos.yaml to colon-delimited file pathos.env (proper name?)

// TODO Make listHeight configurable
const listHeight = 14

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

	str := fmt.Sprintf("%d. %s", index+1, i)

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

	textInput textinput.Model
	msg       tea.Msg
	err       error
	state     sessionState
}

type errMsg error

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// fmt.Printf("Update() msg: %#v\n", msg)
	var cmd tea.Cmd
	var cmds []tea.Cmd
	// tea.KeyMsg.String()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		// if m.state == listView {
		// 	fmt.Println("case tea.KeyMsg m.state: listView")
		// }
		// if m.state == inputView {
		// 	fmt.Println("case tea.KeyMsg m.state: inputView")
		// }
		// if m.textInput.Focused() {
		// 	fmt.Println("focused")
		// }
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.state == inputView {
				// fmt.Println("enter inputView")
				// i, ok := m.list.SelectedItem().(item)
				// if ok {
				// 	m.choice = string(i)
				// }
				// s := m.textInput.Value()
				// fmt.Println("[i] ", i)
				// fmt.Println("[m.choice] ", m.choice)
				cursor := m.list.Cursor()
				fmt.Println("[cursor] ", cursor)
				s := m.textInput.Value()
				// TODO insert at what index???
				m.list.InsertItem(cursor, item(s))
				// cmds = append(cmds, cmd)
				// TODO save cmd?
				savePaths(m)
				// m.list.SetItem()
				// m.list.InsertItem(i, s)
				// value := m.textInput.Value()
				// fmt.Println("text value: ", value)
				m.state = listView

			} else {
				// fmt.Println("enter listView")
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.choice = string(i)
				}
				// return m, tea.Quit
			}

		case "n":
			// if m.state == inputView {
			// fmt.Println("Update() Pressed n")
			// m.textInput, cmd = m.textInput.Update(msg)
			m.state = inputView
			// m.textInput.Placeholder = "hallo moto!"
			// m.textInput.Focus()
			// } else {
			// fmt.Println("Update() Pressed n, but skipping since in listView")
			// }
			// return m, nil

		case "d":
			if m.state == inputView {
				i := m.list.Index()
				m.list.RemoveItem(i)
				// TODO savePaths()
				savePaths(m)
			}
			// return m, nil

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
	// fmt.Println("View")
	// s := "View() choice: %#v, msg: %#v, err: %#v, state: %+v\n"
	// fmt.Printf(s, m.choice, m.msg, m.err, m.state)

	// type model struct {
	// 	list     list.Model
	// 	items    []item
	// 	choice   string
	// 	quitting bool
	//
	// 	textInput textinput.Model
	// 	msg       tea.Msg
	// 	err       error
	// 	state     sessionState
	// }
	// fmt.Println(pretty.Formatter(m))

	switch m.state {
	case inputView:
		// fmt.Println("inputView")
		return m.textInput.View()
	default:
		// fmt.Println("listView")
		return m.list.View()
	}
	// fmt.Println("m.textInput.Value")
	// fmt.Println(m.textInput.Value())

	// if m.textInput.Value() == "n" {
	// 	// fmt.Println("caught n")
	// 	// return fmt.Sprintf("PROMPT>>> %v", m.textInput.Prompt)
	// 	return fmt.Sprintf(
	// 		"What’s your favorite Pokémon?\n\n%s\n\n%s",
	// 		m.textInput.Prompt,
	// 		"(esc to quit)",
	// 	) + "\n"
	// }
	// fmt.Println("m.choice")
	// fmt.Println(m.choice)
	// // fmt.Printf("View() m.textInput %+v", m.textInput)
	// if m.choice != "" {
	// 	i, ok := m.list.SelectedItem().(item)
	// 	// m.list.Item
	// 	// return quitTextStyle.Render(fmt.Sprintf("i=%s, %s, ok = %v Sounds good to me.", i, string(i), ok))
	// 	return quitTextStyle.Render(fmt.Sprintf("i=%s ok=%v m.choice=%s", i, ok, m.choice))
	// }
	// if m.quitting {
	// 	return quitTextStyle.Render("Quitting... Thanks for using pathos.")
	// }
	// // b := &strings.Builder{}
	// // b.WriteString("Enter directory:\n")
	// // // render the text input.  All we need to do to show the full
	// // // input is call View() and return the string.
	// // b.WriteString(m.textInput.View())
	// // return b.String()
	//
	// return "\n" + m.list.View()
}

func getPaths() []string {
	path := os.Getenv("PATH") // TODO read from config file if it exists
	paths := strings.Split(path, ":")
	// fmt.Println("paths:", paths)
	return paths
}

func initialModel() model {
	// Setup textinput
	ti := textinput.New()
	ti.Prompt = "Enter directory: "
	ti.Placeholder = "/usr/bin"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	paths := getPaths()
	// TODO How to extend length?
	items := make([]list.Item, len(paths))
	for i, v := range paths {
		items[i] = item(v)
	}
	// TODO Make configurable
	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "PATHOS - CLI Manager of the PATH env variable"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{
		list:      l,
		textInput: ti,
		err:       nil,
		state:     listView,
	}

	return m
}

func savePaths(m model) (bool, error) {
	viper.Set("paths", m.list.Items())
	viper.WriteConfig()

	return true, nil
}

// TODO Write new file if none exits
func createConfig() (bool, error) {
	paths := getPaths()

	viper.SetConfigName("pathos") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.Set("paths", paths)
	viper.WriteConfig()
	// fmt.Println("after WriteConfig")

	return true, nil
}

func NewPathPrompt() string {
	var path string
	fmt.Println("What directory would you like added to your PATH?")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	path = scanner.Text()
	return path
}

func main() {
	createConfig()

	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// func (m model) Init() tea.Cmd {
//         return textinput.Blink
// }
//
// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//         var cmd tea.Cmd
//
//         switch msg := msg.(type) {
//         case tea.KeyMsg:
//                 switch msg.Type {
//                 case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
//                         return m, tea.Quit
//                 }
//
//         // We handle errors just like any other message
//         case errMsg:
//                 m.err = msg
//                 return m, nil
//         }
//
//         m.textInput, cmd = m.textInput.Update(msg)
//         return m, cmd
// }
//
// func (m model) View() string {
//         return fmt.Sprintf(
//                 "What’s your favorite Pokémon?\n\n%s\n\n%s",
//                 m.textInput.View(),
//                 "(esc to quit)",
//         ) + "\n"
// }
