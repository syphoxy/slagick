package main

import (
	"database/sql"
	"log"
	"strings"

	slagick "./lib"
	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "user=postgres dbname=postgres host=localhost port=5432 sslmode=disable")

	bot := slagick.Bot{
		DB: db,
	}

	discord, err := discordgo.New("Bot MjkyODI4NTkzNzgyNzg0MDA4.C69tsQ.rYZZdB0p6fv_ck55uQ8ffFHSkFI")
	if err != nil {
		log.Println(err)
		return
	}

	discord.Open()

	u, err := discord.User("@me")
	if err != nil {
		log.Println("error obtaining account details:", err)
	}

	SelfID := u.ID

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		log.Printf("%+v %+v\n", s, (*m).Message)

		if m.Author.ID == SelfID {
			return
		}

		fullCommand := strings.ToLower(m.Content)
		commandArgs := strings.Fields(m.Content)

		if strings.HasPrefix(fullCommand, "show me") && len(commandArgs) > 2 {
			cardName := strings.Join(commandArgs[2:], " ")
			card, _ := bot.LoadCardByName(cardName)
			_, _ = s.ChannelMessageSend(m.ChannelID, card.RenderSlackMsg())
		}
	})

	<-make(chan struct{})
	return
}
