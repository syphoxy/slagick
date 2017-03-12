package main

import (
	"../slagick"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	db, err := sql.Open("postgres", "user=postgres dbname=postgres host=localhost port=5432 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	timeout := int32(time.Now().Unix()) + 5

	bot := slagick.Bot{
		Admin: os.Getenv("SLAGICK_BOT_ADMIN"),
		Token: os.Getenv("SLAGICKBOT_TOKEN"),
		DB:    db,
	}

	api := slack.New(os.Getenv("SLAGICKBOT_TOKEN"))
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			bot.Me = ev.Info.User.ID

		case *slack.MessageEvent:
			if ev.User == bot.Me {
				continue
			}

			fullCommand := strings.ToLower(ev.Msg.Text)
			commandArgs := strings.Fields(ev.Msg.Text)

			if strings.HasPrefix(fullCommand, "show me") {
				msg := ""
				name := strings.Join(commandArgs[2:], " ")
				params := slack.PostMessageParameters{
					Username:  "Slagick",
					IconEmoji: ":flower_playing_cards:",
				}

				card, err := bot.LoadCardByName(name)
				if err != nil {
					msg = err.Error()
				} else {
					params.Attachments = []slack.Attachment{
						slack.Attachment{
							Title:      card.Name,
							TitleLink:  card.GathererCardPageURL(),
							Text:       card.RenderSlackMsg(),
							ImageURL:   card.GathererCardImageURL(),
							MarkdownIn: []string{"text"},
						},
					}
				}
				api.PostMessage(ev.Msg.Channel, msg, params)
			}

			if ev.User == bot.Admin {
				if strings.HasPrefix(fullCommand, "populate") {
					params := slack.PostMessageParameters{}
					api.PostMessage(ev.Msg.Channel, "Populating!", params)
				}
			}

			if timeout > int32(time.Now().Unix()) {
				// https://api.slack.com/methods/im.mark
				api.MarkIMChannel(ev.Msg.Channel, ev.Msg.Timestamp)
				timeout = int32(time.Now().Unix()) + 5
			}
		}
	}
}
