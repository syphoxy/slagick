package slagick

import (
	"database/sql"
	"math/rand"
	"time"
)

type Bot struct {
	Admin     string
	DB        *sql.DB
	Token     string
	AuthToken string
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func (b Bot) randStringBytesMaskImprSrc(n int, src rand.Source) string {
	res := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			res[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(res)
}

func (b Bot) GenerateAuthToken() string {
	src := rand.NewSource(time.Now().UnixNano())
	return b.randStringBytesMaskImprSrc(20, src)
}
