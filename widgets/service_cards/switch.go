package service_cards

import (
	"errors"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
	"hkapp/applayout"
	"hkapp/application"
	"image/color"
)

type Switch struct {
	widget.Bool

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device

	events <-chan emitter.Event

	th *material.Theme

	*application.App
}

func NewSwitch(app *application.App, acc *hkontroller.Accessory, dev *hkontroller.Device) (*Switch, error) {
	s := &Switch{acc: acc, dev: dev, th: app.Theme, App: app}

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
	s.label = label

	switchS := acc.GetService(hkontroller.SType_Switch)
	if switchS == nil {
		return nil, errors.New("cannot find Switch service")
	}
	onC := switchS.GetCharacteristic(hkontroller.CType_On)
	if onC == nil {
		return nil, errors.New("cannot find characteristic On")
	}

	convertOnValue := func(value interface{}, sw *Switch) {
		onValue, ok := value.(bool)
		if !ok {
			var onValInt float64
			onValInt, ok = value.(float64)
			if ok {
				onValue = onValInt > 0
			}
		}
		sw.Bool = widget.Bool{Value: onValue}
	}
	withValOnC, err := dev.GetCharacteristic(acc.Id, onC.Iid)
	if err == nil {
		convertOnValue(withValOnC.Value, s)
	} else {
		convertOnValue(onC.Value, s)
	}
	if !ok {
		convertOnValue(false, s)
	}

	return s, nil
}

func (s *Switch) SubscribeToEvents() {
	convertOnValue := func(value interface{}, sw *Switch) {
		onValue, ok := value.(bool)
		if !ok {
			var onValInt float64
			onValInt, ok = value.(float64)
			if ok {
				onValue = onValInt > 0
			}
		}
		sw.Bool = widget.Bool{Value: onValue}
	}
	switchS := s.acc.GetService(hkontroller.SType_Switch)
	if switchS == nil {
		return
	}
	onC := switchS.GetCharacteristic(hkontroller.CType_On)
	if onC == nil {
		return
	}

	onEvent := func(e emitter.Event) {
		value := e.Args[2]
		convertOnValue(value, s)
		s.App.Window.Invalidate()
	}

	// events from HAP
	events, err := s.dev.SubscribeToEvents(s.acc.Id, onC.Iid)
	if err != nil {
		return
	}
	s.events = events
	go func(evs <-chan emitter.Event) {
		for e := range evs {
			onEvent(e)
		}
	}(events)

	// events from GUI
	vals := s.App.OnValueChange(s.dev.Id, s.acc.Id, onC.Iid)
	go func(evs <-chan emitter.Event) {
		for e := range evs {
			onEvent(e)
		}
	}(vals)
}

func (s *Switch) UnsubscribeFromEvents() {
	switchS := s.acc.GetService(hkontroller.SType_Switch)
	if switchS == nil {
		return
	}
	onC := switchS.GetCharacteristic(hkontroller.CType_On)
	if onC == nil {
		return
	}
	s.dev.UnsubscribeFromEvents(s.acc.Id, onC.Iid, s.events)
}

func (s *Switch) QuickAction() {
	s.Bool.Value = !s.Bool.Value
	s.OnBoolValueChanged()
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

	s.App.EmitValueChange(s.dev.Id, s.acc.Id, chr.Iid, s.Bool.Value)

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

	return widget.Border{
		Color: color.NRGBA{
			R: 0,
			G: 255,
			B: 0,
			A: 255,
		},
		Width:        unit.Dp(1),
		CornerRadius: unit.Dp(1),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return applayout.DetailRow{PrimaryWidth: 0.8}.Layout(gtx,
			material.Body1(s.th, s.label).Layout,
			material.Switch(s.th, &s.Bool, s.label).Layout,
		)
	})
}
