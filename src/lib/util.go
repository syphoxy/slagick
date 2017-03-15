package slagick

import (
	"strings"
)

func parseBrackets(s string, offset int) (string, int) {
	start := strings.Index(s[offset:], "[")
	end := strings.Index(s[offset+start:], "]")
	if start == -1 || end == -1 {
		return "", -1
	}
	s = strings.Replace(s[offset+start+1:offset+start+end], "[", "", -1)
	return s, offset + start + end
}

func ParseCardMentions(input string) []string {
	count := 0
	start_count := strings.Count(input, "[")
	end_count := strings.Count(input, "]")
	if start_count <= end_count {
		count = start_count
	} else {
		count = end_count
	}
	names := make([]string, 0, count)
	offset := 0
	for i := 0; i < count; i++ {
		var name string
		name, offset = parseBrackets(input, offset)
		if offset == -1 {
			break
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}
