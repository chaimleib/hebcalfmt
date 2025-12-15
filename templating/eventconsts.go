package templating

import "github.com/hebcal/hebcal-go/event"

var EventConsts = map[string]any{
	"CHAG":                event.CHAG,
	"LIGHT_CANDLES":       event.LIGHT_CANDLES,
	"YOM_TOV_ENDS":        event.YOM_TOV_ENDS,
	"CHUL_ONLY":           event.CHUL_ONLY,
	"IL_ONLY":             event.IL_ONLY,
	"LIGHT_CANDLES_TZEIS": event.LIGHT_CANDLES_TZEIS,
	"CHANUKAH_CANDLES":    event.CHANUKAH_CANDLES,
	"ROSH_CHODESH":        event.ROSH_CHODESH,
	"MINOR_FAST":          event.MINOR_FAST,
	"SPECIAL_SHABBAT":     event.SPECIAL_SHABBAT,
	"PARSHA_HASHAVUA":     event.PARSHA_HASHAVUA,
	"DAF_YOMI":            event.DAF_YOMI,
	"OMER_COUNT":          event.OMER_COUNT,
	"MODERN_HOLIDAY":      event.MODERN_HOLIDAY,
	"MAJOR_FAST":          event.MAJOR_FAST,
	"SHABBAT_MEVARCHIM":   event.SHABBAT_MEVARCHIM,
	"MOLAD":               event.MOLAD,
	"USER_EVENT":          event.USER_EVENT,
	"HEBREW_DATE":         event.HEBREW_DATE,
	"MINOR_HOLIDAY":       event.MINOR_HOLIDAY,
	"EREV":                event.EREV,
	"CHOL_HAMOED":         event.CHOL_HAMOED,
	"MISHNA_YOMI":         event.MISHNA_YOMI,
	"YOM_KIPPUR_KATAN":    event.YOM_KIPPUR_KATAN,
	"ZMANIM":              event.ZMANIM,
	"YERUSHALMI_YOMI":     event.YERUSHALMI_YOMI,
	"NACH_YOMI":           event.NACH_YOMI,
}
