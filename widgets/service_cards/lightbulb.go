package service_cards

import (
	"errors"
	"fmt"
	"hkapp/applayout"
	"hkapp/application"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
)

const brightnessDragDelay = 300 * time.Millisecond

type LightBulb struct {
	quick bool // simplified version to display in list of accs

	on               widget.Bool
	quickOn          widget.Bool // for quick widget
	brightnessWidget widget.Float
	brightnessValue  float32

	dragTimer *time.Timer

	chars map[hkontroller.HapCharacteristicType]*hkontroller.CharacteristicDescription

	hapEvents map[hkontroller.HapCharacteristicType]<-chan emitter.Event
	guiEvents map[hkontroller.HapCharacteristicType]<-chan emitter.Event

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device
	th  *material.Theme

	*application.App
}

func NewLightBulb(app *application.App,
	acc *hkontroller.Accessory,
	dev *hkontroller.Device,
	quickWidget bool) (*LightBulb, error) {

	l := &LightBulb{
		quick:     quickWidget,
		acc:       acc,
		dev:       dev,
		th:        app.Theme,
		App:       app,
		chars:     make(map[hkontroller.HapCharacteristicType]*hkontroller.CharacteristicDescription),
		hapEvents: make(map[hkontroller.HapCharacteristicType]<-chan emitter.Event),
		guiEvents: make(map[hkontroller.HapCharacteristicType]<-chan emitter.Event),
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
	} else {
		l.chars[hkontroller.CType_On] = onC
		withValC, err := dev.GetCharacteristic(acc.Id, onC.Iid)
		if err == nil {
			l.onValue(withValC.Value, hkontroller.CType_On)
		}
	}

	// optional characteristics
	brightnessC := lightbS.GetCharacteristic(hkontroller.CType_Brightness)
	if brightnessC != nil {
		l.chars[hkontroller.CType_Brightness] = brightnessC
		withValC, err := dev.GetCharacteristic(acc.Id, brightnessC.Iid)
		if err == nil {
			l.onValue(withValC.Value, hkontroller.CType_Brightness)
		}
	}
	hueC := lightbS.GetCharacteristic(hkontroller.CType_Hue)
	if hueC != nil {
		l.chars[hkontroller.CType_Hue] = hueC
	}
	satC := lightbS.GetCharacteristic(hkontroller.CType_Saturation)
	if hueC != nil {
		l.chars[hkontroller.CType_Saturation] = satC
	}
	ctempC := lightbS.GetCharacteristic(hkontroller.CType_ColorTemperature)
	if hueC != nil {
		l.chars[hkontroller.CType_ColorTemperature] = ctempC
	}

	return l, nil
}

func (l *LightBulb) onValue(value interface{}, ctype hkontroller.HapCharacteristicType) {
	if ctype == hkontroller.CType_On {
		onValue, ok := value.(bool)
		if !ok {
			var onValInt float64
			onValInt, ok = value.(float64)
			if ok {
				onValue = onValInt > 0
			}
		}
		l.on.Value = onValue
		l.quickOn.Value = onValue
	}
	if ctype == hkontroller.CType_Brightness {
		if valInt, ok := value.(int); ok {
			l.brightnessWidget.Value = float32(valInt)
		} else if valF32, ok := value.(float32); ok {
			l.brightnessWidget.Value = valF32
		} else if valF64, ok := value.(float64); ok {
			l.brightnessWidget.Value = float32(valF64)
		}
	}
}

