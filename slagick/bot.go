package slagick

import (
	"database/sql"
	"strings"
)

type Bot struct {
	Me    string
	Admin string
	DB    *sql.DB
	Token string
}

func (b Bot) LoadCardByName(name string) (CardS, error) {
	b.DB.Exec("CREATE EXTENSION fuzzystrmatch")
	b.DB.Exec("CREATE EXTENSION pg_trgm")

	query := strings.Join([]string{
		"SELECT name, mana_cost, cmc, type, text, flavor, power, toughness, number, multiverseid",
		"FROM cards",
		"WHERE lower(name) % lower($1)",
		"ORDER BY levenshtein(lower(name), lower($1)) ASC, set_release_date DESC",
		"LIMIT 1",
	}, " ")

	var card CardS
	err := b.DB.QueryRow(query, name).Scan(
		&card.Name,
		&card.ManaCost,
		&card.Cmc,
		&card.Type,
		&card.Text,
		&card.Flavor,
		&card.Power,
		&card.Toughness,
		&card.Number,
		&card.MultiverseID)

	if err != nil {
		return card, err
	}

	return card, nil
}
