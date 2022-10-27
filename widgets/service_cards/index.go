package service_cards

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func GetWidgetForService(acc *hkontroller.Accessory, s *hkontroller.ServiceDescription, th *material.Theme) layout.Widget {

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

	switch s.Type {
	case hkontroller.SType_LightBulb:
		return NewLightBulb(acc.Id, label, th).Layout
	case hkontroller.SType_Switch:
		return NewSwitch(acc.Id, label, th).Layout
	default:
		return material.Body2(th, label).Layout
	}
}
