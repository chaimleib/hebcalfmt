package templating

import (
	"fmt"
	"strings"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/sedra"
)

var SedraFuncs = map[string]any{
	"sedra":   Sedra,
	"parasha": LocalizedParasha,
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

func LocalizedParasha(hd hdate.HDate, il bool, lang string) string {
	parashat := Translate(lang, "Parashat")
	parsha := Sedra(hd.Year(), il).Lookup(hd)
	if parsha.Chag {
		return fmt.Sprintf("%s hachag", parashat) // TODO: detect which chag
	}
	return fmt.Sprintf(
		"%s %s",
		parashat,
		strings.Join(Apply(
			parsha.Name,
			func(s string) string { return Translate(lang, s) },
		), "-"),
	)
}
