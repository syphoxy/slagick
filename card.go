package slagick

import (
	"regexp"
	"strconv"
	"strings"
)

type CardS struct {
	Name         string  `json:"name"`
	ManaCost     string  `json:"manaCost"`
	Cmc          float64 `json:"cmc"`
	Type         string  `json:"type"`
	Text         string  `json:"text"`
	Flavor       string  `json:"flavor"`
	Power        string  `json:"power"`
	Toughness    string  `json:"toughness"`
	Number       string  `json:"number"`
	MultiverseID int     `json:"multiverseid"`
}

func (c CardS) filterWrap(input, wrapper string) string {
	return wrapper + input + wrapper
}

func (c CardS) filterBold(input string) string {
	return c.filterWrap(input, "*")
}
func (c CardS) filterItalics(input string) string {
	return c.filterWrap(input, "_")
}

func (c CardS) filterManaCost(input string) string {
	// special cards:
	// - Little Girl
	// - Gleemax
	return input
}

func (c CardS) filterCmc(input float64) string {
	return strconv.FormatFloat(input, 'f', -1, 64)
}

func (c CardS) filterPower(input string) string {
	if input == "*" {
		return "\xE2\x98\x85"
	}
	float, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return input
	}
	return strconv.FormatFloat(float, 'f', -1, 64)
}

func (c CardS) filterToughness(input string) string {
	return c.filterPower(input)
}

func (c CardS) filterSlackEmojiInternal(input string) string {
	if input == "{100}" {
		return ":mtg1h1::mtg1h2:"
	}

	if input == "{1000000}" {
		return ":mtg1m1::mtg1m2::mtg1m3::mtg1m4::mtg1m5::mtg1m6:"
	}

	input = strings.Replace(input, "/", "", -1)
	input = strings.Replace(input, "\xE2\x88\x9E", "inf", -1)
	input = strings.Replace(input, "\xC2\xBD", "half", -1)
	input = strings.Replace(input, "{", ":mtg", -1)
	input = strings.Replace(input, "}", ":", -1)
	return strings.ToLower(input)
}

func (c CardS) filterSlackEmoji(input string) string {
	re := regexp.MustCompile("(?i){([^}]+)}")
	return re.ReplaceAllStringFunc(input, c.filterSlackEmojiInternal)
}

func (c CardS) GathererCardPageURL() string {
	return "http://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=" + strconv.Itoa(c.MultiverseID)
}

func (c CardS) GathererCardImageURL() string {
	return "http://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=" + strconv.Itoa(c.MultiverseID) + "&type=card"
}

func (c CardS) RenderSlackMsg() string {
	power := c.filterPower(c.Power)
	toughness := c.filterToughness(c.Toughness)

	msg := c.Type

	if power != "" && toughness != "" {
		msg += " " + power + "/" + toughness
	}

	if c.ManaCost != "" {
		msg += " " + c.filterSlackEmoji(c.ManaCost)
	}

	msg += "\n"

	if c.Text != "" {
		msg += "\n"
		for _, line := range strings.Split(c.Text, "\n") {
			msg += c.filterSlackEmoji(line) + "\n"
		}
	}

	if c.Flavor != "" {
		msg += "\n"
		for _, line := range strings.Split(c.Flavor, "\n") {
			msg += c.filterItalics(line) + "\n"
		}
	}

	return msg
}
