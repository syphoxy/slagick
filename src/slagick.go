package main

import (
	slagick "./lib"
	"database/sql"
	"fmt"
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

			if (strings.HasPrefix(fullCommand, "show me") || strings.HasPrefix(fullCommand, "shoe me")) && len(commandArgs) > 2 {
				msg := ""
				if strings.HasPrefix(fullCommand, "shoe me") {
					msg = slagick.SHOE_ME_EASTER_EGG
				}
				name := strings.Join(commandArgs[2:], " ")
				card, err := bot.LoadCardByName(name)
				if err != nil {
					if err == sql.ErrNoRows {
						msg = slagick.NOT_FOUND
					} else {
						api.PostMessage(bot.Admin, fmt.Sprintf(slagick.ERROR_REPORT, fullCommand, err.Error()), params)
						msg = slagick.UNKNOWN_ERROR
					}
				} else {
					if strings.ToLower(card.Name) != strings.ToLower(name) {
						msg = slagick.FOUND_FUZZY
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

			if strings.ContainsAny(ev.Msg.Text, "[]") {
				names := slagick.ParseCardMentions(ev.Msg.Text)
				count := len(names)
				if count > 0 {
					params.Attachments = make([]slack.Attachment, 0, count)
				CARDS:
					for _, name := range names {
						card, err := bot.LoadCardByName(name)
						for _, a := range params.Attachments {
							if a.Title == card.Name {
								count--
								continue CARDS
							}
						}
						if err == nil {
							params.Attachments = append(params.Attachments, slack.Attachment{
								Title:      card.Name,
								TitleLink:  card.GathererCardPageURL(),
								Text:       card.RenderSlackMsg(),
								ImageURL:   card.GathererCardImageURL(),
								MarkdownIn: []string{"text"},
							})
						} else if err == sql.ErrNoRows {
							count--
						} else {
							api.PostMessage(ev.Msg.Channel, fmt.Sprintf(slagick.ERROR_REPORT, "["+name+"]", err.Error()), params)
						}
					}
					msg := slagick.NONE_MENTIONED
					if count > 0 {
						if len(params.Attachments) == count {
							msg = slagick.ALL_MENTIONED
						} else if len(params.Attachments) < count {
							msg = slagick.SOME_MENTIONED
						}
					}
					api.PostMessage(ev.Msg.Channel, msg, params)
				}
			}

			if strings.HasPrefix(fullCommand, "slagick ping") {
				api.PostMessage(ev.Msg.Channel, "pong", params)
			}

			if strings.HasPrefix(fullCommand, "slagick update") {
				ignore := false
				msg := "Updated!"
				if strings.HasPrefix(fullCommand, "slagick update ignore cache") && ev.User == bot.Admin {
					ignore = true
				}
				err := bot.UpdateDB(ignore)
				if err != nil {
					api.PostMessage(bot.Admin, fmt.Sprintf(slagick.ERROR_REPORT, fullCommand, err.Error()), params)
					msg = slagick.UNKNOWN_ERROR
				}
				api.PostMessage(ev.Msg.Channel, msg, params)
			}

			if bot.Admin == "" {
				if bot.AuthToken == "" && strings.HasPrefix(fullCommand, "slagick authorize me") {
					bot.AuthToken = bot.GenerateAuthToken()
					log.Println("Please use the command: slagick authorize my token " + bot.AuthToken)
					api.PostMessage(ev.Msg.Channel, slagick.CHECK_BOT_OUTPUT, params)
				}
				if bot.AuthToken != "" && strings.HasPrefix(fullCommand, "slagick authorize my token") && len(commandArgs) > 4 && bot.AuthToken == commandArgs[4] {
					bot.Admin = ev.User
					api.PostMessage(ev.Msg.Channel, slagick.AUTHORIZED_TADA, params)
				}
			}

			if timeout > int32(time.Now().Unix()) {
				api.MarkIMChannel(ev.Msg.Channel, ev.Msg.Timestamp)
				timeout = int32(time.Now().Unix()) + 5
			}
		}
	}
}
