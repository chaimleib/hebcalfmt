package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/locales"
	"github.com/hebcal/hebcal-go/yerushalmi"
	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/daterange"
	"github.com/chaimleib/hebcalfmt/hcfiles"
)

// ErrUnreachable means that there is a coding defect if returned.
var ErrUnreachable = errors.New("unreachable code")

// Config defines the format of the config.json file.
// It also stores program values generated at runtime;
// these are annotated with `json:"-"`
// to distinguish them from other fields expected in the JSON.
type Config struct {
	// ConfigSource names the file which produced this struct.
	ConfigSource string `json:"-"`

	// DateRange specifies the span of the calendrical data
	// which hebcal should produce.
	// This normally gets parsed from CLI arguments.
	DateRange *daterange.DateRange `json:"-"`

	// Now ensures that all parts of the calendar program
	// have a consistent idea of the current time.
	Now time.Time `json:"-"`

	// FS controls where secondary files are loaded from.
	// If replaced, it can allow access to internet-hosted files
	// compiled-in resources, or stubbing out the default FS for testing.
	// If nil, use [os.DirFS] starting from the current working directory.
	FS fs.FS `json:"-"`

	// Language sets the output language.
	// Available options are in locales.AllLocales.
	// Default: en
	Language string `json:"language"`

	// City sets geographical coordinates and a timezone for zmanim.
	// Available options are in [zmanim.AllCities].
	//
	// If no such city is in the internal database, we will error
	// unless `Geo` and `Timezone` are both set.
	// If those are set, `City` will be used as the name for the city.
	// Default: [DefaultCity]
	City string `json:"city"`

	// Geo specifies geographic coordinates for calculating zmanim.
	// This may be left empty if a known City is specified or defaulted..
	// If provided, a Timezone must also be set.
	Geo *Coordinates `json:"geo"`

	// Timezone is the name of a time zone in /usr/share/zoneinfo/
	// (on typical POSIX systems).
	// This may be left empty if a known City is specified or defaulted.
	// If provided, Geo must also be set.
	Timezone string `json:"timezone"`

	// Shiurim lists daily learning schedules to be displayed.
	// Avalable options:
	//
	// - `daf-yomi`
	// - `mishna-yomi`
	// - `nach-yomi`
	// - `yerushalmi` (defaults to Vilna edition)
	// - `yerushalmi:vilna`
	// - `yerushalmi:schottenstein`
	Shiurim []string `json:"shiurim"`

	// Today makes the hebcal calendar functions only list information
	// about today.
	// Implies Omer, AddHebrewDates, and !IsHebrewYear.
	Today bool `json:"today"`

	// ChagOnly filters output events to only show holidays and their endings,
	// during which melacha is prohibited.
	// The event bitmask is set to
	// CHAG | LIGHT_CANDLES | LIGHT_CANDLES_TZEIS | YOM_TOV_ENDS.
	ChagOnly bool `json:"chag_only"`

	// NoJulian disables the use of the Julian calendar for dates before 1752.
	// I.e. use the proleptic Gregorian calendar before then,
	// pretending that the Gregorian calendar existed
	// before it was historically used.
	NoJulian bool `json:"no_julian"`

	// Hour24 makes TimedEvent.Render() return 24-hour time.
	Hour24 bool `json:"hour24"`

	// SunriseSunset adds sunrise and sunset events for every day.
	SunriseSunset bool `json:"sunrise_sunset"`

	// CandleLighting adds entries for candlelighting times.
	CandleLighting bool `json:"candle_lighting"`

	// DailyZmanim adds zmanim events for every day.
	DailyZmanim bool `json:"daily_zmanim"`

	// Molad adds a molad entry on Shabbat Mevorchim.
	Molad bool `json:"molad"`

	// WeeklyAbbreviated gives a weekly view.
	// Omer, dafyomi, and non-date-specific zmanim are shown once a week,
	// on the day of the week which corresponds to the first day in the range.
	WeeklyAbbreviated bool `json:"weekly_abbreviated"`

	// AddHebrewDates adds an entry for the Hebrew date for each day.
	AddHebrewDates bool `json:"add_hebrew_dates"`

	// AddHebrewDatesForEvents adds an entry for the Hebrew date
	// on days with some event.
	AddHebrewDatesForEvents bool `json:"add_hebrew_dates_for_events"`

	// IsHebrewYear means to use Hebrew date ranges.
	IsHebrewYear bool `json:"is_hebrew_year"`
	// YomKippurKatan includes Yom Kippur Katan, a minor day of atonement
	// occurring monthly on the day preceding each Rosh Chodesh.
	YomKippurKatan bool `json:"yom_kippur_katan"`

	// ShabbatMevarchim include Shabbat Mevarchim HaChodesh
	// on the Shabbat before Rosh Chodesh.
	ShabbatMevarchim bool `json:"shabbat_mevarchim"`

	// NoHolidays suppresses default holidays.
	NoHolidays bool `json:"no_holidays"`

	// NoRoshChodesh suppresses Rosh Chodesh.
	NoRoshChodesh bool `json:"no_rosh_chodesh"`

	// IL uses the Israeli holiday and sedra schedule.
	IL bool `json:"il"`

	// NoModern suppresses modern holidays.
	NoModern bool `json:"no_modern"`

	// NoMinorFast suppresses minor fast days.
	NoMinorFast bool `json:"no_minor_fast"`

	// NoSpecialShabbat suppresses special Shabbatot.
	NoSpecialShabbat bool `json:"no_special_shabbat"`

	// Omer adds days of the Omer.
	Omer bool `json:"omer"`

	// Sedrot adds the weekly sedra on Saturdays.
	Sedrot bool `json:"sedrot"`

	// DailySedra adds the weekly sedra to every day.
	DailySedra bool `json:"daily_sedra"`

	// CandleLightingMins sets candle-lighting to occur
	// this many minutes before sundown.
	CandleLightingMins int `json:"candle_lighting_mins"`

	// HavdalahMins sets havdalah to occur this many minutes after sundown.
	HavdalahMins int `json:"havdalah_mins"`

	// HavdalahDeg sets havdalah to occur when the sun is this many degrees
	// below the horizon.
	HavdalahDeg float64 `json:"havdalah_deg"`

	// NumYears is how many years to generate events for.
	// Default: 1
	NumYears int `json:"num_years"`

	// EventsFile is a file of user-defined events.
	// Each line in the file has this format:
	//
	//   MMMM DD Description
	//
	// where MMMM is a string identifying the Hebrew month
	// and DD is a day number 1 through 30.
	// Description is a newline-terminated string.
	// Events are shown regardless of NoHolidays.
	EventsFile string `json:"events_file"`

	// YahrzeitsFile is a file of yartzeit dates.
	// Each line is a death-date with this format:
	//
	//   MM DD YYYY Description
	//
	// Where MM, DD and YYYY are the Gregorian date of death.
	// Description is a newline-terminated string.
	// Events are shown regardless of NoHolidays.
	YahrzeitsFile string `json:"yahrzeits_file"`
}

