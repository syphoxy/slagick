package main

import (
	slagick "./lib"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("SLAGICK_DB_CONFIG"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	timeout := int32(time.Now().Unix()) + 5

	bot := slagick.Bot{
		DB: db,
	}

	api := slack.New(os.Getenv("SLAGICK_API_TOKEN"))
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if ev.User == "" {
				continue
			}

			fullCommand := strings.ToLower(ev.Msg.Text)
			commandArgs := strings.Fields(ev.Msg.Text)
			params := slack.PostMessageParameters{
				Username:  "Slagick",
				IconEmoji: ":flower_playing_cards:",
			}

			if strings.HasPrefix(fullCommand, "show me") {
				msg := ""
				name := strings.Join(commandArgs[2:], " ")

				card, err := bot.LoadCardByName(name)
				if err != nil {
					if err == sql.ErrNoRows {
						msg = "Sorry, I can't find that card. :disappointed:"
					} else {
						api.PostMessage(bot.Admin, "I tried satisfying _'"+fullCommand+"'_ but I received this error: ```\n"+err.Error()+"\n```", params)
						msg = "An unknown error occured. I've notified my administrator. :cry:"
					}
				} else {
					if strings.ToLower(card.Name) != strings.ToLower(name) {
						msg = "Sorry, I can't find that card. Is this what you were looking for? :information_desk_person:"
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

			if strings.HasPrefix(fullCommand, "%update") {
				ignore := false
				msg := "Updated!"
				if len(commandArgs) == 3 && commandArgs[1] == "ignore" && commandArgs[2] == "cache" && ev.User == bot.Admin {
					ignore = true
				}
				err := bot.UpdateDB(ignore)
				if err != nil {
					api.PostMessage(bot.Admin, "I tried satisfying _'"+fullCommand+"'_ but I received this error: ```\n"+err.Error()+"\n```", params)
					msg = "An unknown error occured. I've notified my administrator. :cry:"
				}
				api.PostMessage(ev.Msg.Channel, msg, params)
			}

			if bot.Admin == "" {
				if bot.AuthToken == "" && strings.HasPrefix(fullCommand, "authorize me") {
					bot.AuthToken = bot.GenerateAuthToken()
					log.Println("Please use the command: authorize my token " + bot.AuthToken)
					api.PostMessage(ev.Msg.Channel, "Please check bot's output for the next step. :page_with_curl: :eyes:", params)
				}
				if bot.AuthToken != "" && strings.HasPrefix(fullCommand, "authorize my token") && len(commandArgs) > 3 && bot.AuthToken == commandArgs[3] {
					bot.Admin = ev.User
					api.PostMessage(ev.Msg.Channel, "You have been authorized! :tada:", params)
				}
			}

			if timeout > int32(time.Now().Unix()) {
				api.MarkIMChannel(ev.Msg.Channel, ev.Msg.Timestamp)
				timeout = int32(time.Now().Unix()) + 5
			}
		}
	}
}
