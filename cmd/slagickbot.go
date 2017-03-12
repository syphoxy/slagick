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
					if err == sql.ErrNoRows {
						msg = "Sorry, I can't find that card."
					} else {
						api.PostMessage(bot.Admin, "I tried satisfying _'"+fullCommand+"'_ but I received this error: ```\n"+err.Error()+"\n```", params)
						msg = "An unknown error occured. I've notified my administrator."
					}
				} else {
					if strings.ToLower(card.Name) != strings.ToLower(name) {
						msg = "Sorry, I can't find that card. Is this what you were looking for?"
					}
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
				if strings.HasPrefix(fullCommand, "update") {
					params := slack.PostMessageParameters{}
					force := false
					msg := "Updated!"
					if len(commandArgs) > 1 && commandArgs[1] == "force" {
						force = true
					}
					err := bot.UpdateDB(force)
					if err != nil {
						api.PostMessage(bot.Admin, "I tried satisfying _'"+fullCommand+"'_ but I received this error: ```\n"+err.Error()+"\n```", params)
						msg = "An unknown error occured. I've notified my administrator."
					}
					api.PostMessage(ev.Msg.Channel, msg, params)
				}
			}

			if timeout > int32(time.Now().Unix()) {
				api.MarkIMChannel(ev.Msg.Channel, ev.Msg.Timestamp)
				timeout = int32(time.Now().Unix()) + 5
			}
		}
	}
}
