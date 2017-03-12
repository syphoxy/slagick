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
	query := strings.Join([]string{
		"SELECT name, mana_cost, cmc, type, text, flavor, power, toughness, number, multiverseid",
		"FROM cards",
		"WHERE LOWER(name) LIKE LOWER($1)",
		"ORDER BY set_release_date DESC",
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
