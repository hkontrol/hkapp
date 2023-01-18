package service_cards

import (
	"errors"
	"hkapp/applayout"
	"hkapp/application"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
)

type Switch struct {
	quick bool // simplified version to display in list of accs

	on      widget.Bool
	quickOn widget.Bool

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device

	hapEvents <-chan emitter.Event
	guiEvents <-chan emitter.Event

	th *material.Theme

	*application.App
}

func NewSwitch(app *application.App,
	acc *hkontroller.Accessory,
	dev *hkontroller.Device,
	quickWidget bool) (*Switch, error) {
	s := &Switch{
		quick: quickWidget,
		acc:   acc,
		dev:   dev,
		th:    app.Theme,
		App:   app,
	}

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
		sw.on = widget.Bool{Value: onValue}
		sw.quickOn = widget.Bool{Value: onValue}
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
		sw.on.Value = onValue
		sw.quickOn.Value = onValue
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

	// hapEvents from HAP
	events, err := s.dev.SubscribeToEvents(s.acc.Id, onC.Iid)
	if err != nil {
		return
	}
	s.hapEvents = events
	go func(evs <-chan emitter.Event) {
		for e := range evs {
			onEvent(e)
		}
	}(events)

	// hapEvents from GUI
	vals := s.App.OnValueChange(s.dev.Id, s.acc.Id, onC.Iid)
	s.guiEvents = vals
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

	s.dev.UnsubscribeFromEvents(s.acc.Id, onC.Iid, s.hapEvents)
	s.App.OffValueChange(s.dev.Id, s.acc.Id, onC.Iid, s.guiEvents)
}

func (s *Switch) QuickAction() {
	s.on.Value = !s.on.Value
	s.onBoolValueChanged()
}

func (s *Switch) onBoolValueChanged() error {

	srv := s.acc.GetService(hkontroller.SType_Switch)
	if srv == nil {
		return errors.New("cannot find SwitchService characteristic")
	}
	chr := srv.GetCharacteristic(hkontroller.CType_On)
	if chr == nil {
		return errors.New("cannot find On characteristic")
	}

	go func() {
		err := s.dev.PutCharacteristic(s.acc.Id, chr.Iid, s.on.Value)
		if err != nil {
			return
		}

		s.App.EmitValueChange(s.dev.Id, s.acc.Id, chr.Iid, s.on.Value)
	}()

	return nil
}

func (s *Switch) Layout(gtx C) D {

	var err error
	if s.on.Changed() {
		err = s.onBoolValueChanged()
	}

	if err != nil {
		return material.Body1(s.th, err.Error()).Layout(gtx)
	}

	return widget.Border{
		Color: color.NRGBA{
			R: 0,
			G: 0,
			B: 0,
			A: 0,
		},
		Width:        unit.Dp(1),
		CornerRadius: unit.Dp(1),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if s.quick {
			return material.Switch(s.th, &s.quickOn, s.label).Layout(gtx)
		} else {
			return applayout.DetailRow{PrimaryWidth: 0.8}.Layout(gtx,
				material.Body1(s.th, s.label).Layout,
				material.Switch(s.th, &s.on, s.label).Layout)
		}
	})
}
