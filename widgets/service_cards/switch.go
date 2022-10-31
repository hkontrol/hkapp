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

type Switch struct {
	widget.Bool

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device

	th *material.Theme

	*appmanager.AppManager
}

func NewSwitch(am *appmanager.AppManager, acc *hkontroller.Accessory, dev *hkontroller.Device, th *material.Theme) (*Switch, error) {

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

	switchS := acc.GetService(hkontroller.SType_Switch)
	if switchS == nil {
		return nil, errors.New("cannot find Switch service")
	}
	onC := switchS.GetCharacteristic(hkontroller.CType_On)
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

	s := Switch{
		acc: acc,
		dev: dev,

		th:         th,
		label:      label,
		AppManager: am,
		Bool:       widget.Bool{Value: onValue},
	}
	return &s, nil
}

func (s *Switch) OnBoolValueChanged() error {

	srv := s.acc.GetService(hkontroller.SType_Switch)
	if srv == nil {
		return errors.New("cannot find SwitchService characteristic")
	}
	chr := srv.GetCharacteristic(hkontroller.CType_On)
	if chr == nil {
		return errors.New("cannot find On characteristic")
	}

	err := s.dev.PutCharacteristic(s.acc.Id, chr.Iid, s.Bool.Value)
	if err != nil {
		return err
	}

	return nil
}

func (s *Switch) Layout(gtx C) D {

	var err error
	if s.Bool.Changed() {
		err = s.OnBoolValueChanged()
	}

	if err != nil {
		return material.Body1(s.th, err.Error()).Layout(gtx)
	}

	return applayout.DetailRow{}.Layout(gtx,
		material.Body1(s.th, s.label).Layout,
		material.Switch(s.th, &s.Bool, s.label).Layout,
	)
}
