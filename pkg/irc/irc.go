package irc

import (
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"log"
	"os"
	"strings"
)

type Client struct {
	Config
	conn    *twitch.Client
	msgChan chan string
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
		msgChan: make(chan string),
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
		var t string
		if c.User != "" && strings.Contains(message.Message, c.User) {
			t = fmt.Sprintf("\033[97;41m [%s]%s:[white] %s]", message.User.Color, message.User.DisplayName, message.Message)

		} else {
			t = fmt.Sprintf("[%s]%s:[white] %s", message.User.Color, message.User.DisplayName, message.Message)
		}
		irc.msgChan <- t
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

func (c *Client) Fetch() string {
	return <-c.msgChan
}
