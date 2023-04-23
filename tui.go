package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var m = initialModel()
var p = tea.NewProgram(m)

type (
	errMsg error
	logMsg Log
)

type LogType int

const (
	Info LogType = iota
	Warning
	Error
)

type Log struct {
	Msg  string
	Type LogType
}

type model struct {
	viewport viewport.Model
	textarea textarea.Model
	messages []string
	err      error
}

var (
	BoldStyle    = lipgloss.NewStyle().Bold(true)
	InfoLogStyle = BoldStyle.Copy().
			Foreground(lipgloss.Color("#a6da95"))
	WarningLogStyle = BoldStyle.Copy().
			Foreground(lipgloss.Color("#eed49f"))
	ErrorLogStyle = BoldStyle.Copy().
			Foreground(lipgloss.Color("#ed8796"))
	UserStyle  = BoldStyle.Copy().Foreground(lipgloss.Color("#c6a0f6"))
	ClydeStyle = BoldStyle.Copy().Foreground(lipgloss.Color("#8aadf4"))
)

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Ask Clyde anything here"
	ta.Focus()

	ta.Prompt = UserStyle.Render("❯ ")
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(1) //TODO: Change height when pasting big chunks of text, also allow multiline input

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 3)
	vp.SetContent("Type a prompt and press Enter to ask Clyde AI.")

	return model{
		textarea: ta,
		messages: []string{},
		viewport: vp,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg) //FIX: Type assert here and dont run these 2 lines for DiscordMessage & LogMsg

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Height = msg.Height - 2
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			prompt := m.textarea.Value()

			go AskClyde(prompt)

			m.messages = append(m.messages, UserStyle.Render("You: ")+prompt)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case errMsg:
		m.err = msg
		return m, nil

	case DiscordMessage:
		m.messages = append(m.messages, ClydeStyle.Render("Clyde: ")+msg.Content) //TODO: Use glamour to render codeblocks, markdown etc
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		return m, nil

	case logMsg:
		switch msg.Type {
		case Info:
			m.messages = append(m.messages, InfoLogStyle.Render("System: ")+msg.Msg)
		case Warning:
			m.messages = append(m.messages, WarningLogStyle.Render("System: ")+msg.Msg)
		case Error:
			m.messages = append(m.messages, ErrorLogStyle.Render("System: ")+msg.Msg)
		}

		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}

func RunTUI() {
	if err := p.Start(); err != nil {
		panic(err)
	}
}
