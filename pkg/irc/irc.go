package irc

import (
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/gempir/go-twitch-irc/v4"
)

type PrivateMessageMsg struct {
	User      string
	UserColor string
	Message   string
}
type Client struct {
	Config
	conn    *twitch.Client
	MsgChan chan tea.Msg
}
type Config struct {
	User    string
	Channel string
	Auth    bool
}

func New(c Config) Client {
	var client *twitch.Client
	irc := Client{
		Config:  c,
		MsgChan: make(chan tea.Msg, 200),
	}
	if c.Auth {
		if c.User == "" {
			log.Fatal("user is missing, set TWUSER env var or set -user param")
		}
		client = twitch.NewClient(c.User, os.Getenv("TT"))
	} else {
		client = twitch.NewAnonymousClient()
	}

	if c.Channel == "" {
		log.Fatal("channel is missing, set TWUSER env var or set -channel param")
	}

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		var msg PrivateMessageMsg
		if c.User != "" && strings.Contains(message.Message, c.User) {
			msg = PrivateMessageMsg{
				UserColor: message.User.Color,
				User:      message.User.DisplayName,
				Message:   message.Message,
			}
		} else {
			msg = PrivateMessageMsg{
				UserColor: message.User.Color,
				User:      message.User.DisplayName,
				Message:   message.Message,
			}
		}

		irc.MsgChan <- msg
	})

	client.Join(c.Channel)
	go func() {
		err := client.Connect()
		if err != nil {
			log.Fatal(err)
		}
	}()
	irc.conn = client

	return irc
}

func (c *Client) Send(msg string) {
	c.conn.Say(c.Channel, msg)
}
