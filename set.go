package slagick

type SetS struct {
	Code         string  `json:"code"`
	GathererCode string  `json:"gathererCode"`
	OldCode      string  `json:"oldCode"`
	ReleaseDate  string  `json:"releaseDate"`
	Cards        []CardS `json:"cards"`
}
