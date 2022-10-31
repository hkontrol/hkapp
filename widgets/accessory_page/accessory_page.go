package accessory_page

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"hkapp/appmanager"
	"hkapp/widgets/service_cards"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type AccessoryPage struct {
	acc *hkontroller.Accessory
	dev *hkontroller.Device

	*material.Theme

	widget.List

	srvwidgets []interface {
		Layout(C) D
	}

	*appmanager.AppManager
}

func NewAccessoryPage(am *appmanager.AppManager, acc *hkontroller.Accessory, dev *hkontroller.Device, th *material.Theme) *AccessoryPage {

	ap := AccessoryPage{
		acc:   acc,
		dev:   dev,
		Theme: th,
	}

	ap.srvwidgets =
		make([]interface {
			Layout(C) D
		}, 0, len(acc.Ss))

	for _, s := range acc.Ss {
		w, err := service_cards.GetWidgetForService(am, acc, dev, s, th)
		if err != nil {
			continue
		}
		ap.srvwidgets = append(ap.srvwidgets, w)
	}

	return &ap
}

func (p *AccessoryPage) Layout(gtx C) D {
	// here we loop through all the events associated with this button.
	p.List.Axis = layout.Vertical
	listStyle := material.List(p.Theme, &p.List)
	return listStyle.Layout(gtx, len(p.srvwidgets), func(gtx C, i int) D {
		return p.srvwidgets[i].Layout(gtx)
	})
}
