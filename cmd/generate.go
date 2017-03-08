package main

import (
	"../slagick"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func main() {
	var sets []slagick.SetS
	f, err := os.Open(slagick.JSONFILENAME)
	handleError(err, slagick.ERR_OPEN_FILE)
	defer f.Close()
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&sets)
	handleError(err, slagick.ERR_DECODE_JSON)
	db, err := sql.Open("sqlite3", slagick.DBFILENAME)
	handleError(err, slagick.ERR_OPEN_DB)
	defer db.Close()
	db.Exec("DROP TABLE cards;")
	_, err = db.Exec(`
	CREATE TABLE cards (
		Id              INTEGER PRIMARY KEY,
		Name            TEXT,
		ManaCost        TEXT,
		Cmc             REAL,
		Type            TEXT,
		Text            TEXT,
		Power           TEXT,
		Toughness       TEXT,
		Number          TEXT,
		SetCode         TEXT,
		SetGathererCode TEXT,
		SetOldCode      TEXT,
		SetReleaseDate  TEXT
	);
	`)
	handleError(err, slagick.ERR_CREATE_TABLE)
	for _, set := range sets {
		tx, err := db.Begin()
		handleError(err, slagick.ERR_BEGIN_TX)
		statement, err := tx.Prepare(`
		INSERT INTO cards (
			Name,
			ManaCost,
			Cmc,
			Type,
			Text,
			Power,
			Toughness,
			Number,
			SetCode,
			SetGathererCode,
			SetOldCode,
			SetReleaseDate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		handleError(err, slagick.ERR_PREPARE_INSERT)
		defer statement.Close()
		for _, card := range set.Cards {
			_, err := statement.Exec(
				card.Name,
				card.ManaCost,
				card.Cmc,
				card.Type,
				card.Text,
				card.Power,
				card.Toughness,
				card.Number,
				set.Code,
				set.GathererCode,
				set.OldCode,
				set.ReleaseDate)
			handleError(err, slagick.ERR_EXEC_INSERT)
		}
		tx.Commit()
	}
}

func handleError(err error, code int) {
	if err != nil {
		fmt.Println(err)
		os.Exit(code)
	}
}