// Default holds the default values for [Config].
// It imitates hebcal.
var Default = Config{
	CandleLightingMins: 18,
	NumYears:           1,
}

// FromFile parses the file at `configPath` into a [Config] as JSON.
// For the sake of debugging,
// it then populates `ConfigSource` with `configPath`.
// NOTE: This does not populate DateRange.
func FromFile(configPath string) (*Config, error) {
	fileSystem, err := DefaultFS()
	if err != nil {
		slog.Error("failed to initialize DefaultFS", "error", err)
		return nil, fmt.Errorf("failed to initialize DefaultFS: %w", err)
	}

	f, err := fileSystem.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file could not be read: %w", err)
	}
	defer f.Close()

	return FromReader(f, configPath)
}

// FromReader parses an [io.Reader] into a [Config] as JSON.
// For the sake of debugging,
// it then populates `ConfigSource` with `configPath`.
func FromReader(r io.Reader, configPath string) (*Config, error) {
	cfg := Default
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, fmt.Errorf(
			"failed to parse config from %q: %w",
			configPath,
			err,
		)
	}

	cfg.ConfigSource = configPath

	return &cfg, nil
}

// Normalize returns a version of itself with canonicalized values.
//
// # `Language`
// An error gets returned if the selected `Language` is unknown.
// Otherwise, it defaults the `Language` field if unset or lowercases it,
func (c Config) Normalize() (*Config, error) {
	// work on a copy
	result := c

	// Language
	lang, err := NormalizeLanguage(c.Language)
	if err != nil {
		return nil, err
	}
	result.Language = lang

	return &result, nil
}

