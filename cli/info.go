package cli

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"

	"github.com/hebcal/hebcal-go/locales"
	"github.com/hebcal/hebcal-go/zmanim"
	"github.com/spf13/pflag"

	"github.com/chaimleib/hebcalfmt/config"
)

var infoKeys = []string{
	"cities",
	"default-city",
	"languages",
}

func infoString(key string, fs *pflag.FlagSet) (string, error) {
	switch key {
	case "cities":
		return strings.Join(sortedCities(), "\n"), nil

	case "default-city":
		return config.DefaultCity, nil

	case "languages":
		return strings.Join(sortedLanguages(), "\n"), nil

	case "":
		log.Println("the --info=key option is missing a key")
		log.Printf("Available options: %q", infoKeys)
		return "", errors.New("the --info=key option is missing a key")

	default:
		log.Printf("unrecognized key for --info flag: %q", key)
		log.Printf("Available options: %q", infoKeys)
		log.Println(usage(fs))
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
