package main

import (
	"../slagick"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func setupTables(db *sql.DB) {
	_, err := db.Query("DROP TABLE IF EXISTS sets, cards;")
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = db.Query(`
	CREATE TABLE cards (
		id                SERIAL PRIMARY KEY,
		name              VARCHAR(255),
		mana_cost         VARCHAR(255),
		cmc               REAL,
		type              VARCHAR(255),
		text              TEXT,
		flavor            TEXT,
		power             VARCHAR(255),
		toughness         VARCHAR(255),
		number            VARCHAR(255),
		multiverseid      INTEGER,
		set_code          VARCHAR(20),
		set_gatherer_code VARCHAR(20),
		set_old_code      VARCHAR(20),
		set_release_date  VARCHAR(50)
	);
	`)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func bootstrapTables(db *sql.DB) {
	f, err := os.Open("AllSetsArray-x.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var sets []slagick.SetS
	err = decoder.Decode(&sets)
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, set := range sets {
		tx, err := db.Begin()
		if err != nil {
			fmt.Println(err.Error())
		}
		statement, err := tx.Prepare(`
		INSERT INTO cards (
			name,
			mana_cost,
			cmc,
			type,
			text,
			flavor,
			power,
			toughness,
			number,
			multiverseid,
			set_code,
			set_gatherer_code,
			set_old_code,
			set_release_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer statement.Close()
		for _, card := range set.Cards {
			fmt.Printf(".")
			_, err = statement.Exec(
				card.Name,
				card.ManaCost,
				card.Cmc,
				card.Type,
				card.Text,
				card.Flavor,
				card.Power,
				card.Toughness,
				card.Number,
				card.MultiverseID,
				set.Code,
				set.GathererCode,
				set.OldCode,
				set.ReleaseDate)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		tx.Commit()
	}
	fmt.Println("done.")
}

func addEmoji(input string) string {
	return input
}

func getCardOutput(db *sql.DB, name string) (slack.PostMessageParameters, error) {
	var card slagick.CardS
	var set slagick.SetS
	err := db.QueryRow(`
	SELECT
		name,
		mana_cost,
		cmc,
		type,
		text,
		flavor,
		power,
		toughness,
		number,
		multiverseid,
		set_code
	FROM cards WHERE LOWER(name) LIKE LOWER($1) ORDER BY set_release_date DESC LIMIT 1
	`, name).Scan(
		&card.Name,
		&card.ManaCost,
		&card.Cmc,
		&card.Type,
		&card.Text,
		&card.Flavor,
		&card.Power,
		&card.Toughness,
		&card.Number,
		&card.MultiverseID,
		&set.Code)
	if err != nil {
		return slack.PostMessageParameters{}, err
	}
	output := card.Type
	if card.Power == "*" {
		card.Power = "★"
	}
	if card.Toughness == "*" {
		card.Toughness = "★"
	}
	if card.Power != "" && card.Toughness != "" {
		output += " " + card.Power + "/" + card.Toughness
	}
	if card.ManaCost != "" {
		output += ", " + addEmoji(card.ManaCost) + " "
		if card.Cmc > math.Ceil(card.Cmc) {
			output += fmt.Sprintf("(%f)", card.Cmc)
		} else {
			output += fmt.Sprintf("(%d)", int(card.Cmc))
		}
	}
	output += "\n"
	if card.Text != "" {
		output += "\n"
		for _, line := range strings.Split(card.Text, "\n") {
			output += addEmoji(line) + "\n"
		}
	}
	if card.Flavor != "" {
		output += "\n"
		for _, line := range strings.Split(card.Flavor, "\n") {
			output += "_" + line + "_\n"
		}
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			slack.Attachment{
				AuthorName: "Gatherer",
				AuthorLink: "http://gatherer.wizards.com",
				Title:      card.Name,
				TitleLink:  "http://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=" + strconv.Itoa(card.MultiverseID),
				Text:       output,
				ImageURL:   "http://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=" + strconv.Itoa(card.MultiverseID) + "&type=card",
				MarkdownIn: []string{"text"},
			},
		},
	}
	return params, nil
}

func main() {
	api := slack.New(os.Getenv("SLAGICKBOT_TOKEN"))
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	botID := ""
	botAdminID := os.Getenv("SLAGICK_BOT_ADMIN")

	db, err := sql.Open("postgres", "user=postgres dbname=postgres host=localhost port=5432 sslmode=disable")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			botID = ev.Info.User.ID

		case *slack.MessageEvent:
			if ev.User == botID {
				continue
			}

			// https://api.slack.com/methods/im.mark
			// api.MarkIMChannel(ev.Msg.Channel, ev.Msg.Timestamp)
			if strings.Compare(ev.Msg.Timestamp, strconv.FormatInt(time.Now().Unix(), 10)) <= 0 {
				continue
			}

			commandArgs := strings.Fields(ev.Msg.Text)
			commandLength := len(commandArgs)

			if commandLength == 1 {
				if ev.User == botAdminID {
					if commandArgs[0] == "disconnect" || commandArgs[0] == "leave" {
						rtm.SendMessage(rtm.NewOutgoingMessage("Bye!", ev.Msg.Channel))
					}
					if commandArgs[0] == "generate" {
						setupTables(db)
						bootstrapTables(db)
					}
				}
			}

			if commandLength > 2 {
				if commandArgs[0] == "show" && commandArgs[1] == "me" {
					name := strings.Join(commandArgs[2:], " ")
					params, err := getCardOutput(db, name)
					msg := "Here you go!"
					if err != nil {
						msg = err.Error()
					}
					api.PostMessage(ev.Msg.Channel, msg, params)
				}
			}

			if commandLength == 1 {
				if ev.User == botAdminID {
					if commandArgs[0] == "leave" {
						api.LeaveChannel(ev.Msg.Channel)
					}
					if commandArgs[0] == "disconnect" {
						rtm.Disconnect()
						os.Exit(0)
					}
				}
			}
		}
	}
}
