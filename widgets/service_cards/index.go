package service_cards

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"hkapp/appmanager"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func GetWidgetForService(am *appmanager.AppManager,
	acc *hkontroller.Accessory, dev *hkontroller.Device,
	s *hkontroller.ServiceDescription, th *material.Theme,
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

	switch s.Type {
	case hkontroller.SType_LightBulb:
		return NewLightBulb(am, acc, dev, th)
	case hkontroller.SType_Switch:
		return NewSwitch(am, acc, dev, th)
	default:
		return material.Body2(th, label), nil
	}
}
