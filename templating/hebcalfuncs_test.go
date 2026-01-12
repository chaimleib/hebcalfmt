package templating_test

import (
	"fmt"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/dafyomi"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/mishnayomi"
	"github.com/hebcal/hebcal-go/sedra"
	"github.com/hebcal/hebcal-go/zmanim"

	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestLookupCity(t *testing.T) {
	cases := []struct {
		Name string
		Err  string
	}{
		{Name: "New York"},
		{Name: "Austin"},
		{Name: "Invalid City", Err: `unknown city "Invalid City"`},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := templating.LookupCity(c.Name)
			test.CheckErr(t, err, c.Err)
			if c.Err == "" && got == nil {
				t.Error("no error expected, but got no city!")
			}
		})
	}
}

// wrapAsEvent converts the provided asEvent function
// into another function which returns an event.CalEvent.
// This type normalization helps to easily call asEvent in table-driven tests.
func wrapAsEvent[T event.CalEvent](
	asEvent func(event.CalEvent) *T,
) func(event.CalEvent) event.CalEvent {
	return func(e event.CalEvent) event.CalEvent {
		typed := asEvent(e)
		if typed == nil {
			return nil
		}
		return *typed
	}
}

func dayAt(dt time.Time, h, m int) time.Time {
	y, mo, d := dt.Date()
	return time.Date(y, mo, d, h, m, 0, 0, dt.Location())
}