// CalOptions builds a [hebcal.CalOptions] from a [Config].
func (c *Config) CalOptions() (*hebcal.CalOptions, error) {
	cOpts := new(hebcal.CalOptions)

	cOpts.NoJulian = c.NoJulian
	cOpts.Hour24 = c.Hour24
	cOpts.SunriseSunset = c.SunriseSunset
	cOpts.CandleLighting = c.CandleLighting
	cOpts.DailyZmanim = c.DailyZmanim
	cOpts.Molad = c.Molad
	cOpts.WeeklyAbbreviated = c.WeeklyAbbreviated
	cOpts.AddHebrewDates = c.AddHebrewDates
	cOpts.AddHebrewDatesForEvents = c.AddHebrewDatesForEvents
	cOpts.IsHebrewYear = c.IsHebrewYear
	cOpts.YomKippurKatan = c.YomKippurKatan
	cOpts.ShabbatMevarchim = c.ShabbatMevarchim
	cOpts.NoHolidays = c.NoHolidays
	cOpts.NoRoshChodesh = c.NoRoshChodesh
	cOpts.IL = c.IL
	cOpts.NoModern = c.NoModern
	cOpts.NoMinorFast = c.NoMinorFast
	cOpts.NoSpecialShabbat = c.NoSpecialShabbat
	cOpts.Omer = c.Omer
	cOpts.Sedrot = c.Sedrot
	cOpts.DailySedra = c.DailySedra
	cOpts.CandleLightingMins = c.CandleLightingMins
	cOpts.HavdalahMins = c.HavdalahMins
	cOpts.HavdalahDeg = c.HavdalahDeg
	cOpts.NumYears = c.NumYears

	// Location
	loc, err := c.Location()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve place configs: %w", err)
	}
	cOpts.Location = loc
	if c.Geo != nil || c.City != "" {
		cOpts.CandleLighting = true
	}

	if err := c.SetDateRange(cOpts); err != nil {
		return nil, err
	}

	// YerushalmiYomi, YershushalmiEdition, MishnaYomi, DafYomi, NachYomi
	if err := SetShiurim(cOpts, c.Shiurim); err != nil {
		return nil, err
	}

	// AddHebrewDates, Omer, IsHebrewYear
	if c.Today {
		SetToday(cOpts)
	}

	// Mask
	if c.ChagOnly {
		SetChagOnly(cOpts)
	}

	// HavdalahMins default
	if cOpts.CandleLighting && cOpts.HavdalahDeg == 0 && cOpts.HavdalahMins == 0 {
		cOpts.HavdalahMins = 72
	}

	// prep for accessing secondary files
	if c.FS == nil {
		c.FS, err = DefaultFS()
		if err != nil {
			slog.Error("failed to initialize DefaultFS", "error", err)
			return nil, fmt.Errorf("failed to initialize DefaultFS: %w", err)
		}
	}

	// Read secondary files
	// UserEvents
	err = ParseFile(
		c.FS,
		c.EventsFile,
		hcfiles.ParseEvents,
		&cOpts.UserEvents,
	)
	if err != nil {
		return nil, err
	}

	// Yahrzeits
	err = ParseFile(
		c.FS,
		c.YahrzeitsFile,
		hcfiles.ParseYahrzeits,
		&cOpts.Yahrzeits,
	)
	if err != nil {
		return nil, err
	}

	return cOpts, nil
}

