package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/leekchan/accounting"
	_ "github.com/lib/pq"
	"github.com/miguelmota/go-coinmarketcap"
	"github.com/nlopes/slack"
	"github.com/syphoxy/slagick"
)

func main() {
	dbconfig := os.Getenv("SLAGICK_DB_CONFIG")
	token := os.Getenv("SLAGICK_API_TOKEN")

	if dbconfig == "" {
		log.Fatal("SLAGICK_DB_CONFIG not set")
	}

	if token == "" {
		log.Fatal("SLAGICK_API_TOKEN not set")
	}

	db, err := sql.Open("postgres", dbconfig)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	timeout := int32(time.Now().Unix()) + 5

	bot := slagick.Bot{
		DB: db,
	}

	api := slack.New(token)

	debug, err := strconv.ParseBool(os.Getenv("SLAGICK_DEBUG"))
	if err == nil && debug {
		logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
		slack.SetLogger(logger)
		api.SetDebug(debug)
	}

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if ev.User == "" {
				break
			}

			args := strings.Fields(ev.Msg.Text)
			argc := len(args)
			cmd := strings.ToLower(strings.Join(args, " "))

			params := slack.PostMessageParameters{
				Username:  "Slagick",
				IconEmoji: ":flower_playing_cards:",
			}

			if strings.HasPrefix(cmd, slagick.CmdShow) || strings.HasPrefix(cmd, slagick.CmdShoe) {
				if argc <= 2 {
					break
				}

				msg := ""
				if strings.HasPrefix(cmd, slagick.CmdShoe) {
					msg = slagick.RespShoeMeEasterEgg
				}

				name := strings.Join(args[2:], " ")
				if card, err := bot.LoadCardByName(name); err == nil {
					if strings.ToLower(card.Name) != strings.ToLower(name) {
						msg = slagick.RespFoundFuzzy
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
					api.PostMessage(ev.Msg.Channel, msg, params)
				} else {
					if err == sql.ErrNoRows {
						msg = slagick.RespNotFound
					} else {
						api.PostMessage(bot.Admin, fmt.Sprintf(slagick.RespErrorReport, cmd, err.Error()), params)
						msg = slagick.RespUnknownError
					}
				}
			}

			if strings.HasPrefix(cmd, ".update") {
				ignore := false
				msg := "Updated!"
				if strings.HasPrefix(cmd, ".update ignore cache") && ev.User == bot.Admin {
					ignore = true
					msg = "Updated! (ignored cache)"
				}
				err := bot.UpdateDB(ignore)
				if err != nil {
					api.PostMessage(bot.Admin, fmt.Sprintf(slagick.RespErrorReport, cmd, err.Error()), params)
					msg = slagick.RespUnknownError
				}
				api.PostMessage(ev.Msg.Channel, msg, params)
			}

			if bot.Admin == "" {
				if bot.AuthToken == "" && strings.HasPrefix(cmd, ".authorize me") {
					bot.AuthToken = bot.GenerateAuthToken()
					log.Println("Please use the command: #mtg .authorize my token " + bot.AuthToken)
					api.PostMessage(ev.Msg.Channel, slagick.RespCheckBotOutput, params)
				}
				if bot.AuthToken != "" && strings.HasPrefix(cmd, ".authorize my token") && len(args) > 4 && bot.AuthToken == args[4] {
					bot.Admin = ev.User
					api.PostMessage(ev.Msg.Channel, slagick.RespAuthorizedTada, params)
				}
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
							api.PostMessage(ev.Msg.Channel, fmt.Sprintf(slagick.RespErrorReport, "["+name+"]", err.Error()), params)
						}
					}
					msg := slagick.RespNoneMentioned
					if count > 0 {
						if len(params.Attachments) == count {
							msg = slagick.RespAllMentioned
						} else if len(params.Attachments) < count {
							msg = slagick.RespSomeMentioned
						}
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
