package service_cards

import (
	"hkapp/application"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type withLayoutFunc struct {
	content func(C) D
}

func (t withLayoutFunc) Layout(gtx C) D {
	return t.content(gtx)
}

func GetWidgetForService(app *application.App,
	acc *hkontroller.Accessory, dev *hkontroller.Device,
	s *hkontroller.ServiceDescription,
	quickWidget bool,
) (interface {
	Layout(C) D
}, error) {

	label := ""
	accInfo := acc.GetService(hkontroller.SType_AccessoryInfo)
	if accInfo != nil {
		accName := accInfo.GetCharacteristic(hkontroller.CType_Name)
		if accName != nil {
			name, ok := accName.Value.(string)
			if ok {
				label = name
			}
		}
	}

	var w interface {
		Layout(C) D
	}
	var err error

	switch s.Type {
	case hkontroller.SType_LightBulb:
		w, err = NewLightBulb(app, acc, dev, quickWidget)
	case hkontroller.SType_Switch:
		w, err = NewSwitch(app, acc, dev, quickWidget)
	case hkontroller.SType_AccessoryInfo:
		w, err = NewAccessoryInfo(app, acc, dev, quickWidget)
	case hkontroller.SType_Thermostat:
		w, err = NewThermostat(app, acc, dev, quickWidget)
	default:
		w = material.Body2(app.Theme, label)
	}

	return w, err
}
