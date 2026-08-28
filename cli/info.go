package cli

import (
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"

	"github.com/hebcal/hebcal-go/dailylearning"
	"github.com/hebcal/hebcal-go/zmanim"
	"github.com/hebcal/locales"

	"github.com/chaimleib/hebcalfmt/config"
)

// InfoKeys lists types of data which can be queried with the --info CLI option.
var InfoKeys = []string{
	"cities",
	"daily-learning",
	"default-city",
	"languages",
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func infoString(key string) (string, error) {
	switch key {
	case "cities":
		return joinLines(sortedCities()), nil

	case "daily-learning":
		return joinLines(dailylearning.GetCalendars()), nil

	case "default-city":
		return config.DefaultCity, nil

	case "languages":
		return joinLines(sortedLanguages()), nil

	default:
		log.Printf("unrecognized key for --info flag: %q", key)
		log.Printf("Available options: %q", InfoKeys)
		return "", fmt.Errorf("unrecognized key for --info flag: %q", key)
	}
}

func sortedCities() []string {
	locs := zmanim.AllCities()
	cities := make([]string, 0, len(locs))
	for _, loc := range locs {
		cities = append(cities, loc.Name)
	}
	sort.Strings(cities)
	return cities
}

func sortedLanguages() []string {
	langs := slices.Clone(locales.AllLocales)
	sort.Strings(langs)
	return langs
}
