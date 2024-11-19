package tui

import (
	"fmt"
	"github.com/NachoNievaG/streems/pkg/irc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChatClient interface {
	Send(msg string)
}

type Config struct {
	IRC irc.Client
}

type UI struct {
	app *tview.Application
}

func Build(cfg Config) UI {

	app := tview.NewApplication()

	// Create a TextView that takes up most of the screen
	textView := buildTextView(cfg, app)
	// Create an InputField for capturing user input
	inputField := buildInputField(cfg, textView)
	// Create a Flex layout to place the TextView and InputField
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false) // TextView at the top, takes remaining space
	if cfg.IRC.Auth {
		flex.AddItem(inputField, 3, 0, true) // InputField at the bottom, with fixed height
	}
	app.SetRoot(flex, true)
	textView.SetBackgroundColor(tcell.Color(12))
	return UI{app}
}

func buildTextView(cfg Config, app *tview.Application) *tview.TextView {
	textView := tview.NewTextView().
		SetText("Welcome to the Chat! \n").
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetScrollable(true)
	go func() {
		for {
			if x := cfg.IRC.Fetch(); x != "" {
				redrawTextView(textView, x)
				textView.ScrollToEnd()
			}
		}
	}()
	textView.SetBorder(true).
		SetTitle(cfg.IRC.Channel + " - Chat").
		SetTitleAlign(0) // Handle mouse clicks on the TextView

	return textView

}

func buildInputField(cfg Config, textView *tview.TextView) *tview.TextArea {

	inputField := tview.NewTextArea().
		SetWrap(false).
		SetPlaceholder("Enter text here...")
	inputField.SetTitle("Text Area").SetBorder(true)

	// Handle the input field's "Enter" key press
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			inputText := inputField.GetText()

			redrawTextView(textView, fmt.Sprintf("%s: %s", cfg.IRC.User, inputText))
			cfg.IRC.Send(inputText)
			inputField.SetText("", false)
		}

		return event
	})
	inputField.SetBackgroundColor(tcell.ColorBlack).SetBorder(true).SetTitle("Send a message").SetTitleAlign(0)

	return inputField
}

func redrawTextView(t *tview.TextView, s string) {
	fmt.Fprintf(t, "%s \n", s)

}

func (ui UI) Start() {
	// Set up the application with the Flex layout as the root
	if err := ui.app.EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