func TestAsEvent(t *testing.T) {
	hd := hdate.New(5775, hdate.Kislev, 25)
	d := hd.Gregorian()
	opts := new(hebcal.CalOptions)

	events := []event.CalEvent{
		event.HolidayEvent{
			Date:        hd,
			Desc:        "Chanukah: 2 Candles",
			ChanukahDay: 1,
		},
		event.UserEvent{
			Date: hd,
			Desc: "Birthday - Joe Shmo",
		},
		event.UserEvent{
			Date: hd,
			Desc: "Yahrzeit - Jane Doe",
		},
		event.NewYerushalmiYomiEvent(hd, dafyomi.Daf{
			Name:  "Terumot",
			Blatt: 22,
		}),
		event.NewDafYomiEvent(hd, dafyomi.Daf{
			Name:  "Yevamot",
			Blatt: 74,
		}),
		event.NewMishnaYomiEvent(hd, mishnayomi.MishnaPair{
			mishnayomi.Mishna{Tractate: "Kelim", Chap: 9, Verse: 5},
			mishnayomi.Mishna{Tractate: "Kelim", Chap: 9, Verse: 6},
		}),
		event.NewNachYomiEvent(hd, dafyomi.Daf{
			Name:  "Malachi",
			Blatt: 1,
		}),
		event.NewHebrewDateEvent(hd),
		event.NewParshaEvent(hd, sedra.Parsha{
			Name: []string{"Miketz"},
			Num:  []int{9},
			Chag: false,
		}, false),
	}

	// link this event to the one at index 0 about Chanukah
	events = append(events, hebcal.NewTimedEvent(
		hd, "Chanukah: 2 Candles", 0x0, dayAt(d, 16, 53), 0, events[0], opts),
	)

	events = append(events, []event.CalEvent{
		hebcal.NewTimedEvent(
			hd, "Alot haShachar", 0x0, dayAt(d, 5, 46), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Misheyakir", 0x0, dayAt(d, 6, 11), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Misheyakir Machmir", 0x0, dayAt(d, 6, 19), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Sunrise", 0x0, dayAt(d, 7, 14), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Kriat Shema, sof zeman (MGA)", 0x0, dayAt(d, 8, 57), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Kriat Shema, sof zeman (GRA)", 0x0, dayAt(d, 9, 33), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Tefilah, sof zeman (MGA)", 0x0, dayAt(d, 9, 55), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Tefilah, sof zeman (GRA)", 0x0, dayAt(d, 10, 19), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Chatzot hayom", 0x0, dayAt(d, 11, 52), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Mincha Gedolah", 0x0, dayAt(d, 12, 15), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Mincha Ketanah", 0x0, dayAt(d, 14, 34), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Plag HaMincha", 0x0, dayAt(d, 15, 32), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Sunset", 0x0, dayAt(d, 16, 30), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Bein HaShemashot", 0x0, dayAt(d, 16, 53), 0, nil, opts),
		hebcal.NewTimedEvent(
			hd, "Tzeit HaKochavim", 0x0, dayAt(d, 17, 15), 0, nil, opts),
	}...)

	cases := []struct {
		Name string
		F    func(event.CalEvent) event.CalEvent
		Want []string
		Err  string
	}{
		{
			Name: "HolidayEvent",
			F:    wrapAsEvent(templating.AsEvent[event.HolidayEvent]),
			Want: []string{"Chanukah: 2 Candles"},
		},
		{
			Name: "UserEvent",
			F:    wrapAsEvent(templating.AsEvent[event.UserEvent]),
			Want: []string{
				"Birthday - Joe Shmo",
				"Yahrzeit - Jane Doe",
			},
		},
		{
			Name: "TimedEvent",
			F:    wrapAsEvent(templating.AsEvent[hebcal.TimedEvent]),
			Want: []string{
				"Chanukah: 2 Candles: 4:53",
				"Alot haShachar: 5:46",
				"Misheyakir: 6:11",
				"Misheyakir Machmir: 6:19",
				"Sunrise: 7:14",
				"Kriat Shema, sof zeman (MGA): 8:57",
				"Kriat Shema, sof zeman (GRA): 9:33",
				"Tefilah, sof zeman (MGA): 9:55",
				"Tefilah, sof zeman (GRA): 10:19",
				"Chatzot hayom: 11:52",
				"Mincha Gedolah: 12:15",
				"Mincha Ketanah: 2:34",
				"Plag HaMincha: 3:32",
				"Sunset: 4:30",
				"Bein HaShemashot: 4:53",
				"Tzeit HaKochavim: 5:15",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var got []string
			for _, e := range events {
				if match := c.F(e); match != nil {
					render := match.Render("en")
					got = append(got, render)
				}
			}
			test.CheckSlice(t, "matched renderings", c.Want, got)
		})
	}
}

func TestCompareTimedEvents(t *testing.T) {
	tz, err := time.LoadLocation("US/Central")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name  string
		Input []hebcal.TimedEvent
		Want  []hebcal.TimedEvent
	}{
		{Name: "empties"},
		{
			Name:  "single",
			Input: []hebcal.TimedEvent{{}},
			Want:  []hebcal.TimedEvent{{}},
		},
		{
			Name: "Havdalah in order",
			Input: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Havdalah"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Tzeit HaKochavim"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
			},
			Want: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Havdalah"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Tzeit HaKochavim"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
			},
		},
		{
			Name: "Havdalah reverse order",
			Input: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Tzeit HaKochavim"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Havdalah"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
			},
			Want: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Havdalah"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Tzeit HaKochavim"},
					EventTime:    time.Date(1970, 3, 7, 18, 30, 0, 0, tz),
				},
			},
		},
		{
			Name: "Candle lighting in order",
			Input: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Chanukah: 1 Candle"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
			},
			Want: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Chanukah: 1 Candle"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
			},
		},
		{
			Name: "Candle lighting reverse order",
			Input: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Chanukah: 1 Candle"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
			},
			Want: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Chanukah: 1 Candle"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
			},
		},
		{
			Name: "events in order",
			Input: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Sunset"},
					EventTime:    time.Date(1970, 12, 14, 16, 17, 0, 0, tz),
				},
			},
			Want: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Sunset"},
					EventTime:    time.Date(1970, 12, 14, 16, 17, 0, 0, tz),
				},
			},
		},
		{
			Name: "events reverse order",
			Input: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Sunset"},
					EventTime:    time.Date(1970, 12, 14, 16, 17, 0, 0, tz),
				},
			},
			Want: []hebcal.TimedEvent{
				{
					HolidayEvent: event.HolidayEvent{Desc: "Candle lighting"},
					EventTime:    time.Date(1970, 12, 14, 15, 59, 0, 0, tz),
				},
				{
					HolidayEvent: event.HolidayEvent{Desc: "Sunset"},
					EventTime:    time.Date(1970, 12, 14, 16, 17, 0, 0, tz),
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := slices.Clone(c.Input)
			sort.Slice(got, templating.CompareTimedEvents(got))
			if !slices.Equal(c.Want, got) {
				t.Errorf(
					"want:\n  %s\ngot:\n  %s",
					test.AsStrings(c.Want),
					test.AsStrings(got),
				)
			}
		})
	}
}

