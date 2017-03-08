package main

import (
	"../slagick"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gin-gonic/gin.v1"
	"math"
	"net/http"
	"net/url"
	"strings"
)

type PayloadS struct {
	Text string `json:"text",omitempty`
}

func getCardOutput(name string) (string, error) {
	db, err := sql.Open("sqlite3", slagick.DBFILENAME)
	if err != nil {
		return "", err
	}
	var card slagick.CardS
	var set slagick.SetS
	err = db.QueryRow("SELECT Name, ManaCost, Cmc, Type, Text, Power, Toughness, Number, SetCode FROM cards WHERE Name LIKE ? ORDER BY SetReleaseDate DESC LIMIT 1", name).Scan(
		&card.Name,
		&card.ManaCost,
		&card.Cmc,
		&card.Type,
		&card.Text,
		&card.Power,
		&card.Toughness,
		&card.Number,
		&set.Code)
	if err != nil {
		return "", err
	}
	output := fmt.Sprintf(">*<http://magiccards.info/%s/en/%s.html|%s>*\n", (&url.URL{Path: strings.ToLower(set.Code)}).EscapedPath(), (&url.URL{Path: card.Number}).EscapedPath(), card.Name)
	output += fmt.Sprintf(">%s", card.Type)
	if card.Power != "" && card.Toughness != "" {
		output += fmt.Sprintf(" %s/%s", card.Power, card.Toughness)
	}
	output += fmt.Sprintf(", %s ", card.ManaCost) // TODO apply emoji filter
	if card.Cmc > math.Ceil(card.Cmc) {
		output += fmt.Sprintf("(%f)", card.Cmc)
	} else {
		output += fmt.Sprintf("(%d)", int(card.Cmc))
	}
	output += "\n"
	output += ">\n"
	for _, line := range strings.Split(card.Text, "\n") {
		output += fmt.Sprintf(">_%s_\n", line) // TODO apply emoji filter
	}
	output += ">\n"
	output += fmt.Sprintf(">http://magiccards.info/scans/en/%s/%s.jpg\n", (&url.URL{Path: strings.ToLower(set.Code)}).EscapedPath(), (&url.URL{Path: strings.ToLower(card.Number)}).EscapedPath())
	return output, nil
}

func main() {
	router := gin.Default()
	router.POST("/magic", func(c *gin.Context) {
		go func(name string) {
			output, err := getCardOutput(name)
			switch {
			case err == sql.ErrNoRows:
				// TODO send this over the other thing
				c.String(http.StatusNotFound, "No such card found!")
				return
			case err != nil:
				// TODO send this over the other thing
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			bytes, _ := json.Marshal(PayloadS{
				Text: output,
			})
			url := os.Getenv("SLAGICK_MTG_HOOK")
			http.PostForm(url, url.Values{"payload": {string(bytes)}})
		}(c.PostForm("text"))
		// purposefully send nothing back
		c.String(http.StatusOK, "")
	})
	router.Run(":9999")
}
