package reqapi

import (
	"github.com/elgatito/elementum/util"
)

type APIIdent int

const (
	TMDBIdent APIIdent = iota
	TraktIdent
	FanArtIdent
)

type API struct {
	Ident       APIIdent
	RateLimiter *util.RateLimiter
	Endpoint    string
	RetriesLeft int
}
