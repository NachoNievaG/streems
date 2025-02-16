package main

import (
	"flag"
	"os"

	"github.com/NachoNievaG/streems/pkg/irc"
	"github.com/NachoNievaG/streems/pkg/tui2"
)

func main() {
	user := flag.String("user", os.Getenv("TWUSER"), "twitch user set in TWUSER env var")
	channel := flag.String("channel", os.Getenv("TWUSER"), "twitch channel to attach to (default: self)")
	oauth := flag.Bool("auth", false, "set to true if you want to send messages as well")
	flag.Parse()

	config := irc.Config{
		User:    *user,
		Channel: *channel,
		Auth:    *oauth,
	}

	c := irc.New(config)
	conf := tui2.Config{IRC: c}

	tui2.Run(conf)

}
