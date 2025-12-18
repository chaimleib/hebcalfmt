# hebcalfmt

Formats [hebcal](https://github.com/hebcal/hebcal) information using
the [Go Template language](https://pkg.go.dev/text/template).

```text
usage:
  hebcalfmt [{ --config | -c } config.json ] template.tmpl [[ month [ day ]] year ]
  hebcalfmt { --info | -i } { cities | default-city | languages }
  hebcalfmt [ -h | --help | --version ]

OPTIONS:
  -c, --config string   select a JSON config file (default $HOME/.config/hebcalfmt/config.json)
  -h, --help            print this help text
  -i, --info string     show data from the internal databases or compiled values. Available options: cities, default-city, languages
      --version         show version number
```

## Install

```bash
go install github.com/chaimleib/hebcalfmt
```

## Example: Show today's date

examples/today.tmpl
```tmpl
Gregorian: {{$.now.Format $.time.DateOnly}}
Hebrew: {{hdateFromTime $.now}}
```

```bash
$ hebcalfmt examples/today.tmpl
Today: 2025-12-14
Hebrew: 24 Kislev 5786
```

## Example: Convert dates between Hebrew and Gregorian

examples/date.tmpl
```tmpl
{{- /* Read date from CLI args, as in classic hebcal, with NoJulian=false. */}}
{{- $d := $.dateRange.Start false}}
{{- /* In classic hebcal, no args defaults to just the year. */}}
{{- /* We want today's full date. */}}
{{- if $.dateRange.Source.IsZero}}
{{-   $d = hdateFromTime $.now}}
{{- end -}}

Gregorian: {{$d.Gregorian.Format $.time.DateOnly}}
Hebrew: {{$d.Day}} {{$d.MonthName $.language}} {{$d.Year}}
```

```bash
$ hebcalfmt examples/date.tmpl
Gregorian: 2025-12-18
Hebrew: 28 Kislev 5786

$ hebcalfmt examples/date.tmpl 12 1 2020
Gregorian: 2020-12-01
Hebrew: 15 Kislev 5781

$ hebcalfmt examples/date.tmpl Kislev 25 5786
Gregorian: 2025-12-15
Hebrew: 25 Kislev 5786
```

## Example: Classic Hebcal

This example emulates the classic hebcal program,
while exposing some internals for you to customize, if you wish.

examples/hebcalClassic.tmpl
```tmpl
{{- range hebcal}}
{{-   .GetDate.Gregorian.Format "01/02/2006 "}}
{{-     .Render $.language}}
{{  end}}
```

examples/yahrzeit.txt
```text
1 2 1967 Joe Shmo
```

```bash
$ hebcalfmt examples/hebcalClassic.tmpl
01/01/2025 Chanukah: 8 Candles
01/01/2025 Rosh Chodesh Tevet
01/02/2025 Chanukah: 8th Day
...
12/21/2025 Chanukah: 8 Candles
12/21/2025 Rosh Chodesh Tevet
12/22/2025 Chanukah: 8th Day
12/30/2025 Asara B'Tevet

$ hebcalfmt -c <(echo '{"today": true}') examples/hebcalClassic.tmpl
12/18/2025 28th of Kislev, 5786
12/18/2025 Chanukah: 5 Candles

$ hebcalfmt examples/hebcalClassic.tmpl Tishrei 5787
09/12/2026 Rosh Hashana 5787
09/13/2026 Rosh Hashana II
09/14/2026 Tzom Gedaliah
09/19/2026 Shabbat Shuva
09/20/2026 Erev Yom Kippur
09/21/2026 Yom Kippur
09/25/2026 Erev Sukkot
09/26/2026 Sukkot I
09/27/2026 Sukkot II
09/28/2026 Sukkot III (CH''M)
09/29/2026 Sukkot IV (CH''M)
09/30/2026 Sukkot V (CH''M)
10/01/2026 Sukkot VI (CH''M)
10/02/2026 Sukkot VII (Hoshana Raba)
10/03/2026 Shmini Atzeret
10/04/2026 Simchat Torah
10/11/2026 Rosh Chodesh Cheshvan

$ hebcalfmt -c <(echo '{"yahrzeits_file": "examples/yahrzeit.txt"}') examples/hebcalClassic.tmpl 1 1968
01/01/1968 Chag HaBanot
01/01/1968 Chanukah: 7 Candles
01/01/1968 Rosh Chodesh Tevet
01/02/1968 Chanukah: 8 Candles
01/02/1968 Rosh Chodesh Tevet
01/03/1968 Chanukah: 8th Day
01/11/1968 Asara B'Tevet
01/21/1968 Joe Shmo
01/31/1968 Rosh Chodesh Sh'vat
```

## Example: Calculate Mincha times

Some shuls adjust when Mincha begins
so that it always starts at least 15 minutes before sunset,
rounded to the quarter hour or five minutes.

examples/mincha.tmpl
```tmpl
{{- $d := timeParse $.time.DateOnly "2025-09-01"}}
{{- $rounding := timeParseDuration "5m"}}
{{- $advance := timeParseDuration "-15m"}}
{{- range 14}}
{{-   $z := forDate $d}}
{{-   $mincha := ($z.Sunset.Add $advance).Truncate $rounding}}
{{-   $mincha.Format "Mon Jan 02, 2006: 3:04 PM\n"}}
{{-   $d = $d.AddDate 0 0 1}}
{{- end}}
```

```bash
$ hebcalfmt examples/mincha.tmpl
Mon Sep 01, 2025: 6:35 PM
Tue Sep 02, 2025: 6:35 PM
Wed Sep 03, 2025: 6:35 PM
Thu Sep 04, 2025: 6:35 PM
Fri Sep 05, 2025: 6:30 PM
Sat Sep 06, 2025: 6:30 PM
Sun Sep 07, 2025: 6:30 PM
Mon Sep 08, 2025: 6:25 PM
Tue Sep 09, 2025: 6:25 PM
Wed Sep 10, 2025: 6:25 PM
Thu Sep 11, 2025: 6:25 PM
Fri Sep 12, 2025: 6:20 PM
Sat Sep 13, 2025: 6:20 PM
Sun Sep 14, 2025: 6:20 PM
```

## Example: Custom zmanim for a configurable day and city

Although `$.z`, `$.location`, `$.now`, and `$.tz` are provided for convenience,
you aren't limited to using preconfigured values.
You have the power to parse them from environment variables of your choosing.

You also can choose which zmanim to display,
and even how to compute your zmanim.
We provide `timeAtAngle`, `hourOffset` and `timeParseDuration` for this purpose.
These are useful if you use a different opinion than hebcal's defaults,
or if you simply want to switch your water sprinkler on after dark.

<details>
    <summary>examples/customZmanim.tmpl</summary>

```tmpl
{{- $city := "Phoenix" -}}
{{- with getenv "CITY"}}{{$city = .}}{{end -}}
{{- $loc := lookupCity $city -}}

{{- $tz := timeLoadLocation $loc.TimeZoneId -}}

{{- $d := $.now -}}
{{- with getenv "DATE"}}{{$d = timeParseInLocation $.time.DateOnly . $tz}}{{end -}}

{{- $z := forLocationDate $loc $d -}}
Displaying zmanim for {{$d.Format $.time.DateOnly}} in {{$loc.Name}}.

{{- $fmt := $.time.TimeOnly}}

{{$z.AlotHaShachar.Format $fmt}}: Alot HaShachar
{{$z.Misheyakir.Format $fmt}}: Misheyakir
{{$z.Sunrise.Format $fmt}}: Netz
{{$z.Chatzot.Format $fmt}}: Chatzot
{{$z.Sunset.Format $fmt}}: Shkiah
{{ $dark85 := timeAtAngle $z $tz 8.5 false -}}
{{ $dark85.Format $fmt -}}
  : 8.5 degrees below horizon
{{( $dark85.Add (timeParseDuration "72m") ).Format $fmt -}}
  : 72m after 8.5 degrees below horizon

A halachic hour is {{ ($z.Hour | secondsDuration).Round $.time.Second}}.
{{- range 13}}
{{    (hourOffset $z $tz (itof .)).Format $fmt -}}
  : {{.}} halachic hour{{if ne . 1}}s{{end}}
{{- end}}
```

</details>

```bash
$ CITY="Los Angeles" DATE=2025-12-14 hebcalfmt examples/customZmanim.tmpl
Displaying zmanim for 2025-12-14 in Los Angeles.

05:30:55: Alot HaShachar
05:54:24: Misheyakir
06:50:50: Netz
11:47:54: Chatzot
16:44:58: Shkiah
17:25:51: 8.5 degrees below horizon
18:37:51: 72m after 8.5 degrees below horizon

A halachic hour is 49m31s.
06:50:50: 0 halachic hours
07:40:20: 1 halachic hour
08:29:51: 2 halachic hours
09:19:22: 3 halachic hours
10:08:52: 4 halachic hours
10:58:23: 5 halachic hours
11:47:54: 6 halachic hours
12:37:24: 7 halachic hours
13:26:55: 8 halachic hours
14:16:26: 9 halachic hours
15:05:56: 10 halachic hours
15:55:27: 11 halachic hours
16:44:58: 12 halachic hours
```

## Example: Show zmanim for this Shabbos

Showing zmanim for upcoming days is also possible,
since you have the power to do arithmetic with times and durations.

If hebcal's classic output is mostly what you want,
but you are only interested in the zmanim,
we provide a `timedEvents` function for you.
You can then filter these `hebcal.TimedEvent` structs yourself,
based on their `.Flags` or `.Desc` or anything else.

<details>
    <summary>examples/thisShabbat.tmpl</summary>

```tmpl
This Shabbat in {{$.location.Name}}:

{{- /* Read date from CLI args, as in classic hebcal, with NoJulian=false. */}}
{{- $d := $.dateRange.Start false}}
{{- /* In classic hebcal, no args defaults to just the year. */}}
{{- /* We want today's full date. */}}
{{- if $.dateRange.Source.IsZero}}
{{-   $d = hdateFromTime $.now}}
{{- end}}

{{- $timeFormat := "03:04 PM"}}

{{- $d = $d.OnOrAfter $.time.Saturday}}
{{- $z := forLocationDate $.location $d.Gregorian}}
{{- $erev := $d.Prev}}
{{- $zErev := forLocationDate $.location $erev.Gregorian}}

Erev Shabbat: {{$erev.Gregorian.Format "Mon Jan 02 2006"}} / {{$erev}}
{{- range timedEvents $zErev}}
{{- /* Skip normal zmanim, only show candlelighting. */}}
{{-   if eq .Flags $.event.ZMANIM}}
{{-     continue}}
{{-   end}}
{{    .EventTime.Format $timeFormat -}}
        : {{.Desc}}
{{- end}}

Shabbat: {{$d.Gregorian.Format "Mon Jan 02 2006"}} / {{$d}}
{{- range timedEvents $z}}
{{-   if eq .Desc "Tzeit HaKochavim"}}{{/* redundant, so skip */}}
{{-     continue}}
{{-   end}}
{{    .EventTime.Format $timeFormat -}}
        : {{.Desc}}
{{- end}}
```

</details>

<details>
    <summary>examples/thisShabbat.json</summary>

```json
{
  "daily_zmanim": true,
  "candle_lighting": true,
  "city": "Phoenix"
}
```

</details>

```bash
$ hebcalfmt -c examples/thisShabbat.json examples/thisShabbat.tmpl
This Shabbat in Phoenix:

Erev Shabbat: Fri Dec 19 2025 / 29 Kislev 5786
05:05 PM: Chanukah: 6 Candles
05:05 PM: Candle lighting

Shabbat: Sat Dec 20 2025 / 30 Kislev 5786
06:08 AM: Alot haShachar
06:32 AM: Misheyakir
06:38 AM: Misheyakir Machmir
07:28 AM: Sunrise
09:21 AM: Kriat Shema, sof zeman (MGA)
09:57 AM: Kriat Shema, sof zeman (GRA)
10:22 AM: Tefilah, sof zeman (MGA)
10:46 AM: Tefilah, sof zeman (GRA)
12:26 PM: Chatzot hayom
12:50 PM: Mincha Gedolah
03:19 PM: Mincha Ketanah
04:21 PM: Plag HaMincha
05:24 PM: Sunset
05:43 PM: Bein HaShemashot
06:36 PM: Havdalah
06:36 PM: Chanukah: 7 Candles
```

## Example: Show this month's calendar with Hebrew dates

The control flow in the Go templating language is so powerful
that you can program your own calendar in it.
This one outputs a calendar in Markdown table format,
but JSON, and even HTML can be hacked together.

<details>
    <summary>examples/monthCalendar.tmpl</summary>

```tmpl
{{- /* A hyphen ("-") at the beginning or end of a directive means
to delete whitespace in that direction until a file boundary,
non-whitespace, or another directive. */}}

{{- /* English month label */}}
{{- $monthDay := timeDate $.now.Year $.now.Month 1 0 0 0 0 $.tz}}
{{- $nextMonth := $monthDay.AddDate 0 1 0}}
{{- $monthDay.Month}} {{$.now.Year}}

{{- /* Hebrew month label(s) */}}
{{- $hdate := hdateFromTime $monthDay}}
{{- $hdateLast := hdateFromTime ($nextMonth.AddDate 0 0 -1)}}
{{$hdate.Month}}
{{- if ne $hdate.Year $hdateLast.Year}} {{$hdate.Year}}{{end}}
{{- /* If the Hebrew month changes, hyphenate the range. */}}
{{- if ne $hdateLast.Month $hdate.Month}}
{{-   " - "}}{{$hdateLast.Month}}
{{- end}}
{{- " "}}{{$hdateLast.Year}}

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
{{- $cellWidth := 7}}
{{- /* header rule */}}
{{repeat (printf "|%s" (repeat "-" $cellWidth)) 7}}|

{{- /* Skip days of the week until day 1 of the month. */}}
{{- range 7}}
{{-   if eq $monthDay.Weekday .}}{{break}}{{end}}
{{-   "\n|"}}{{repeat " " 7}}
{{- end}}

{{- /* Number the cells with the Gregorian and Hebrew days of the month. */}}
{{- range 31}}
{{-   if not ($monthDay.Before $nextMonth)}}{{break}}{{end}}
{{-   "|" -}}
{{    if datePartsEqual $monthDay $.now}}*{{else}} {{end}}
{{-   printf "%2d %2d" $monthDay.Day (hdateFromTime $monthDay).Day}}
{{-   if datePartsEqual $monthDay $.now}}*{{else}} {{end}}
{{-   if eq $monthDay.Weekday $.time.Saturday}}
{{-     "|\n"}}
{{-   end}}
{{-   $monthDay = $monthDay.AddDate 0 0 1}}
{{- end}}

{{- /* blank boxes until end of week */ -}}
{{- range 8}}
{{-   if eq $monthDay.Weekday $.time.Sunday}}
{{-     "|"}}
{{-     break}}
{{-   end}}
{{-   "|"}}{{repeat " " $cellWidth}}
{{-   $monthDay = $monthDay.AddDate 0 0 1}}
{{- end}}
```

</details>

```bash
$ hebcalfmt examples/monthCalendar.tmpl
December 2025
Kislev - Tevet 5786

|  Sun  |  Mon  |  Tue  |  Wed  | Thurs |  Fri  |  Sat  |
|-------|-------|-------|-------|-------|-------|-------|
|       |  1 11 |  2 12 |  3 13 |  4 14 |  5 15 |  6 16 |
|  7 17 |  8 18 |  9 19 | 10 20 | 11 21 | 12 22 | 13 23 |
|*14 24*| 15 25 | 16 26 | 17 27 | 18 28 | 19 29 | 20 30 |
| 21  1 | 22  2 | 23  3 | 24  4 | 25  5 | 26  6 | 27  7 |
| 28  8 | 29  9 | 30 10 | 31 11 |       |       |       |
```

## Documentation for going deep

If you want to get the most out of `hebcalfmt`,
it helps to know these documents well:

 * The [`text/template` package docs](https://pkg.go.dev/text/template), for
 learning the templating language.
 * The
 [`github.com/hebcal/hdate.HDate`](https://pkg.go.dev/github.com/hebcal/hdate#HDate)
 type's fields and methods.
 * The [`github.com/hebcal/hebcal-go/zmanim.Zmanim`](https://pkg.go.dev/github.com/hebcal/hebcal-go/zmanim#Zmanim) type's fields and methods.
 * The [`github.com/chaimleib/hebcalfmt/templating`](https://github.com/chaimleib/hebcalfmt/tree/master/templating) folder
   contains \*funcs.go and \*consts.go files which are made available
   inside the template.