func TestTimedEvents(t *testing.T) {
	hd := hdate.New(5740, hdate.Nisan, 15)

	baseOpts := hebcal.CalOptions{
		CandleLightingMins: 18,
		NumYears:           1,
		Location:           zmanim.LookupCity("Milwaukee"),
	}

	cases := []struct {
		Name  string
		Opts  *hebcal.CalOptions
		Dates []hdate.HDate
		Err   string
		Want  []string
	}{
		{
			Name:  "too many dates",
			Opts:  new(hebcal.CalOptions),
			Dates: []hdate.HDate{hd, hd, hd},
			Err:   "expected 0-2 dates, got 3",
		},
		{
			Name: "invalid opts",
			Opts: &hebcal.CalOptions{
				CandleLighting: true,
				Location:       nil,
			},
			Err: "opts.CandleLighting requires opts.Location",
		},
		{
			Name: "default to opts year",
			Opts: func() *hebcal.CalOptions {
				opts := baseOpts
				opts.Year = 1970
				opts.CandleLighting = true
				return &opts
			}(),
			Want: []string{
				"1970-01-02 - Parashat Shemot - Candle lighting: 4:10",
				"1970-01-03 - Parashat Shemot - Havdalah: 5:16",
				"1970-01-09 - Parashat Vaera - Candle lighting: 4:17",
				"1970-01-10 - Parashat Vaera - Havdalah: 5:23",
				"1970-01-16 - Parashat Bo - Candle lighting: 4:25",
				"1970-01-17 - Parashat Bo - Havdalah: 5:30",
				"1970-01-23 - Parashat Beshalach - Candle lighting: 4:34",
				"1970-01-24 - Parashat Beshalach - Havdalah: 5:38",
				"1970-01-30 - Parashat Yitro - Candle lighting: 4:43",
				"1970-01-31 - Parashat Yitro - Havdalah: 5:47",
				"1970-02-06 - Parashat Mishpatim - Candle lighting: 4:52",
				"1970-02-07 - Parashat Mishpatim - Havdalah: 5:55",
				"1970-02-13 - Parashat Terumah - Candle lighting: 5:02",
				"1970-02-14 - Parashat Terumah - Havdalah: 6:04",
				"1970-02-20 - Parashat Tetzaveh - Candle lighting: 5:11",
				"1970-02-21 - Parashat Tetzaveh - Havdalah: 6:13",
				"1970-02-27 - Parashat Ki Tisa - Candle lighting: 5:20",
				"1970-02-28 - Parashat Ki Tisa - Havdalah: 6:21",
				"1970-03-06 - Parashat Vayakhel - Candle lighting: 5:29",
				"1970-03-07 - Parashat Vayakhel - Havdalah: 6:30",
				"1970-03-13 - Parashat Pekudei - Candle lighting: 5:37",
				"1970-03-14 - Parashat Pekudei - Havdalah: 6:38",
				"1970-03-19 - Ta'anit Esther - Fast begins: 4:31",
				"1970-03-19 - Ta'anit Esther - Fast ends: 6:36",
				"1970-03-20 - Parashat Vayikra - Candle lighting: 5:45",
				"1970-03-21 - Parashat Vayikra - Havdalah: 6:47",
				"1970-03-27 - Parashat Tzav - Candle lighting: 5:53",
				"1970-03-28 - Parashat Tzav - Havdalah: 6:55",
				"1970-04-03 - Parashat Shmini - Candle lighting: 6:01",
				"1970-04-04 - Parashat Shmini - Havdalah: 7:04",
				"1970-04-10 - Parashat Tazria - Candle lighting: 6:09",
				"1970-04-11 - Parashat Tazria - Havdalah: 7:12",
				"1970-04-17 - Parashat Metzora - Candle lighting: 6:18",
				"1970-04-18 - Parashat Metzora - Havdalah: 7:21",
				"1970-04-20 - Ta'anit Bechorot - Fast begins: 3:28",
				"1970-04-20 - Erev Pesach - Candle lighting: 6:21",
				"1970-04-21 - Pesach I - Candle lighting: 7:25",
				"1970-04-22 - Pesach II - Havdalah: 7:27",
				"1970-04-24 - Parashat  - Candle lighting: 6:26",
				"1970-04-25 - Parashat  - Havdalah: 7:31",
				"1970-04-26 - Pesach VI (CH''M) - Candle lighting: 7:28",
				"1970-04-27 - Pesach VII - Candle lighting: 8:33",
				"1970-04-28 - Pesach VIII - Havdalah: 8:35",
				"1970-05-01 - Parashat Achrei Mot - Candle lighting: 7:34",
				"1970-05-02 - Parashat Achrei Mot - Havdalah: 8:40",
				"1970-05-08 - Parashat Kedoshim - Candle lighting: 7:42",
				"1970-05-09 - Parashat Kedoshim - Havdalah: 8:49",
				"1970-05-15 - Parashat Emor - Candle lighting: 7:50",
				"1970-05-16 - Parashat Emor - Havdalah: 8:58",
				"1970-05-22 - Parashat Behar - Candle lighting: 7:57",
				"1970-05-23 - Parashat Behar - Havdalah: 9:07",
				"1970-05-29 - Parashat Bechukotai - Candle lighting: 8:03",
				"1970-05-30 - Parashat Bechukotai - Havdalah: 9:14",
				"1970-06-05 - Parashat Bamidbar - Candle lighting: 8:09",
				"1970-06-06 - Parashat Bamidbar - Havdalah: 9:20",
				"1970-06-09 - Erev Shavuot - Candle lighting: 8:11",
				"1970-06-10 - Shavuot I - Candle lighting: 9:23",
				"1970-06-11 - Shavuot II - Havdalah: 9:24",
				"1970-06-12 - Parashat Nasso - Candle lighting: 8:13",
				"1970-06-13 - Parashat Nasso - Havdalah: 9:25",
				"1970-06-19 - Parashat Beha'alotcha - Candle lighting: 8:15",
				"1970-06-20 - Parashat Beha'alotcha - Havdalah: 9:28",
				"1970-06-26 - Parashat Sh'lach - Candle lighting: 8:16",
				"1970-06-27 - Parashat Sh'lach - Havdalah: 9:28",
				"1970-07-03 - Parashat Korach - Candle lighting: 8:16",
				"1970-07-04 - Parashat Korach - Havdalah: 9:27",
				"1970-07-10 - Parashat Chukat - Candle lighting: 8:13",
				"1970-07-11 - Parashat Chukat - Havdalah: 9:23",
				"1970-07-17 - Parashat Balak - Candle lighting: 8:09",
				"1970-07-18 - Parashat Balak - Havdalah: 9:17",
				"1970-07-21 - Tzom Tammuz - Fast begins: 3:42",
				"1970-07-21 - Tzom Tammuz - Fast ends: 9:04",
				"1970-07-24 - Parashat Pinchas - Candle lighting: 8:03",
				"1970-07-25 - Parashat Pinchas - Havdalah: 9:10",
				"1970-07-31 - Parashat Matot-Masei - Candle lighting: 7:56",
				"1970-08-01 - Parashat Matot-Masei - Havdalah: 9:01",
				"1970-08-07 - Parashat Devarim - Candle lighting: 7:47",
				"1970-08-08 - Parashat Devarim - Havdalah: 8:51",
				"1970-08-10 - Erev Tish'a B'Av - Fast begins: 8:01",
				"1970-08-11 - Tish'a B'Av - Fast ends: 8:37",
				"1970-08-14 - Parashat Vaetchanan - Candle lighting: 7:37",
				"1970-08-15 - Parashat Vaetchanan - Havdalah: 8:40",
				"1970-08-21 - Parashat Eikev - Candle lighting: 7:27",
				"1970-08-22 - Parashat Eikev - Havdalah: 8:28",
				"1970-08-28 - Parashat Re'eh - Candle lighting: 7:15",
				"1970-08-29 - Parashat Re'eh - Havdalah: 8:15",
				"1970-09-04 - Parashat Shoftim - Candle lighting: 7:03",
				"1970-09-05 - Parashat Shoftim - Havdalah: 8:03",
				"1970-09-11 - Parashat Ki Teitzei - Candle lighting: 6:51",
				"1970-09-12 - Parashat Ki Teitzei - Havdalah: 7:50",
				"1970-09-18 - Parashat Ki Tavo - Candle lighting: 6:38",
				"1970-09-19 - Parashat Ki Tavo - Havdalah: 7:36",
				"1970-09-25 - Parashat Nitzavim-Vayeilech - Candle lighting: 6:25",
				"1970-09-26 - Parashat Nitzavim-Vayeilech - Havdalah: 7:23",
				"1970-09-30 - Erev Rosh Hashana - Candle lighting: 6:16",
				"1970-10-01 - Rosh Hashana 5731 - Candle lighting: 7:14",
				"1970-10-02 - Rosh Hashana II - Candle lighting: 6:13",
				"1970-10-03 - Parashat Ha'azinu - Havdalah: 7:11",
				"1970-10-04 - Tzom Gedaliah - Fast begins: 5:28",
				"1970-10-04 - Tzom Gedaliah - Fast ends: 7:01",
				"1970-10-09 - Erev Yom Kippur - Candle lighting: 6:00",
				"1970-10-10 - Yom Kippur - Havdalah: 6:59",
				"1970-10-14 - Erev Sukkot - Candle lighting: 5:52",
				"1970-10-15 - Sukkot I - Candle lighting: 6:51",
				"1970-10-16 - Sukkot II - Candle lighting: 5:49",
				"1970-10-17 - Parashat  - Havdalah: 6:48",
				"1970-10-21 - Sukkot VII (Hoshana Raba) - Candle lighting: 5:41",
				"1970-10-22 - Shmini Atzeret - Candle lighting: 6:40",
				"1970-10-23 - Simchat Torah - Candle lighting: 5:38",
				"1970-10-24 - Parashat Bereshit - Havdalah: 6:37",
				"1970-10-30 - Parashat Noach - Candle lighting: 4:28",
				"1970-10-31 - Parashat Noach - Havdalah: 5:28",
				"1970-11-06 - Parashat Lech-Lecha - Candle lighting: 4:19",
				"1970-11-07 - Parashat Lech-Lecha - Havdalah: 5:20",
				"1970-11-13 - Parashat Vayera - Candle lighting: 4:12",
				"1970-11-14 - Parashat Vayera - Havdalah: 5:14",
				"1970-11-20 - Parashat Chayei Sara - Candle lighting: 4:06",
				"1970-11-21 - Parashat Chayei Sara - Havdalah: 5:09",
				"1970-11-27 - Parashat Toldot - Candle lighting: 4:01",
				"1970-11-28 - Parashat Toldot - Havdalah: 5:05",
				"1970-12-04 - Parashat Vayetzei - Candle lighting: 3:59",
				"1970-12-05 - Parashat Vayetzei - Havdalah: 5:04",
				"1970-12-11 - Parashat Vayishlach - Candle lighting: 3:58",
				"1970-12-12 - Parashat Vayishlach - Havdalah: 5:04",
				"1970-12-18 - Parashat Vayeshev - Candle lighting: 4:00",
				"1970-12-19 - Parashat Vayeshev - Havdalah: 5:06",
				"1970-12-22 - Chanukah: 1 Candle - Chanukah: 1 Candle: 4:45",
				"1970-12-23 - Chanukah: 2 Candles - Chanukah: 2 Candles: 4:46",
				"1970-12-24 - Chanukah: 3 Candles - Chanukah: 3 Candles: 4:46",
				"1970-12-25 - Chanukah: 4 Candles - Chanukah: 4 Candles: 4:03",
				"1970-12-25 - Parashat Miketz - Candle lighting: 4:03",
				"1970-12-26 - Parashat Miketz - Havdalah: 5:10",
				"1970-12-26 - Chanukah: 5 Candles - Chanukah: 5 Candles: 5:10",
				"1970-12-27 - Chanukah: 6 Candles - Chanukah: 6 Candles: 4:48",
				"1970-12-28 - Chanukah: 7 Candles - Chanukah: 7 Candles: 4:49",
				"1970-12-29 - Chanukah: 8 Candles - Chanukah: 8 Candles: 4:50",
			},
		},
		{
			Name: "sort Candle lighting last on Chanukah",
			// Dates contains an erev Shabbos Chanukah.
			// Shabbos candles should be lit after the menorah.
			Dates: []hdate.HDate{hdate.New(5740, hdate.Kislev, 24)},
			Opts: func() *hebcal.CalOptions {
				opts := baseOpts
				opts.CandleLighting = true
				opts.Location = zmanim.LookupCity("Milwaukee")
				opts.DailyZmanim = true
				return &opts
			}(),
			Want: []string{
				"1979-12-14 - Parashat Vayeshev - Alot haShachar: 5:43",
				"1979-12-14 - Parashat Vayeshev - Misheyakir: 6:10",
				"1979-12-14 - Parashat Vayeshev - Misheyakir Machmir: 6:17",
				"1979-12-14 - Parashat Vayeshev - Sunrise: 7:15",
				"1979-12-14 - Parashat Vayeshev - Kriat Shema, sof zeman (MGA): 8:54",
				"1979-12-14 - Parashat Vayeshev - Kriat Shema, sof zeman (GRA): 9:30",
				"1979-12-14 - Parashat Vayeshev - Tefilah, sof zeman (MGA): 9:52",
				"1979-12-14 - Parashat Vayeshev - Tefilah, sof zeman (GRA): 10:16",
				"1979-12-14 - Parashat Vayeshev - Chatzot hayom: 11:46",
				"1979-12-14 - Parashat Vayeshev - Mincha Gedolah: 12:09",
				"1979-12-14 - Parashat Vayeshev - Mincha Ketanah: 2:24",
				"1979-12-14 - Parashat Vayeshev - Plag HaMincha: 3:20",
				"1979-12-14 - Chanukah: 1 Candle - Chanukah: 1 Candle: 3:59",
				"1979-12-14 - Parashat Vayeshev - Candle lighting: 3:59",
				"1979-12-14 - Parashat Vayeshev - Sunset: 4:17",
				"1979-12-14 - Parashat Vayeshev - Bein HaShemashot: 4:42",
				"1979-12-14 - Parashat Vayeshev - Tzeit HaKochavim: 5:04",
			},
		},
		{
			Name: "sort Candle lighting last Yom Tov Sheni",
			// Dates contains an erev Shabbos Chanukah.
			// Shabbos candles should be lit after the menorah.
			Dates: []hdate.HDate{hdate.New(5740, hdate.Sivan, 6)},
			Opts: func() *hebcal.CalOptions {
				opts := baseOpts
				opts.CandleLighting = true
				opts.Location = zmanim.LookupCity("Milwaukee")
				opts.DailyZmanim = true
				return &opts
			}(),
			Want: []string{
				"1980-05-21 - Parashat Nasso - Alot haShachar: 3:33",
				"1980-05-21 - Parashat Nasso - Misheyakir: 4:09",
				"1980-05-21 - Parashat Nasso - Misheyakir Machmir: 4:19",
				"1980-05-21 - Parashat Nasso - Sunrise: 5:22",
				"1980-05-21 - Parashat Nasso - Kriat Shema, sof zeman (MGA): 8:29",
				"1980-05-21 - Parashat Nasso - Kriat Shema, sof zeman (GRA): 9:05",
				"1980-05-21 - Parashat Nasso - Tefilah, sof zeman (MGA): 9:55",
				"1980-05-21 - Parashat Nasso - Tefilah, sof zeman (GRA): 10:19",
				"1980-05-21 - Parashat Nasso - Chatzot hayom: 12:48",
				"1980-05-21 - Parashat Nasso - Mincha Gedolah: 1:25",
				"1980-05-21 - Parashat Nasso - Mincha Ketanah: 5:08",
				"1980-05-21 - Parashat Nasso - Plag HaMincha: 6:41",
				"1980-05-21 - Parashat Nasso - Sunset: 8:14",
				"1980-05-21 - Parashat Nasso - Bein HaShemashot: 8:42",
				"1980-05-21 - Parashat Nasso - Tzeit HaKochavim: 9:05",
				"1980-05-21 - Shavuot I - Candle lighting: 9:05",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			events, err := templating.TimedEvents(c.Opts)(c.Dates...)
			test.CheckErr(t, err, c.Err)

			got := make([]string, 0, len(events))
			sedraCache := make(map[int]sedra.Sedra)
			for _, event := range events {
				if event.LinkedEvent == nil { // Is Shabbat or weekday event
					// Get parsha through sedraCache.
					hYear := event.Date.Year()
					yearSedra, ok := sedraCache[hYear]
					if !ok {
						yearSedra = sedra.New(hYear, c.Opts.IL)
						sedraCache[hYear] = yearSedra
					}
					parsha := yearSedra.Lookup(event.Date)

					got = append(got, fmt.Sprintf(
						"%s - %s - %s",
						event.Date.Gregorian().Format(time.DateOnly),
						parsha.String(),
						event.Render("en"),
					))
				} else {
					got = append(got, fmt.Sprintf(
						"%s - %s - %s",
						event.Date.Gregorian().Format(time.DateOnly),
						event.LinkedEvent.Render("en"),
						event.Render("en"),
					))
				}
			}
			test.CheckSlice(t, "events", c.Want, got)
		})
	}
}
