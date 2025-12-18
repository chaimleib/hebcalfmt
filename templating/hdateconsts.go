package templating

import (
	"github.com/hebcal/hdate"
)

var HDateConsts = map[string]any{
	"Nisan":    hdate.Nisan,
	"Iyyar":    hdate.Iyyar,
	"Sivan":    hdate.Sivan,
	"Tamuz":    hdate.Tamuz,
	"Av":       hdate.Av,
	"Elul":     hdate.Elul,
	"Tishrei":  hdate.Tishrei,
	"Cheshvan": hdate.Cheshvan,
	"Kislev":   hdate.Kislev,
	"Tevet":    hdate.Tevet,
	"Shvat":    hdate.Shvat,
	"Adar1":    hdate.Adar1,
	"Adar2":    hdate.Adar2,

	"Epoch": hdate.Epoch,
}
