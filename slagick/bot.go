package slagick

import (
	"database/sql"
)

type Bot struct {
	Me    string
	Admin string
	DB    *sql.DB
	Token string
}
