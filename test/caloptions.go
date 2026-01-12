package test

import (
	"slices"

	"github.com/hebcal/hdate"
	"github.com/hebcal/hebcal-go/hebcal"
	"github.com/hebcal/hebcal-go/zmanim"
)

func CheckCalOptions(t Test, want, got *hebcal.CalOptions) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("expected nil, got: %#v", got)
		}
		return
	}
	if got == nil {
		t.Errorf("got nil, want: %#v", want)
		return
	}

	for _, field := range []struct {
		Name      string
		Want, Got any
	}{
		{"Location", want.Location, got.Location},
		{"Year", want.Year, got.Year},
		{"IsHebrewYear", want.IsHebrewYear, got.IsHebrewYear},
		{"NoJulian", want.NoJulian, got.NoJulian},
		{"Month", want.Month, got.Month},
		{"NumYears", want.NumYears, got.NumYears},
		{"Start", want.Start, got.Start},
		{"End", want.End, got.End},
		{"CandleLighting", want.CandleLighting, got.CandleLighting},
		{"CandleLightingMins", want.CandleLightingMins, got.CandleLightingMins},
		{"HavdalahMins", want.HavdalahMins, got.HavdalahMins},
		{"HavdalahDeg", want.HavdalahDeg, got.HavdalahDeg},
		{"Sedrot", want.Sedrot, got.Sedrot},
		{"IL", want.IL, got.IL},
		{"NoMinorFast", want.NoMinorFast, got.NoMinorFast},
		{"NoModern", want.NoModern, got.NoModern},
		{"NoRoshChodesh", want.NoRoshChodesh, got.NoRoshChodesh},
		{"ShabbatMevarchim", want.ShabbatMevarchim, got.ShabbatMevarchim},
		{"NoSpecialShabbat", want.NoSpecialShabbat, got.NoSpecialShabbat},
		{"NoHolidays", want.NoHolidays, got.NoHolidays},
		{"DafYomi", want.DafYomi, got.DafYomi},
		{"MishnaYomi", want.MishnaYomi, got.MishnaYomi},
		{"YerushalmiYomi", want.YerushalmiYomi, got.YerushalmiYomi},
		{"NachYomi", want.NachYomi, got.NachYomi},
		{"YerushalmiEdition", want.YerushalmiEdition, got.YerushalmiEdition},
		{"Omer", want.Omer, got.Omer},
		{"Molad", want.Molad, got.Molad},
		{"AddHebrewDates", want.AddHebrewDates, got.AddHebrewDates},
		{"AddHebrewDatesForEvents", want.AddHebrewDatesForEvents, got.AddHebrewDatesForEvents},
		{"Mask", want.Mask, got.Mask},
		{"YomKippurKatan", want.YomKippurKatan, got.YomKippurKatan},
		{"Hour24", want.Hour24, got.Hour24},
		{"SunriseSunset", want.SunriseSunset, got.SunriseSunset},
		{"DailyZmanim", want.DailyZmanim, got.DailyZmanim},
		{"Yahrzeits", want.Yahrzeits, got.Yahrzeits},
		{"UserEvents", want.UserEvents, got.UserEvents},
		{"WeeklyAbbreviated", want.WeeklyAbbreviated, got.WeeklyAbbreviated},
		{"DailySedra", want.DailySedra, got.DailySedra},
	} {
		switch typedWant := field.Want.(type) {
		case *zmanim.Location:
			CheckNilPtrThen(t, CheckComparable,
				field.Name, typedWant, field.Got)

		case []hebcal.UserYahrzeit:
			typedGot := field.Got.([]hebcal.UserYahrzeit)
			if !slices.Equal(typedWant, typedGot) {
				t.Errorf("%s's do not match - want:\n%v\ngot:\n%v",
					field.Name, field.Want, field.Got)
			}

		case []hebcal.UserEvent:
			typedGot := field.Got.([]hebcal.UserEvent)
			if !slices.Equal(typedWant, typedGot) {
				t.Errorf("%s's do not match - want:\n%v\ngot:\n%v",
					field.Name, field.Want, field.Got)
			}

		case hdate.HDate:
			typedGot := field.Got.(hdate.HDate)
			CheckHDate(t, field.Name, typedWant, typedGot)

		default:
			CheckComparable(t, field.Name, field.Want, field.Got)
		}
	}
}
