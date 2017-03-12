package slagick

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func (b Bot) createTables() error {
	_, err := b.DB.Query(`
	CREATE TABLE IF NOT EXISTS version (
		name              VARCHAR(255),
		value             VARCHAR(255)
	)
	CREATE TABLE IF NOT EXISTS cards (
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
		set_release_date  VARCHAR(50)
	);
	`)
	return err
}

func (b Bot) populateTables() error {
	var sets []SetS
	theirVersion := ""
	ourVersion := ""

	b.createTables()

	resp, err := http.Get("https://mtgjson.com/json/version.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&theirVersion)
	if err != nil {
		return err
	}

	err = b.DB.QueryRow("SELECT value FROM version WHERE name = $1", "version").Scan(&ourVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = b.DB.Exec("INSERT INTO version (name, value) VALUES ($1, $2);", "version", theirVersion)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if theirVersion == ourVersion {
		return nil
	}

	_, err = b.DB.Exec("UPDATE version SET value = $2 WHERE name = $1 LIMIT 1", "version", theirVersion)
	if err != nil {
		return err
	}

	resp, err = http.Get("https://mtgjson.com/json/AllSetsArray-x.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&sets)
	if err != nil {
		return err
	}

	for _, set := range sets {
		tx, err := b.DB.Begin()
		if err != nil {
			return err
		}

		stmt, err := tx.Prepare(`
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
			set_release_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, card := range set.Cards {
			_, err = stmt.Exec(
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
				set.ReleaseDate)
			if err != nil {
				return err
			}
		}
		tx.Commit()
	}
	return nil
}
