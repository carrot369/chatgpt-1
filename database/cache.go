package database

import (
	"github.com/muesli/cache2go"
)

var GPTCache = cache2go.Cache("chatgpt")

type GPTCacheItem struct {
	ParentID string
	Prompt   string
	NowID    string
	Answer   string
}
