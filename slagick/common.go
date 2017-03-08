package slagick

const (
	ERR_OPEN_FILE      = 1
	ERR_OPEN_DB        = 2
	ERR_DECODE_JSON    = 3
	ERR_CREATE_TABLE   = 4
	ERR_BEGIN_TX       = 5
	ERR_PREPARE_INSERT = 6
	ERR_EXEC_INSERT    = 7

	DBFILENAME   = "cards.db"
	JSONFILENAME = "AllSetsArray-x.json"
)

type SetS struct {
	Code         string  `json:"code"`
	GathererCode string  `json:"gathererCode"`
	OldCode      string  `json:"oldCode"`
	ReleaseDate  string  `json:"releaseDate"`
	Cards        []CardS `json:"cards"`
}

type CardS struct {
	Name      string  `json:"name"`
	ManaCost  string  `json:"manaCost"`
	Cmc       float64 `json:"cmc"`
	Type      string  `json:"type"`
	Text      string  `json:"text"`
	Power     string  `json:"power"`
	Toughness string  `json:"toughness"`
	Number    string  `json:"number"`
}