func (l *LightBulb) SubscribeToEvents() {

	var err error
	var ev <-chan emitter.Event
	devId := l.dev.Name
	aid := l.acc.Id
	onEvent := func(e emitter.Event, ctype hkontroller.HapCharacteristicType) {
		value := e.Args[2]
		l.onValue(value, ctype)
		l.App.Window.Invalidate()
	}
	for ctype, cdescr := range l.chars {
		iid := cdescr.Iid
		ev, err = l.dev.SubscribeToEvents(aid, iid)
		if err != nil {
			fmt.Println("err subscribing: ", cdescr.Type.String(), err)
			continue
		}
		go func(evs <-chan emitter.Event, ct hkontroller.HapCharacteristicType) {
			for e := range evs {
				onEvent(e, ct)
			}
		}(ev, ctype)
		l.hapEvents[ctype] = ev

		ev = l.App.OnValueChange(devId, aid, iid)
		l.guiEvents[ctype] = ev
		go func(evs <-chan emitter.Event, ct hkontroller.HapCharacteristicType) {
			for e := range evs {
				onEvent(e, ct)
			}
		}(ev, ctype)
	}
}
func (l *LightBulb) UnsubscribeFromEvents() {
	aid := l.acc.Id
	devId := l.dev.Name
	for ctype, ee := range l.hapEvents {
		iid := l.chars[ctype].Iid
		err := l.dev.UnsubscribeFromEvents(aid, iid, ee)
		if err != nil {
			continue
		}
		delete(l.hapEvents, ctype)
	}
	for ctype, ee := range l.guiEvents {
		iid := l.chars[ctype].Iid
		l.App.OffValueChange(devId, aid, iid, ee)
		delete(l.hapEvents, ctype)
	}
	return
}

func (l *LightBulb) QuickAction() {
	l.on.Value = !l.on.Value
	l.onBoolValueChanged()
}

func (l *LightBulb) onBoolValueChanged() error {

	chr := l.chars[hkontroller.CType_On]
	go func() {
		err := l.dev.PutCharacteristic(l.acc.Id, chr.Iid, l.on.Value)
		if err != nil {
			return
		}

		l.App.EmitValueChange(l.dev.Name, l.acc.Id, chr.Iid, l.on.Value)
	}()

	return nil
}
func (l *LightBulb) onBrightnessSlider() error {
	if l.dragTimer != nil {
		l.dragTimer.Stop()
	}
	l.dragTimer = time.AfterFunc(brightnessDragDelay, func() {
		chr := l.chars[hkontroller.CType_Brightness]
		val := math.Floor(float64(l.brightnessWidget.Value))
		go func() {
			l.dev.PutCharacteristic(l.acc.Id, chr.Iid, val)
			l.App.EmitValueChange(l.dev.Name, l.acc.Id, chr.Iid, l.brightnessWidget.Value)
		}()
	})

	return nil
}

func (l *LightBulb) Layout(gtx C) D {

	if l.brightnessWidget.Changed() {
		l.onBrightnessSlider()
	}

	if l.on.Changed() {
		l.onBoolValueChanged()
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
			var children []layout.FlexChild
			children = append(children,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Switch(l.th, &l.quickOn, l.label).Layout(gtx)
				}))
			if _, ok := l.chars[hkontroller.CType_Brightness]; ok {
				brStr := fmt.Sprintf(" B: %d",
					int(math.Floor(float64(l.brightnessWidget.Value))))
				children = append(children,
					layout.Rigid(material.Body1(l.th, brStr).Layout))
			}
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				children...,
			)

		} else {

			var children []layout.FlexChild
			children = append(children,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return applayout.DetailRow{PrimaryWidth: 0.8}.Layout(gtx,
						material.Body1(l.th, l.label).Layout,
						material.Switch(l.th, &l.on, l.label).Layout)
				}))

			if _, ok := l.chars[hkontroller.CType_Brightness]; ok {
				children = append(children,
					layout.Rigid(material.Body1(l.th, "Brightness").Layout))
				children = append(children,
					layout.Rigid(material.Slider(l.th, &l.brightnessWidget, 0, 100).Layout))
			}

			if _, ok := l.chars[hkontroller.CType_Hue]; ok {
				children = append(children,
					layout.Rigid(material.Body1(l.th, "supports Hue").Layout))
			}
			if _, ok := l.chars[hkontroller.CType_Saturation]; ok {
				children = append(children,
					layout.Rigid(material.Body1(l.th, "supports Saturation").Layout))
			}
			if _, ok := l.chars[hkontroller.CType_ColorTemperature]; ok {
				children = append(children,
					layout.Rigid(material.Body1(l.th, "supports ColorTemperature").Layout))
			}

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				children...,
			)
		}
	})
}
