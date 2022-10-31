package service_cards

import (
	"errors"
	"fmt"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"hkapp/applayout"
	"hkapp/appmanager"
	"reflect"
)

type LightBulb struct {
	widget.Bool

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device
	th  *material.Theme

	*appmanager.AppManager
}

func NewLightBulb(am *appmanager.AppManager, acc *hkontroller.Accessory, dev *hkontroller.Device, th *material.Theme) (*LightBulb, error) {

	infoS := acc.GetService(hkontroller.SType_AccessoryInfo)
	if infoS == nil {
		return nil, errors.New("cannot get AccessoryInfo service")
	}
	labelC := infoS.GetCharacteristic(hkontroller.CType_Name)
	if labelC == nil {
		return nil, errors.New("cannot get characteristic Name")
	}
	label, ok := labelC.Value.(string)
	if !ok {
		return nil, errors.New("cannot extract accessory name")
	}

	lightbS := acc.GetService(hkontroller.SType_LightBulb)
	if lightbS == nil {
		return nil, errors.New("cannot find LightBulb service")
	}
	onC := lightbS.GetCharacteristic(hkontroller.CType_On)
	if onC == nil {
		return nil, errors.New("cannot find characteristic On")
	}

	var onValue bool
	withValOnC, err := dev.GetCharacteristic(acc.Id, onC.Iid)
	if err == nil {
		fmt.Println(withValOnC.Value, reflect.TypeOf(withValOnC.Value))
		onValue, ok = withValOnC.Value.(bool)
		if !ok {
			var onValInt float64
			onValInt, ok = withValOnC.Value.(float64)
			if ok {
				onValue = onValInt > 0
			}
		}
	} else {
		onValue, ok = onC.Value.(bool)
	}
	if !ok {
		onValue = false
	}

	return &LightBulb{
		acc:   acc,
		dev:   dev,
		th:    th,
		label: label,

		Bool: widget.Bool{Value: onValue},

		AppManager: am,
	}, nil
}

func (s *LightBulb) Layout(gtx C) D {

	if s.Bool.Changed() {
		fmt.Println("changed ", s.Bool.Value)
	}

	return applayout.DetailRow{}.Layout(gtx,
		material.Body1(s.th, s.label).Layout,
		material.Switch(s.th, &s.Bool, s.label).Layout,
	)
}
