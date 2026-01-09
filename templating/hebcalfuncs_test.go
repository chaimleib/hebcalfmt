package templating_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/dafyomi"
	"github.com/hebcal/hebcal-go/event"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/mishnayomi"
	"github.com/hebcal/hebcal-go/sedra"

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
			if !slices.Equal(c.Want, got) {
				t.Errorf(
					"matched renderings did not match - want:\n%s\ngot:\n%s",
					strings.Join(c.Want, "\n"),
					strings.Join(got, "\n"),
				)
			}
		})
	}
}
