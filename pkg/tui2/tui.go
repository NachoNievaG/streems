package tui2

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NachoNievaG/streems/pkg/irc"
	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/davecgh/go-spew/spew"
	"github.com/lucasb-eyer/go-colorful"
)

type Config struct {
	IRC     irc.Client
	LogFile *os.File
}

type WindowSizeMsg tea.WindowSizeMsg

var (
	textAreaStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	borderStyle   = textAreaStyle.Padding(1)
	headerStyle   = lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15")).Bold(true).Padding(0, 1).Height(1)
	footerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true).Padding(0, 1).Height(1)
	dividerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Padding(0, 1)
	scrollStep    = 1 // Number of lines to scroll per key press
)

type model struct {
	scrollOffset  int
	width, height int
	dump          io.Writer
	viewport      viewport.Model
	messages      []string
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	err           error
	irc           irc.Client
}

func Run(c Config) {
	var dump *os.File
	var err error
	dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		os.Exit(1)
	}
	c.LogFile = dump
	p := tea.NewProgram(
		initialModel(c),
		tea.WithColorProfile(colorprofile.ANSI),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}

func initialModel(c Config) model {
	var ta textarea.Model
	if c.IRC.Auth {
		ta = textarea.New()
		ta.Placeholder = "Send a message..."
		ta.Focus()
		ta.SetWidth(100)
		ta.SetHeight(0)
		ta.ShowLineNumbers = false
		ta.KeyMap.InsertNewline.SetEnabled(false)
	} else {
		ta = textarea.Model{CharLimit: 0}
	}

	vp := viewport.New(viewport.WithWidth(100), viewport.WithHeight(5))
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	return model{
		dump:        c.LogFile,
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		irc:         c.IRC,
	}
}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, listenToChannel(m.irc.MsgChan)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.viewport.SetWidth(msg.Width)
		if m.textarea.CharLimit != 0 {
			m.textarea.SetWidth(msg.Width)
		}
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			// Quit.
			return m, tea.Quit
		case "enter":
			// Send Message
			v := m.textarea.Value()

			if v == "" {
				return m, nil
			}

			m.irc.Send(v)

			if strings.Contains(v, "@"+m.irc.User) {
				v = "\033[48;2;139;0;0m" + v + "\033[0m"
			}
			m.messages = append(m.messages, m.senderStyle.Render(m.irc.User)+": "+v)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.calcOffset()
			return m, listenToChannel(m.irc.MsgChan)
		default:
			// Send all other keypresses to the textarea.
			var cmd tea.Cmd
			if m.textarea.CharLimit != 0 {
				m.textarea, cmd = m.textarea.Update(msg)
			}
			m.calcOffset()
			return m, tea.Batch(cmd, listenToChannel(m.irc.MsgChan))
		}

	case irc.PrivateMessageMsg:
		var color colorful.Color
		color, err := colorful.Hex(msg.UserColor)
		if err != nil {
			color, _ = colorful.Hex("#FFFFFF")
		}
		if strings.Contains(msg.Message, "@"+m.irc.User) {
			msg.Message = "\033[48;2;139;0;0m" + msg.Message + "\033[0m"
		}
		chat := ansi.Style{}.ForegroundColor(color).Bold().Styled(msg.User) + ": " + msg.Message
		m.messages = append(m.messages, chat)
		m.calcOffset()
		m.dump.Write([]byte(fmt.Sprint(m.textarea.Height())))

		return m, listenToChannel(m.irc.MsgChan)

	}
	return m, nil
}

func (m model) calcOffset() {
	// Calculate available height
	availableHeight := m.height - 4 // Header, footer, and borders
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Set scrollOffset to show the last message
	m.scrollOffset = len(m.messages) - availableHeight
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

}

func (m model) View() string {

	// Header and Footer
	header := headerStyle.Render("Twitch IRC Viewer")
	footer := footerStyle.Render("↑/↓ to scroll | q to quit")

	// Calculate visible messages
	availableHeight := m.height - 4
	if availableHeight < 0 {
		availableHeight = 0
	}
	start := m.scrollOffset
	end := start + availableHeight
	if end > len(m.messages) {
		end = len(m.messages)
	}
	messageList := strings.Join(m.messages[start:end], "\n")

	// Main content with border
	mainContent := borderStyle.Width(m.width - 4).Height(m.height - 7).Render(messageList)
	var area string
	if m.textarea.CharLimit != 0 {
		area = textAreaStyle.Width(m.width - 4).Render(m.textarea.View())
	}
	// Combine everything
	return lipgloss.JoinVertical(lipgloss.Top, header, mainContent, area, footer)
}

func listenToChannel(msgChannel chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-msgChannel
	}
}