// SetDateRange validates the `DateRange` of the [Config].
// If it is valid and consistent with the rest of the Config,
// it gets copied to the [hebcal.CalOptions],
// along with other hebcal-like settings corresponding to the `DateRange`.
// If it is unset, a default is set on the [hebcal.CalOptions],
// imitating hebcal.
//
// It reads the following fields from the Config:
//   - `DateRange`
//   - `NoJulian`
//   - `Now`
//   - `NumYears`
//   - `Today`
func (c Config) SetDateRange(cOpts *hebcal.CalOptions) error {
	dr := c.DateRange

	if dr == nil {
		cOpts.Year = c.Now.Year()
		return nil
	}

	cOpts.IsHebrewYear = dr.IsHebrewDate

	if c.NumYears < 1 {
		return fmt.Errorf("invalid num_years: %d", c.NumYears)
	}

	if c.NumYears != 1 && dr.RangeType != daterange.RangeTypeYear {
		return fmt.Errorf(
			"num_years was %d, but the parsed date range spec was %s, not just a year",
			c.NumYears,
			dr,
		)
	}

	// Sanity checks
	if c.Today {
		switch dr.RangeType {
		case daterange.RangeTypeDay, daterange.RangeTypeToday:
			break
		default:
			return fmt.Errorf(
				"today option works only with single-day calendars, but date range spec was %s",
				dr,
			)
		}
	}

	switch dr.RangeType {
	case daterange.RangeTypeDay, daterange.RangeTypeToday:
		if dr.Day == 0 {
			return fmt.Errorf(
				"range type is %s, but the date provided is missing the day of the month: %s",
				dr.RangeType,
				dr,
			)
		}
		fallthrough

	case daterange.RangeTypeMonth:
		if dr.IsHebrewDate {
			if dr.HebMonth == 0 {
				return fmt.Errorf(
					"range type is %s, but the Hebrew date provided is missing the month: %s",
					dr.RangeType,
					dr,
				)
			}
		} else {
			if dr.GregMonth == 0 {
				return fmt.Errorf(
					"range type is %s, but the Gregorian date provided is missing the month: %s",
					dr.RangeType,
					dr,
				)
			}
		}
		fallthrough

	case daterange.RangeTypeYear:
		if dr.Year == 0 {
			return fmt.Errorf(
				"range type is %s, but the date provided is missing the year: %s",
				dr.RangeType,
				dr,
			)
		}

	default:
		slog.Error("invalid RangeType value", "rangeType", dr.RangeType)
		return fmt.Errorf(
			"%w: invalid RangeType value: %v",
			ErrUnreachable,
			dr.RangeType,
		)
	}

	// Actually set up the DateRange on the CalOptions.
	switch dr.RangeType {
	case daterange.RangeTypeDay, daterange.RangeTypeToday:
		cOpts.AddHebrewDates = true
		cOpts.Start = dr.Start(c.NoJulian)
		cOpts.End = cOpts.Start

	case daterange.RangeTypeMonth:
		cOpts.Start = dr.Start(c.NoJulian)
		cOpts.End = dr.End(c.NoJulian)

	case daterange.RangeTypeYear:
		cOpts.Year = dr.Year
	}

	return nil
}

// SetShiurim reads `shiurim`
// and sets the appropriate options on the [hebcal.CalOptions].
//
// Available shiurim values permitted in the list:
//   - `yerushalmi` `yerushalmi:vilna`
//   - `yerushalmi:schottenstein`
//   - `mishna-yomi`
//   - `daf-yomi`
//   - `nach-yomi`
func SetShiurim(cOpts *hebcal.CalOptions, shiurim []string) error {
	var unknowns []string
	for _, shiur := range shiurim {
		switch shiur {
		case "yerushalmi", "yerushalmi:vilna":
			cOpts.YerushalmiYomi = true
			cOpts.YerushalmiEdition = yerushalmi.Vilna
		case "yerushalmi:schottenstein":
			cOpts.YerushalmiYomi = true
			cOpts.YerushalmiEdition = yerushalmi.Schottenstein
		case "mishna-yomi":
			cOpts.MishnaYomi = true
		case "daf-yomi":
			cOpts.DafYomi = true
		case "nach-yomi":
			cOpts.NachYomi = true
		default:
			unknowns = append(unknowns, shiur)
		}
	}

	if len(unknowns) != 0 {
		return fmt.Errorf("unrecognized item(s) in shiurim: %q", unknowns)
	}

	return nil
}

