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

type AccessoryInfo struct {
	widget.List

	manufacturer string
	model        string
	name         string
	serial       string
	fwrevision   string

	acc *hkontroller.Accessory
	dev *hkontroller.Device
	th  *material.Theme

	events <-chan emitter.Event

	*application.App
}

func NewAccessoryInfo(app *application.App, acc *hkontroller.Accessory, dev *hkontroller.Device) (*AccessoryInfo, error) {

	i := &AccessoryInfo{acc: acc, dev: dev, th: app.Theme, App: app}

	infoS := acc.GetService(hkontroller.SType_AccessoryInfo)
	if infoS == nil {
		return nil, errors.New("cannot get AccessoryInfo service")
	}

	if cc := infoS.GetCharacteristic(hkontroller.CType_Name); cc != nil {
		if vv, ok := cc.Value.(string); ok {
			i.name = vv
		}
	}

	if cc := infoS.GetCharacteristic(hkontroller.CType_Manufacturer); cc != nil {
		if vv, ok := cc.Value.(string); ok {
			i.manufacturer = vv
		}
	}

	if cc := infoS.GetCharacteristic(hkontroller.CType_Model); cc != nil {
		if val, ok := cc.Value.(string); ok {
			i.model = val
		}
	}

	if cc := infoS.GetCharacteristic(hkontroller.CType_SerialNumber); cc != nil {
		if val, ok := cc.Value.(string); ok {
			i.serial = val
		}
	}

	if cc := infoS.GetCharacteristic(hkontroller.CType_FirmwareRevision); cc != nil {
		if val, ok := cc.Value.(string); ok {
			i.fwrevision = val
		}
	}

	return i, nil
}

func (i *AccessoryInfo) SubscribeToEvents() {
}
func (i *AccessoryInfo) UnsubscribeFromEvents() {
}

func (i *AccessoryInfo) Layout(gtx C) D {
	i.List.Axis = layout.Vertical

	return widget.Border{
		Color: color.NRGBA{
			R: 0,
			G: 0,
			B: 0,
			A: 64,
		},
		Width:        unit.Dp(1),
		CornerRadius: unit.Dp(1),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.List(i.th, &i.List).Layout(gtx, 1, func(gtx C, _ int) D {
			return layout.Flex{
				Alignment: layout.Middle,
				Axis:      layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return applayout.DetailRow{PrimaryWidth: 0.5}.Layout(gtx,
						material.Body2(i.th, "Manufacturer").Layout,
						material.Body2(i.th, i.manufacturer).Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return applayout.DetailRow{PrimaryWidth: 0.5}.Layout(gtx,
						material.Body2(i.th, "Model").Layout,
						material.Body2(i.th, i.model).Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return applayout.DetailRow{PrimaryWidth: 0.5}.Layout(gtx,
						material.Body2(i.th, "Name").Layout,
						material.Body2(i.th, i.name).Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return applayout.DetailRow{PrimaryWidth: 0.5}.Layout(gtx,
						material.Body2(i.th, "SerialNumber").Layout,
						material.Body2(i.th, i.serial).Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return applayout.DetailRow{PrimaryWidth: 0.5}.Layout(gtx,
						material.Body2(i.th, "Firmware Revision").Layout,
						material.Body2(i.th, i.fwrevision).Layout,
					)
				}),
			)
		})
	})
}
