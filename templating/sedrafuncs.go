package templating

import (
	"github.com/hebcal/hebcal-go/sedra"
)

var SedraFuncs = map[string]any{
	"sedra": Sedra,
}

var sedraCache, sedraCacheIL map[int]sedra.Sedra

// Sedra returns a sedra.Sedra
// which can be used to query for the Parasha of the week.
func Sedra(year int, il bool) *sedra.Sedra {
	if il {
		if sedraCacheIL == nil {
			sedraCacheIL = make(map[int]sedra.Sedra)
		}
	} else {
		if sedraCache == nil {
			sedraCache = make(map[int]sedra.Sedra)
		}
	}

	cache := sedraCache
	if il {
		cache = sedraCacheIL
	}

	got, ok := cache[year]
	if !ok {
		got = sedra.New(year, il)
		cache[year] = got
	}

	return &got
}