// NormalizeLanguage checks that `language` is equal to some known language
// under case folding, and returns the canonical form from [locales.AllLocales].
// If language is empty, we return the default language `"en"`.
// If no valid value is recognized, we return an error.
func NormalizeLanguage(language string) (string, error) {
	result := strings.ToLower(language)
	switch result {
	case "", "en":
		result = "en"

	default:
		// Some language codes have capitals, like "he-x-NoNikud",
		// so build a map of lowercase to canonical form.
		foldMap := make(map[string]string, len(locales.AllLocales))
		for _, l := range locales.AllLocales {
			foldMap[strings.ToLower(l)] = l
		}
		canonical, ok := foldMap[result]
		if !ok {
			log.Printf("unknown language: %q", language)
			log.Println("To show the available languages, run")
			log.Println("  hebcalfmt --info languages")
			return "", fmt.Errorf("unknown language: %q", language)
		}
		result = canonical
	}

	return result, nil
}

// Location builds data with which to calculate zmanim.
//
// If `Geo` and `Timezone` are set and valid,
// we build a new [zmanim.Location] entry from them.
// Its `CountryCode` will be "IL" if `IL` is set, otherwise "ZZ".
// If `City` is also set, we use that as the name,
// otherwise we will use "User Defined City" like in hebcal.
//
// If `City` is set and valid, we return its [zmanim.Location] entry.
// If `Timezone` is also set and valid,
// we override the timezone for that city with the one provided,
// and note the timezone modification as a suffix to its Name.
//
// If no `Geo` or `City` is set,
// use the [DefaultCity], like in hebcal.
//
// The result of this method affects the output of hebcal functions when, e.g.,
// `CandleLighting`, `SunriseSunset`, or `DailyZmanim` is set.
//
// The following fields are read from the Config:
//   - `City`
//   - `Geo`
//   - `IL`
//   - `Timezone`
func (c Config) Location() (*zmanim.Location, error) {
	if c.Timezone != "" {
		if _, err := time.LoadLocation(c.Timezone); err != nil {
			return nil, err
		}
	}

	if c.Geo != nil {
		if c.Timezone == "" {
			return nil, errors.New("geo is set, but timezone is missing")
		}

		if err := c.Geo.Validate(); err != nil {
			return nil, fmt.Errorf("invalid geo: %w", err)
		}

		// Create a location entry.
		city := "User Defined City"
		if c.City != "" {
			city = c.City
		}
		country := "ZZ"
		if c.IL {
			country = "IL"
		}
		loc := &zmanim.Location{
			Name:        city,
			CountryCode: country,
			Latitude:    c.Geo.Lat,
			Longitude:   c.Geo.Lon,
			TimeZoneId:  c.Timezone,
		}
		return loc, nil
	}

	city := c.City
	if city == "" {
		city = DefaultCity
	}

	loc := zmanim.LookupCity(city)
	if loc == nil {
		log.Printf("unknown city: %q", c.City)
		log.Println(
			"Use a nearby city; or set geo.lat, geo.lon, and timezone.",
			"To show available cities, run:",
			"  hebcalfmt --info cities",
		)
		return nil, fmt.Errorf("unknown city: %q", c.City)
	}

	if loc.TimeZoneId != c.Timezone && c.Timezone != "" {
		loc.TimeZoneId = c.Timezone
		loc.Name = fmt.Sprintf("%s (times in timezone %s)", c.City, c.Timezone)
	}

	return loc, nil
}

// SetToday sets options on [hebcal.CalOptions]
// which hebcal itself would have set
// if the -T (--today) flag were set on its CLI,
// aside from what [(Config).SetDateRange] does already.
func SetToday(cOpts *hebcal.CalOptions) {
	cOpts.AddHebrewDates = true
	cOpts.Omer = true
	cOpts.IsHebrewYear = false
}

// SetChagOnly sets options on [hebcal.CalOptions]
// which hebcal itself would have set
// if the --chag-only flag were set on its CLI.
// This limits hebcal's output to days when melacha is prohibited.
func SetChagOnly(cOpts *hebcal.CalOptions) {
	cOpts.Mask = event.CHAG | event.LIGHT_CANDLES |
		event.LIGHT_CANDLES_TZEIS | event.YOM_TOV_ENDS
}

// ParseFile opens `fpath` from `fs`, calls `parse` on the file,
// and saves the result to `target`.
func ParseFile[T any](
	fs fs.FS,
	fpath string,
	parse func(io.Reader, string) (T, error),
	target *T,
) error {
	if fpath == "" {
		return nil
	}

	f, err := fs.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()
	item, err := parse(f, fpath)
	if err != nil {
		return err
	}
	*target = item
	return nil
}
