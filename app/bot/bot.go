package bot

import (
	"github.com/Xacnio/img-host-go/pkg/configs"
	tele "gopkg.in/telebot.v3"
)

var (
	Bot *tele.Bot
)

func Create() error {
	var err error
	Bot, err = tele.NewBot(tele.Settings{
		Token:  configs.Get("TG_BOT_TOKEN"),
		//Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	// Create bot but we won't connect and listen to updates for run multiple replicas.
	return err
}