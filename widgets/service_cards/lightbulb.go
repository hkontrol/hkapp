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

type LightBulb struct {
	quick bool // simplified version to display in list of accs

	widget.Bool

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device
	th  *material.Theme

	events <-chan emitter.Event

	*application.App
}

func NewLightBulb(app *application.App,
	acc *hkontroller.Accessory,
	dev *hkontroller.Device,
	quickWidget bool) (*LightBulb, error) {

	l := &LightBulb{
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
	l.label = label

	lightbS := acc.GetService(hkontroller.SType_LightBulb)
	if lightbS == nil {
		return nil, errors.New("cannot find LightBulb service")
	}
	onC := lightbS.GetCharacteristic(hkontroller.CType_On)
	if onC == nil {
		return nil, errors.New("cannot find characteristic On")
	}

	convertOnValue := func(value interface{}, ll *LightBulb) {
		onValue, ok := value.(bool)
		if !ok {
			var onValInt float64
			onValInt, ok = value.(float64)
			if ok {
				onValue = onValInt > 0
			}
		}
		ll.Bool = widget.Bool{Value: onValue}
	}

	withValOnC, err := dev.GetCharacteristic(acc.Id, onC.Iid)
	if err == nil {
		convertOnValue(withValOnC.Value, l)
	} else {
		convertOnValue(onC.Value, l)
	}
	if !ok {
		convertOnValue(false, l)
	}

	return l, nil
}

func (l *LightBulb) SubscribeToEvents() {
	convertOnValue := func(value interface{}, sw *LightBulb) {
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
	lbS := l.acc.GetService(hkontroller.SType_LightBulb)
	if lbS == nil {
		return
	}
	onC := lbS.GetCharacteristic(hkontroller.CType_On)
	if onC == nil {
		return
	}

	onEvent := func(e emitter.Event) {
		value := e.Args[2]
		convertOnValue(value, l)
		l.App.Window.Invalidate()
	}
	events, err := l.dev.SubscribeToEvents(l.acc.Id, onC.Iid)
	if err != nil {
		return
	}
	l.events = events
	go func(evs <-chan emitter.Event) {
		for e := range evs {
			onEvent(e)
		}
	}(events)

	// hapEvents from GUI
	vals := l.App.OnValueChange(l.dev.Id, l.acc.Id, onC.Iid)
	go func(evs <-chan emitter.Event) {
		for e := range evs {
			onEvent(e)
		}
	}(vals)
}
func (l *LightBulb) UnsubscribeFromEvents() {
	lbS := l.acc.GetService(hkontroller.SType_LightBulb)
	if lbS == nil {
		return
	}
	onC := lbS.GetCharacteristic(hkontroller.CType_On)
	if onC == nil {
		return
	}
	l.dev.UnsubscribeFromEvents(l.acc.Id, onC.Iid, l.events)
}
func (l *LightBulb) QuickAction() {
	l.Bool.Value = !l.Bool.Value
	l.OnBoolValueChanged()
}

func (l *LightBulb) OnBoolValueChanged() error {

	srv := l.acc.GetService(hkontroller.SType_LightBulb)
	if srv == nil {
		return errors.New("cannot find LightBulb service")
	}
	chr := srv.GetCharacteristic(hkontroller.CType_On)
	if chr == nil {
		return errors.New("cannot find On characteristic")
	}

	err := l.dev.PutCharacteristic(l.acc.Id, chr.Iid, l.Bool.Value)
	if err != nil {
		return err
	}

	l.App.EmitValueChange(l.dev.Id, l.acc.Id, chr.Iid, l.Bool.Value)

	return nil
}

func (l *LightBulb) Layout(gtx C) D {

	if l.Bool.Changed() {
		l.OnBoolValueChanged()
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
		if l.quick {
			return material.Switch(l.th, &l.Bool, l.label).Layout(gtx)
		} else {
			return applayout.DetailRow{PrimaryWidth: 0.8}.Layout(gtx,
				material.Body1(l.th, l.label).Layout,
				material.Switch(l.th, &l.Bool, l.label).Layout)
		}
	})
}
