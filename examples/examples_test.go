package examples_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/chaimleib/hebcalfmt/cli"
	"github.com/chaimleib/hebcalfmt/fsys"
	"github.com/chaimleib/hebcalfmt/templating"
	"github.com/chaimleib/hebcalfmt/test"
)

func setupExample(t *testing.T) (fs.FS, time.Time) {
	cli.InitLogging()

	files, err := fsys.DefaultFS()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Block access to the default config inside ~/.config/... .
	// If HOME is empty, the program gives up on the default config file.
	t.Setenv("HOME", "")

	now := time.Date(2024, time.May, 6, 0, 0, 0, 0, time.UTC)

	return files, now
}

func TestExamples_Today(t *testing.T) {
	files, now := setupExample(t)
	const fpath = "today.tmpl"

	var buf bytes.Buffer
	err := cli.RunInEnvironment(
		[]string{fpath}, files, now, templating.BuildData, &buf)
	if err != nil {
		t.Error(err)
	}

	want := "Gregorian: 2024-05-06\nHebrew: 28 Nisan 5784\n"
	if buf.String() != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, buf.String())
	}
}

func TestExamples_MonthCalendar(t *testing.T) {
	files, now := setupExample(t)
	const fpath = "monthCalendar.tmpl"

	cases := []struct {
		Name string
		Args string
		Want string
		Err  string
	}{
		{
			Name: "empty args, marks today",
			Args: "",
			Want: `May 2024
Nisan - Iyyar 5784

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |       |       |  1 23 |  2 24 |  3 25 |  4 26 |
|  5 27 |* 6 28*|  7 29 |  8 30 |  9  1 | 10  2 | 11  3 |
| 12  4 | 13  5 | 14  6 | 15  7 | 16  8 | 17  9 | 18 10 |
| 19 11 | 20 12 | 21 13 | 22 14 | 23 15 | 24 16 | 25 17 |
| 26 18 | 27 19 | 28 20 | 29 21 | 30 22 | 31 23 |       |
`,
		},
		{
			Name: "Gregorian day",
			Args: "2024-02-03",
			Want: `February 2024
Sh'vat - Adar I 5784

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |       |       |       |  1 22 |  2 23 |  3 24 |
|  4 25 |  5 26 |  6 27 |  7 28 |  8 29 |  9 30 | 10  1 |
| 11  2 | 12  3 | 13  4 | 14  5 | 15  6 | 16  7 | 17  8 |
| 18  9 | 19 10 | 20 11 | 21 12 | 22 13 | 23 14 | 24 15 |
| 25 16 | 26 17 | 27 18 | 28 19 | 29 20 |       |       |
`,
		},
		{
			Name: "Hebrew day",
			Args: "Shevat 15 5762",
			Want: `January 2002
Tevet - Sh'vat 5762

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |       |  1 17 |  2 18 |  3 19 |  4 20 |  5 21 |
|  6 22 |  7 23 |  8 24 |  9 25 | 10 26 | 11 27 | 12 28 |
| 13 29 | 14  1 | 15  2 | 16  3 | 17  4 | 18  5 | 19  6 |
| 20  7 | 21  8 | 22  9 | 23 10 | 24 11 | 25 12 | 26 13 |
| 27 14 | 28 15 | 29 16 | 30 17 | 31 18 |       |       |
`,
		},
		{
			Name: "Hebrew month",
			Args: "Adar1 5763",
			Want: `February 2003
Sh'vat - Adar I 5763

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |       |       |       |       |       |  1 29 |
|  2 30 |  3  1 |  4  2 |  5  3 |  6  4 |  7  5 |  8  6 |
|  9  7 | 10  8 | 11  9 | 12 10 | 13 11 | 14 12 | 15 13 |
| 16 14 | 17 15 | 18 16 | 19 17 | 20 18 | 21 19 | 22 20 |
| 23 21 | 24 22 | 25 23 | 26 24 | 27 25 | 28 26 |       |
`,
		},
		{
			Name: "Gregorian month starting Sunday",
			Args: "10 2023",
			Want: `October 2023
Tishrei - Cheshvan 5784

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|  1 16 |  2 17 |  3 18 |  4 19 |  5 20 |  6 21 |  7 22 |
|  8 23 |  9 24 | 10 25 | 11 26 | 12 27 | 13 28 | 14 29 |
| 15 30 | 16  1 | 17  2 | 18  3 | 19  4 | 20  5 | 21  6 |
| 22  7 | 23  8 | 24  9 | 25 10 | 26 11 | 27 12 | 28 13 |
| 29 14 | 30 15 | 31 16 |       |       |       |       |
`,
		},
		{
			Name: "Gregorian month ending Shabbos, spanning Hebrew years",
			Args: "9 2023",
			Want: `September 2023
Elul 5783 - Tishrei 5784

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |       |       |       |       |  1 15 |  2 16 |
|  3 17 |  4 18 |  5 19 |  6 20 |  7 21 |  8 22 |  9 23 |
| 10 24 | 11 25 | 12 26 | 13 27 | 14 28 | 15 29 | 16  1 |
| 17  2 | 18  3 | 19  4 | 20  5 | 21  6 | 22  7 | 23  8 |
| 24  9 | 25 10 | 26 11 | 27 12 | 28 13 | 29 14 | 30 15 |
`,
		},
		{
			Name: "Gregorian month starting Shabbos",
			Args: "7 2023",
			Want: `July 2023
Tammuz - Av 5783

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |       |       |       |       |       |  1 12 |
|  2 13 |  3 14 |  4 15 |  5 16 |  6 17 |  7 18 |  8 19 |
|  9 20 | 10 21 | 11 22 | 12 23 | 13 24 | 14 25 | 15 26 |
| 16 27 | 17 28 | 18 29 | 19  1 | 20  2 | 21  3 | 22  4 |
| 23  5 | 24  6 | 25  7 | 26  8 | 27  9 | 28 10 | 29 11 |
| 30 12 | 31 13 |       |       |       |       |       |
`,
		},
		{
			Name: "Gregorian month ending Sunday",
			Args: "4 2023",
			Want: `April 2023
Nisan - Iyyar 5783

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |       |       |       |       |       |  1 10 |
|  2 11 |  3 12 |  4 13 |  5 14 |  6 15 |  7 16 |  8 17 |
|  9 18 | 10 19 | 11 20 | 12 21 | 13 22 | 14 23 | 15 24 |
| 16 25 | 17 26 | 18 27 | 19 28 | 20 29 | 21 30 | 22  1 |
| 23  2 | 24  3 | 25  4 | 26  5 | 27  6 | 28  7 | 29  8 |
| 30  9 |       |       |       |       |       |       |
`,
		},
		{
			Name: "Gregorian month spanning 3 Hebrew months",
			Args: "5 2022",
			Want: `May 2022
Nisan - Sivan 5782

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|  1 30 |  2  1 |  3  2 |  4  3 |  5  4 |  6  5 |  7  6 |
|  8  7 |  9  8 | 10  9 | 11 10 | 12 11 | 13 12 | 14 13 |
| 15 14 | 16 15 | 17 16 | 18 17 | 19 18 | 20 19 | 21 20 |
| 22 21 | 23 22 | 24 23 | 25 24 | 26 25 | 27 26 | 28 27 |
| 29 28 | 30 29 | 31  1 |       |       |       |       |
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			args := []string{fpath}
			if c.Args != "" {
				args = append(args, strings.Fields(c.Args)...)
			}

			var buf bytes.Buffer

			err := cli.RunInEnvironment(
				args, files, now, templating.BuildData, &buf)
			test.CheckErr(t, err, c.Err)

			if buf.String() != c.Want {
				t.Errorf("want:\n%s\ngot:\n%s", c.Want, buf.String())
			}
		})
	}
}

func TestExamples_Date(t *testing.T) {
	files, now := setupExample(t)
	const fpath = "date.tmpl"

	cases := []struct {
		Name  string
		Args  string
		Want  string
		Usage bool
		Err   string
	}{
		{
			Name: "empty args",
			Args: "",
			Want: "Gregorian: 2024-05-06\nHebrew: 28 Nisan 5784\n",
		},
		{
			Name: "Gregorian day",
			Args: "6 7 2024",
			Want: "Gregorian: 2024-06-07\nHebrew: 1 Sivan 5784\n",
		},
		{
			Name: "Hebrew day",
			Args: "Sivan 1 5784",
			Want: "Gregorian: 2024-06-07\nHebrew: 1 Sivan 5784\n",
		},
		{
			Name: "Gregorian month",
			Args: "7 2024",
			Want: "Gregorian: 2024-07-01\nHebrew: 25 Sivan 5784\n",
		},
		{
			Name: "Hebrew month",
			Args: "Shevat 5784",
			Want: "Gregorian: 2024-01-11\nHebrew: 1 Sh'vat 5784\n",
		},
		{
			Name: "Gregorian year",
			Args: "2024",
			Want: "Gregorian: 2024-01-01\nHebrew: 20 Tevet 5784\n",
		},
		{
			Name: "Hebrew year",
			Args: "-c hebrewYear.json 5784",
			Want: "Gregorian: 2023-09-16\nHebrew: 1 Tishrei 5784\n",
		},
		{
			Name: "Hebrew year, with .. in config path",
			// This is a test of how our FSes get set up.
			// fs.Sub checks the path with fs.ValidPath,
			// which prevents use of .. and . .
			// The test here makes sure that even if the user chose a config
			// in such a path, hebcalfmt continues to function.
			Args: "-c ../examples/hebrewYear.json 5784",
			Want: "Gregorian: 2023-09-16\nHebrew: 1 Tishrei 5784\n",
		},
		{
			Name:  "Invalid Hebrew month",
			Args:  "badmonth 5784",
			Usage: true,
			Err:   "usage error: Gregorian months must be numeric, got \"badmonth\"",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			logBuf := test.Logger(t)
			args := []string{fpath}
			if c.Args != "" {
				args = append(args, strings.Fields(c.Args)...)
			}

			var buf bytes.Buffer

			err := cli.RunInEnvironment(
				args, files, now, templating.BuildData, &buf)
			test.CheckErr(t, err, c.Err)

			if buf.String() != c.Want {
				t.Errorf("want:\n%s\ngot:\n%s", c.Want, buf.String())
			}

			if c.Usage {
				if !strings.HasPrefix(logBuf.String(), "usage:\n  hebcalfmt ") {
					t.Errorf("expected usage message, got:\n%s", logBuf)
				}
			}
		})
	}
}

func TestExamples_HebcalClassic(t *testing.T) {
	files, now := setupExample(t)
	const fpath = "hebcalClassic.tmpl"

	cases := []struct {
		Name string
		Args string
		Want string
		Err  string
	}{
		{
			Name: "empty args",
			Args: "",
			Want: `1/11/2024 Rosh Chodesh Sh'vat
1/25/2024 Tu BiShvat
1/27/2024 Shabbat Shirah
2/9/2024 Rosh Chodesh Adar I
2/10/2024 Rosh Chodesh Adar I
2/23/2024 Purim Katan
2/24/2024 Shushan Purim Katan
3/9/2024 Shabbat Shekalim
3/10/2024 Rosh Chodesh Adar II
3/11/2024 Rosh Chodesh Adar II
3/21/2024 Ta'anit Esther
3/23/2024 Erev Purim
3/23/2024 Shabbat Zachor
3/24/2024 Purim
3/25/2024 Shushan Purim
3/30/2024 Shabbat Parah
4/6/2024 Shabbat HaChodesh
4/9/2024 Rosh Chodesh Nisan
4/18/2024 Yom HaAliyah
4/20/2024 Shabbat HaGadol
4/22/2024 Erev Pesach
4/22/2024 Ta'anit Bechorot
4/23/2024 Pesach I
4/24/2024 Pesach II
4/25/2024 Pesach III (CH''M)
4/26/2024 Pesach IV (CH''M)
4/27/2024 Pesach V (CH''M)
4/28/2024 Pesach VI (CH''M)
4/29/2024 Pesach VII
4/30/2024 Pesach VIII
5/6/2024 Yom HaShoah
5/8/2024 Rosh Chodesh Iyyar
5/9/2024 Rosh Chodesh Iyyar
5/13/2024 Yom HaZikaron
5/14/2024 Yom HaAtzma'ut
5/22/2024 Pesach Sheni
5/26/2024 Lag BaOmer
6/5/2024 Yom Yerushalayim
6/7/2024 Rosh Chodesh Sivan
6/11/2024 Erev Shavuot
6/12/2024 Shavuot I
6/13/2024 Shavuot II
7/6/2024 Rosh Chodesh Tammuz
7/7/2024 Rosh Chodesh Tammuz
7/23/2024 Tzom Tammuz
8/5/2024 Rosh Chodesh Av
8/10/2024 Shabbat Chazon
8/12/2024 Erev Tish'a B'Av
8/13/2024 Tish'a B'Av
8/17/2024 Shabbat Nachamu
8/19/2024 Tu B'Av
9/3/2024 Rosh Chodesh Elul
9/4/2024 Rosh Chodesh Elul
9/4/2024 Rosh Hashana LaBehemot
9/28/2024 Leil Selichot
10/2/2024 Erev Rosh Hashana
10/3/2024 Rosh Hashana 5785
10/4/2024 Rosh Hashana II
10/5/2024 Shabbat Shuva
10/6/2024 Tzom Gedaliah
10/11/2024 Erev Yom Kippur
10/12/2024 Yom Kippur
10/16/2024 Erev Sukkot
10/17/2024 Sukkot I
10/18/2024 Sukkot II
10/19/2024 Sukkot III (CH''M)
10/20/2024 Sukkot IV (CH''M)
10/21/2024 Sukkot V (CH''M)
10/22/2024 Sukkot VI (CH''M)
10/23/2024 Sukkot VII (Hoshana Raba)
10/24/2024 Shmini Atzeret
10/25/2024 Simchat Torah
11/1/2024 Rosh Chodesh Cheshvan
11/2/2024 Rosh Chodesh Cheshvan
11/28/2024 Sigd
12/1/2024 Rosh Chodesh Kislev
12/2/2024 Rosh Chodesh Kislev
12/25/2024 Chanukah: 1 Candle
12/26/2024 Chanukah: 2 Candles
12/27/2024 Chanukah: 3 Candles
12/28/2024 Chanukah: 4 Candles
12/29/2024 Chanukah: 5 Candles
12/30/2024 Chanukah: 6 Candles
12/31/2024 Chag HaBanot
12/31/2024 Chanukah: 7 Candles
12/31/2024 Rosh Chodesh Tevet
`,
		},
		{
			Name: "non-holiday",
			Args: "8 1 2024",
			Want: `8/1/2024 26th of Tammuz, 5784
`,
		},
		{
			Name: "today",
			Args: "-c today.json",
			Want: `5/6/2024 28th of Nisan, 5784
5/6/2024 Yom HaShoah
5/6/2024 13th day of the Omer
`,
		},
		{
			Name: "Chanukah",
			Args: "Kislev 28 5786",
			Want: `12/18/2025 28th of Kislev, 5786
12/18/2025 Chanukah: 5 Candles
`,
		},
		{
			Name: "Tishrei 5787",
			Args: "Tishrei 5787",
			Want: `9/12/2026 Rosh Hashana 5787
9/13/2026 Rosh Hashana II
9/14/2026 Tzom Gedaliah
9/19/2026 Shabbat Shuva
9/20/2026 Erev Yom Kippur
9/21/2026 Yom Kippur
9/25/2026 Erev Sukkot
9/26/2026 Sukkot I
9/27/2026 Sukkot II
9/28/2026 Sukkot III (CH''M)
9/29/2026 Sukkot IV (CH''M)
9/30/2026 Sukkot V (CH''M)
10/1/2026 Sukkot VI (CH''M)
10/2/2026 Sukkot VII (Hoshana Raba)
10/3/2026 Shmini Atzeret
10/4/2026 Simchat Torah
10/11/2026 Rosh Chodesh Cheshvan
`,
		},
		{
			Name: "events and yahrzeits files",
			Args: "--config events.json 9 1968",
			Want: `9/24/1968 2nd of Tishrei, 5729
9/24/1968 Birthday - Ben Ploni (5713)
9/26/1968 4th of Tishrei, 5729
9/26/1968 Yahrzeit - Joe Shmo
`,
		},
		{
			Name: "invalid events file path",
			Args: "--config invalidEvents.json 9 1968",
			Err:  "failed to build hebcal options from invalidEvents.json: attempted access outside of the BaseDir .: open ../examples/event.txt",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			args := []string{fpath}
			if c.Args != "" {
				args = append(args, strings.Fields(c.Args)...)
			}

			var buf bytes.Buffer

			err := cli.RunInEnvironment(
				args, files, now, templating.BuildData, &buf)
			test.CheckErr(t, err, c.Err)

			if buf.String() != c.Want {
				t.Errorf("want:\n%s\ngot:\n%s", c.Want, buf.String())
			}
		})
	}
}

func TestExamples_Mincha(t *testing.T) {
	files, now := setupExample(t)
	const fpath = "mincha.tmpl"

	cases := []struct {
		Name string
		Args string
		Want string
		Err  string
	}{
		{
			Name: "empty args",
			Args: "",
			Want: `Mon May 06, 2024: 7:40 PM
Tue May 07, 2024: 7:40 PM
Wed May 08, 2024: 7:40 PM
Thu May 09, 2024: 7:45 PM
Fri May 10, 2024: 7:45 PM
Sat May 11, 2024: 7:45 PM
Sun May 12, 2024: 7:45 PM
Mon May 13, 2024: 7:45 PM
Tue May 14, 2024: 7:50 PM
Wed May 15, 2024: 7:50 PM
Thu May 16, 2024: 7:50 PM
Fri May 17, 2024: 7:50 PM
Sat May 18, 2024: 7:50 PM
Sun May 19, 2024: 7:55 PM
`,
		},
		{
			Name: "decreasing times",
			Args: "8 1 2024",
			Want: `Thu Aug 01, 2024: 7:55 PM
Fri Aug 02, 2024: 7:55 PM
Sat Aug 03, 2024: 7:50 PM
Sun Aug 04, 2024: 7:50 PM
Mon Aug 05, 2024: 7:50 PM
Tue Aug 06, 2024: 7:50 PM
Wed Aug 07, 2024: 7:45 PM
Thu Aug 08, 2024: 7:45 PM
Fri Aug 09, 2024: 7:45 PM
Sat Aug 10, 2024: 7:45 PM
Sun Aug 11, 2024: 7:40 PM
Mon Aug 12, 2024: 7:40 PM
Tue Aug 13, 2024: 7:40 PM
Wed Aug 14, 2024: 7:40 PM
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			args := []string{fpath}
			if c.Args != "" {
				args = append(args, strings.Fields(c.Args)...)
			}

			var buf bytes.Buffer

			err := cli.RunInEnvironment(
				args, files, now, templating.BuildData, &buf)
			test.CheckErr(t, err, c.Err)

			if buf.String() != c.Want {
				t.Errorf("want:\n%s\ngot:\n%s", c.Want, buf.String())
			}
		})
	}
}

