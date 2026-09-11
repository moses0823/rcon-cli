package tui

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/gorcon/rcon"
	"github.com/gorcon/rcon-cli/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	serverScreen screen = iota
	consoleScreen
)

type Model struct {
	cfg      *config.Config
	servers  []string
	cursor   int
	screen   screen
	client   *rcon.Conn
	status   string
	input    string
	cursorAt int
	history  []string
	historyAt int
	output   []string
	width    int
	height   int
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))
)

func New(cfg *config.Config) Model {
	servers := make([]string, 0, len(*cfg))

	for name := range *cfg {
		servers = append(servers, name)
	}

	sort.Strings(servers)

	return Model{
		cfg:        cfg,
		servers:    servers,
		screen:     serverScreen,
		status:     "Select a server",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type connectedMsg struct {
	client *rcon.Conn
	name   string
}

type connectErrorMsg struct {
	err error
}

type commandResultMsg struct {
	command string
	result  string
	err     error
}

func connectCmd(session config.Session, name string) tea.Cmd {
	return func() tea.Msg {
		timeout := session.Timeout

		if timeout <= 0 {
			timeout = config.DefaultTimeout
		}

		client, err := rcon.Dial(
			session.Address,
			session.Password,
			rcon.SetDialTimeout(timeout),
			rcon.SetDeadline(timeout),
		)

		if err != nil {
			return connectErrorMsg{err: err}
		}

		return connectedMsg{
			client: client,
			name:   name,
		}
	}
}

func executeCmd(client *rcon.Conn, command string) tea.Cmd {
	return func() tea.Msg {
		result, err := client.Execute(command)

		return commandResultMsg{
			command: command,
			result:  result,
			err:     err,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case connectedMsg:
		m.client = msg.client
		m.screen = consoleScreen
		m.status = "Connected"
		m.output = append(m.output,
			successStyle.Render("● Connected to "+msg.name),
		)
		return m, nil

	case connectErrorMsg:
		m.status = "Connection failed: " + msg.err.Error()
		return m, nil

	case commandResultMsg:
		if msg.err != nil {
			m.output = append(
				m.output,
				errorStyle.Render("Error: "+msg.err.Error()),
			)
		} else {
			if msg.result == "" {
				m.output = append(
					m.output,
					normalStyle.Render("> "+msg.command),
				)
			} else {
				m.output = append(
					m.output,
					normalStyle.Render("> "+msg.command),
				)

				m.output = append(
					m.output,
					minecraftColors(strings.TrimSpace(msg.result)),
				)
			}
		}

		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		if m.client != nil {
			_ = m.client.Close()
		}

		return m, tea.Quit
	}

	if m.screen == serverScreen {
		return m.handleServerKey(key)
	}

	return m.handleConsoleKey(key)
}

func (m Model) handleServerKey(key string) (tea.Model, tea.Cmd) {
	switch key {

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.servers)-1 {
			m.cursor++
		}

	case "enter":
		if len(m.servers) == 0 {
			return m, nil
		}

		name := m.servers[m.cursor]
		session := (*m.cfg)[name]

		m.status = "Connecting to " + name + "..."

		return m, connectCmd(session, name)

	case "q", "esc":
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleConsoleKey(key string) (tea.Model, tea.Cmd) {
	switch key {

	case "esc":
		if m.client != nil {
			_ = m.client.Close()
			m.client = nil
		}

		m.screen = serverScreen
		m.status = "Select a server"
		m.output = nil
		m.input = ""
		m.cursorAt = 0
		m.historyAt = len(m.history)

	case "enter":
		if strings.TrimSpace(m.input) == "" {
			return m, nil
		}

		command := m.input

		m.history = append(m.history, command)
		m.historyAt = len(m.history)

		m.input = ""
		m.cursorAt = 0

		if m.client == nil {
			return m, nil
		}

		return m, executeCmd(m.client, command)

	case "up":
		if len(m.history) == 0 {
			return m, nil
		}

		if m.historyAt > 0 {
			m.historyAt--
		}

		m.input = m.history[m.historyAt]
		m.cursorAt = len(m.input)

	case "down":
		if len(m.history) == 0 {
			return m, nil
		}

		if m.historyAt < len(m.history)-1 {
			m.historyAt++
			m.input = m.history[m.historyAt]
			m.cursorAt = len(m.input)
		} else {
			m.historyAt = len(m.history)
			m.input = ""
			m.cursorAt = 0
		}

	case "left":
		if m.cursorAt > 0 {
			m.cursorAt--
		}

	case "right":
		if m.cursorAt < len(m.input) {
			m.cursorAt++
		}

	case "home":
		m.cursorAt = 0

	case "end":
		m.cursorAt = len(m.input)

	case "backspace":
		if m.cursorAt > 0 {
			m.input = m.input[:m.cursorAt-1] + m.input[m.cursorAt:]
			m.cursorAt--
		}

	case "delete":
		if m.cursorAt < len(m.input) {
			m.input = m.input[:m.cursorAt] + m.input[m.cursorAt+1:]
		}

	default:
		if len(key) == 1 {
			m.input = m.input[:m.cursorAt] + key + m.input[m.cursorAt:]
			m.cursorAt++
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.screen == serverScreen {
		return m.serverView()
	}

	return m.consoleView()
}

func (m Model) serverView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Minecraft RCON"))
	b.WriteString("\n\n")

	b.WriteString(normalStyle.Render("Servers"))
	b.WriteString("\n\n")

	if len(m.servers) == 0 {
		b.WriteString(errorStyle.Render("  No servers configured."))
		b.WriteString("\n")
	} else {
		for i, name := range m.servers {
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("❯ " + name))
			} else {
				b.WriteString(normalStyle.Render("  " + name))
			}

			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	if strings.Contains(m.status, "failed") {
		b.WriteString(errorStyle.Render(m.status))
	} else if strings.Contains(m.status, "Connecting") {
		b.WriteString(normalStyle.Render(m.status))
	} else {
		b.WriteString(helpStyle.Render(m.status))
	}

	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render(
		"↑ ↓ Navigate    Enter Connect    Q Quit",
	))

	return b.String()
}

func (m Model) consoleView() string {
	var b strings.Builder

	name := m.servers[m.cursor]

	b.WriteString(titleStyle.Render("● " + name))
	b.WriteString("  ")
	b.WriteString(successStyle.Render("Connected"))
	b.WriteString("\n")

	b.WriteString(helpStyle.Render(
		"Esc Back    Ctrl+C Quit",
	))
	b.WriteString("\n\n")

	maxLines := m.height - 6

	if maxLines < 1 {
		maxLines = 10
	}

	start := 0

	if len(m.output) > maxLines {
		start = len(m.output) - maxLines
	}

	for _, line := range m.output[start:] {
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	prefix := normalStyle.Render("rcon> ")

	input := m.input

	if m.cursorAt >= 0 && m.cursorAt <= len(input) {
		input =
			input[:m.cursorAt] +
				"█" +
				input[m.cursorAt:]
	}

	b.WriteString(prefix)
	b.WriteString(input)

	return b.String()
}

func minecraftColors(input string) string {
	var output strings.Builder
	var text strings.Builder

	style := lipgloss.NewStyle()

	flush := func() {
		if text.Len() == 0 {
			return
		}

		output.WriteString(style.Render(text.String()))
		text.Reset()
	}

	for i := 0; i < len(input); {
		if input[i] == '§' && i+1 < len(input) {
			flush()

			code := input[i+1]
			i += 2

			switch code {
			case '0':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("0"))
			case '1':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
			case '2':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
			case '3':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
			case '4':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
			case '5':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
			case '6':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
			case '7':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
			case '8':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			case '9':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

			case 'a':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
			case 'b':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
			case 'c':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
			case 'd':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
			case 'e':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
			case 'f':
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

			case 'l':
				style = style.Bold(true)

			case 'n':
				style = style.Underline(true)

			case 'o':
				style = style.Italic(true)

			case 'm':
				style = style.Strikethrough(true)

			case 'r':
				style = lipgloss.NewStyle()

			case 'k':
				// Minecraft obfuscated 暫不處理
			}

			continue
		}

		r, size := utf8.DecodeRuneInString(input[i:])
		text.WriteRune(r)
		i += size
	}

	flush()

	return output.String()
}

func Run(cfg *config.Config) error {
	model := New(cfg)

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	_, err := program.Run()
	return err
}