package main

import (
	//"log"
	"os"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New(os.Getenv("SLAGICKBOT_TOKEN"))
	//logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	//slack.SetLogger(logger)
	//api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if len(ev.Msg.Text) >= 9 {
				if ev.Msg.Text[0:7] == "show me" {
					rtm.SendMessage(rtm.NewOutgoingMessage("Fetching "+ev.Msg.Text[8:], ev.Msg.Channel))
				}
			}
		}
	}
}