func TestExamples_ThisShabbat(t *testing.T) {
	files, now := setupExample(t)
	const fpath = "thisShabbat.tmpl"

	cases := []struct {
		Name string
		Args string
		Want string
		Err  string
	}{
		{
			Name: "empty args",
			Args: "",
			Want: `This Shabbat in Phoenix:

Erev Shabbat: Fri May 10 2024 / 2 Iyyar 5784
07:00 PM: Candle lighting

Shabbat: Sat May 11 2024 / 3 Iyyar 5784
04:07 AM: Alot haShachar
04:33 AM: Misheyakir
04:40 AM: Misheyakir Machmir
05:30 AM: Sunrise
08:21 AM: Kriat Shema, sof zeman (MGA)
08:57 AM: Kriat Shema, sof zeman (GRA)
09:42 AM: Tefilah, sof zeman (MGA)
10:06 AM: Tefilah, sof zeman (GRA)
12:24 PM: Chatzot hayom
12:59 PM: Mincha Gedolah
04:26 PM: Mincha Ketanah
05:52 PM: Plag HaMincha
07:18 PM: Sunset
07:38 PM: Bein HaShemashot
08:31 PM: Havdalah
`,
		},
		{
			Name: "provide a Friday",
			Args: "5 10 2024",
			Want: `This Shabbat in Phoenix:

Erev Shabbat: Fri May 10 2024 / 2 Iyyar 5784
07:00 PM: Candle lighting

Shabbat: Sat May 11 2024 / 3 Iyyar 5784
04:07 AM: Alot haShachar
04:33 AM: Misheyakir
04:40 AM: Misheyakir Machmir
05:30 AM: Sunrise
08:21 AM: Kriat Shema, sof zeman (MGA)
08:57 AM: Kriat Shema, sof zeman (GRA)
09:42 AM: Tefilah, sof zeman (MGA)
10:06 AM: Tefilah, sof zeman (GRA)
12:24 PM: Chatzot hayom
12:59 PM: Mincha Gedolah
04:26 PM: Mincha Ketanah
05:52 PM: Plag HaMincha
07:18 PM: Sunset
07:38 PM: Bein HaShemashot
08:31 PM: Havdalah
`,
		},
		{
			Name: "provide that Shabbos",
			Args: "5 11 2024",
			Want: `This Shabbat in Phoenix:

Erev Shabbat: Fri May 10 2024 / 2 Iyyar 5784
07:00 PM: Candle lighting

Shabbat: Sat May 11 2024 / 3 Iyyar 5784
04:07 AM: Alot haShachar
04:33 AM: Misheyakir
04:40 AM: Misheyakir Machmir
05:30 AM: Sunrise
08:21 AM: Kriat Shema, sof zeman (MGA)
08:57 AM: Kriat Shema, sof zeman (GRA)
09:42 AM: Tefilah, sof zeman (MGA)
10:06 AM: Tefilah, sof zeman (GRA)
12:24 PM: Chatzot hayom
12:59 PM: Mincha Gedolah
04:26 PM: Mincha Ketanah
05:52 PM: Plag HaMincha
07:18 PM: Sunset
07:38 PM: Bein HaShemashot
08:31 PM: Havdalah
`,
		},
		{
			Name: "provide a Sunday",
			Args: "5 12 2024",
			Want: `This Shabbat in Phoenix:

Erev Shabbat: Fri May 17 2024 / 9 Iyyar 5784
07:05 PM: Candle lighting

Shabbat: Sat May 18 2024 / 10 Iyyar 5784
04:00 AM: Alot haShachar
04:27 AM: Misheyakir
04:34 AM: Misheyakir Machmir
05:25 AM: Sunrise
08:19 AM: Kriat Shema, sof zeman (MGA)
08:55 AM: Kriat Shema, sof zeman (GRA)
09:41 AM: Tefilah, sof zeman (MGA)
10:05 AM: Tefilah, sof zeman (GRA)
12:24 PM: Chatzot hayom
12:59 PM: Mincha Gedolah
04:29 PM: Mincha Ketanah
05:56 PM: Plag HaMincha
07:24 PM: Sunset
07:44 PM: Bein HaShemashot
08:36 PM: Havdalah
`,
		},
		{
			Name: "different date, single-digit day",
			Args: "8 1 2024",
			Want: `This Shabbat in Phoenix:

Erev Shabbat: Fri Aug 2 2024 / 27 Tammuz 5784
07:09 PM: Candle lighting

Shabbat: Sat Aug 3 2024 / 28 Tammuz 5784
04:19 AM: Alot haShachar
04:45 AM: Misheyakir
04:53 AM: Misheyakir Machmir
05:42 AM: Sunrise
08:32 AM: Kriat Shema, sof zeman (MGA)
09:08 AM: Kriat Shema, sof zeman (GRA)
09:53 AM: Tefilah, sof zeman (MGA)
10:17 AM: Tefilah, sof zeman (GRA)
12:34 PM: Chatzot hayom
01:08 PM: Mincha Gedolah
04:34 PM: Mincha Ketanah
06:00 PM: Plag HaMincha
07:26 PM: Sunset
07:45 PM: Bein HaShemashot
08:38 PM: Havdalah
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			args := []string{fpath, "--config", "thisShabbat.json"}
			if c.Args != "" {
				args = append(args, strings.Fields(c.Args)...)
			}

			var buf bytes.Buffer

			err := cli.RunInEnvironment(
				args, files, now, templating.BuildData, &buf)
			test.CheckErr(t, err, c.Err)

			if buf.String() != c.Want {
				t.Errorf("want:\n%s\ngot:\n%s", c.Want, buf.String())
			}
		})
	}
}

func TestExamples_CustomZmanim(t *testing.T) {
	files, now := setupExample(t)
	const fpath = "customZmanim.tmpl"

	cases := []struct {
		Name string
		City string
		Date string
		Want string
		Err  string
	}{
		{
			Name: "empty inputs",
			Want: `Displaying zmanim for 2024-05-06 in Phoenix.

04:12:31: Alot HaShachar
04:38:19: Misheyakir
05:34:45: Netz
12:24:51: Chatzot
19:14:57: Shkiah
19:55:07: 8.5 degrees below horizon
21:07:07: 72m after 8.5 degrees below horizon

A halachic hour is 1h8m21s.
05:34:45: 0 halachic hours
06:43:06: 1 halachic hour
07:51:27: 2 halachic hours
08:59:48: 3 halachic hours
10:08:09: 4 halachic hours
11:16:30: 5 halachic hours
12:24:51: 6 halachic hours
13:33:12: 7 halachic hours
14:41:33: 8 halachic hours
15:49:54: 9 halachic hours
16:58:15: 10 halachic hours
18:06:36: 11 halachic hours
19:14:57: 12 halachic hours
`,
		},
		{
			Name: "city=Johannesburg",
			City: "Johannesburg",
			Want: `Displaying zmanim for 2024-05-06 in Johannesburg.

05:23:58: Alot HaShachar
05:44:56: Misheyakir
06:34:20: Netz
12:04:24: Chatzot
17:34:29: Shkiah
18:10:06: 8.5 degrees below horizon
19:22:06: 72m after 8.5 degrees below horizon

A halachic hour is 55m1s.
06:34:20: 0 halachic hours
07:29:20: 1 halachic hour
08:24:21: 2 halachic hours
09:19:22: 3 halachic hours
10:14:23: 4 halachic hours
11:09:23: 5 halachic hours
12:04:24: 6 halachic hours
12:59:25: 7 halachic hours
13:54:26: 8 halachic hours
14:49:26: 9 halachic hours
15:44:27: 10 halachic hours
16:39:28: 11 halachic hours
17:34:29: 12 halachic hours
`,
		},
		{
			Name: "date=2024-05-18",
			Date: "2024-05-18",
			Want: `Displaying zmanim for 2024-05-18 in Phoenix.

04:00:19: Alot HaShachar
04:27:23: Misheyakir
05:25:45: Netz
12:24:53: Chatzot
19:24:01: Shkiah
20:05:27: 8.5 degrees below horizon
21:17:27: 72m after 8.5 degrees below horizon

A halachic hour is 1h9m51s.
05:25:45: 0 halachic hours
06:35:36: 1 halachic hour
07:45:27: 2 halachic hours
08:55:19: 3 halachic hours
10:05:10: 4 halachic hours
11:15:01: 5 halachic hours
12:24:53: 6 halachic hours
13:34:44: 7 halachic hours
14:44:35: 8 halachic hours
15:54:27: 9 halachic hours
17:04:18: 10 halachic hours
18:14:09: 11 halachic hours
19:24:01: 12 halachic hours
`,
		},
		{
			Name: "invalid date",
			Date: "bad date",
			Err:  `template: customZmanim.tmpl:6:31: executing "customZmanim.tmpl" at <timeParse $.time.DateOnly .>: error calling timeParse: parsing time "bad date" as "2006-01-02": cannot parse "bad date" as "2006"`,
		},
		{
			Name: "invalid city",
			City: "Bad City",
			Err:  `template: customZmanim.tmpl:3:12: executing "customZmanim.tmpl" at <lookupCity $city>: error calling lookupCity: unknown city "Bad City"`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			args := []string{fpath}
			if c.City != "" {
				t.Setenv("CITY", c.City)
			}
			if c.Date != "" {
				t.Setenv("DATE", c.Date)
			}

			var buf bytes.Buffer

			err := cli.RunInEnvironment(
				args, files, now, templating.BuildData, &buf)
			test.CheckErr(t, err, c.Err)

			if buf.String() != c.Want {
				t.Errorf("want:\n%s\ngot:\n%s", c.Want, buf.String())
			}
		})
	}
}
