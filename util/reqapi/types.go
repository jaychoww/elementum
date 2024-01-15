package reqapi

import (
	"time"
)

type APIIdent int

const (
	TMDBIdent APIIdent = iota
	TraktIdent
	FanArtIdent
)

type API struct {
	Ident APIIdent

	Endpoint string

	BurstRate  int
	BurstTime  time.Duration
	Concurrent int

	RetriesLeft int
}
