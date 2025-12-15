# hebcalfmt

Formats hebcal information using the Go Template language.

Usage: `hebcalfmt path/to/template.tmpl`

## Example: Show today's date

`examples/date.tmpl`
```tmpl
Today: {{$.now.Format $.time.DateOnly}}
Hebrew: {{hdateFromTime $.now}}
```

```bash
$ hebcalfmt examples/date.tmpl
Today: 2025-12-14
Hebrew: 24 Kislev 5786
```

## Example: Custom zmanim for a configurable day

<details>
    <summary>`examples/customZmanim.tmpl`</summary>

```tmpl
{{- $loc := getenv "CITY" | lookupCity -}}
{{- $tz := timeLoadLocation $loc.TimeZoneId -}}
{{- $d := timeParseInLocation $.time.DateOnly (getenv "DATE") $tz -}}
{{- $z := forLocationDate $loc $d -}}
Displaying zmanim for {{$d.Format $.time.DateOnly}} in {{$loc.Name}}.

{{- $fmt := $.time.TimeOnly}}

{{( (timeAtAngle $z $tz 8.5 true).Add (timeParseDuration "-72m") ).Format $fmt}} : 72 minutes before 8.5 degrees below horizon
{{ (timeAtAngle $z $tz 8.5 true).Format $fmt}} : 8.5 degrees below horizon
{{$z.AlotHaShachar.Format $fmt}} : Alot HaShachar
{{$z.Misheyakir.Format $fmt}} : Misheyakir 
{{$z.Sunrise.Format $fmt}} : Netz
{{$z.Chatzot.Format $fmt}} : Chatzot
{{$z.Sunset.Format $fmt}} : Shkiah

A halachic hour is {{ ($z.Hour | secondsDuration).Round $.time.Second}}.
{{- range 13}}
{{    (hourOffset $z $tz (itof .)).Format $fmt}} : {{.}} halachic hour{{if ne . 1}}s{{end}}
{{- end}}
```

</details>

```bash
$ CITY="Los Angeles" DATE=2025-12-14 go run . examples/customZmanim.tmpl
Displaying zmanim for 2025-12-14 in Los Angeles.

04:57:57 : 72 minutes before 8.5 degrees below horizon
06:09:57 : 8.5 degrees below horizon
05:30:55 : Alot HaShachar
05:54:24 : Misheyakir
06:50:50 : Netz
11:47:54 : Chatzot
16:44:58 : Shkiah

A halachic hour is 49m31s.
06:50:50 : 0 halachic hours
07:40:20 : 1 halachic hour
08:29:51 : 2 halachic hours
09:19:22 : 3 halachic hours
10:08:52 : 4 halachic hours
10:58:23 : 5 halachic hours
11:47:54 : 6 halachic hours
12:37:24 : 7 halachic hours
13:26:55 : 8 halachic hours
14:16:26 : 9 halachic hours
15:05:56 : 10 halachic hours
15:55:27 : 11 halachic hours
16:44:58 : 12 halachic hours
```

## Example: Show zmanim for this Shabbos

<details>
    <summary>`examples/thisShabbos.tmpl`</summary>

```tmpl
{{- $d := .now}}
{{- $loc := LookupCity "Phoenix"}}
{{- $timeFormat := "03:04 PM"}}
{{- range 7}}
{{-   if ne $d.Weekday $.time.Saturday}}
{{-     $d = $d.AddDate 0 0 1}}
{{-     continue}}
{{-   end}}
{{-   $z := forLocationDate $loc $d}}
{{-   $erev := $d.AddDate 0 0 -1}}
{{-   $zErev := forLocationDate $loc $erev}}

{{- /* swallow whitespace */ -}}

Erev Shabbat: {{$erev.Format "Mon Jan 02 2006"}} / {{hdateFromTime $erev}}
{{-   range timedEvents $zErev}}
{{-     if eq .Flags $.event.ZMANIM}}
{{-       continue}}
{{-     end}}
{{      .EventTime.Format $timeFormat}}
{{- ": "}}
{{-     .Desc}}
{{-   end}}

Shabbat: {{$d.Format "Mon Jan 02 2006"}} / {{hdateFromTime $d}}
{{-   range timedEvents $z}}
{{-     if eq .Desc "Tzeit HaKochavim"}}
{{-       continue}}
{{-     end}}
{{      .EventTime.Format $timeFormat}}
{{- ": "}}
{{-     .Desc}}
{{-   end}}

{{- end}}
```

</details>

```bash
$ hebcalfmt thisShabbat.tmpl
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
06:04 PM: Havdalah
06:04 PM: Chanukah: 7 Candles
```

## Example: Show this month's calendar with Hebrew dates

<details>
    <summary>`examples/monthCalendar.tmpl`</summary>

```tmpl
{{- /* A hyphen ("-") at the beginning or end of a directive means
to delete whitespace in that direction until a file boundary,
non-whitespace, or another directive. */}}

{{- /* Month calendar */}}
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

