package service_cards

import (
	"errors"
	"fmt"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"hkapp/applayout"
	"hkapp/application"
	"reflect"
)

type LightBulb struct {
	widget.Bool

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device
	th  *material.Theme

	*application.App
}

func NewLightBulb(app *application.App, acc *hkontroller.Accessory, dev *hkontroller.Device, th *material.Theme) (*LightBulb, error) {

	l := &LightBulb{acc: acc, dev: dev, th: th, App: app}

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
		fmt.Println(withValOnC.Value, reflect.TypeOf(withValOnC.Value))

		convertOnValue(withValOnC.Value, l)
	} else {
		convertOnValue(onC.Value, l)
	}
	if !ok {
		convertOnValue(false, l)
	}
	dev.SubscribeToEvents(acc.Id, withValOnC.Iid, func(aid uint64, iid uint64, value interface{}) {
		convertOnValue(value, l)
	})

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

	l.dev.SubscribeToEvents(l.acc.Id, onC.Iid, func(aid uint64, iid uint64, value interface{}) {
		convertOnValue(value, l)
		l.App.Window.Invalidate()
	})
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
	l.dev.UnsubscribeFromEvents(l.acc.Id, onC.Iid)
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
