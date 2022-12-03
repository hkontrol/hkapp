package accessory_card

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"hkapp/application"
	"hkapp/widgets"
	"hkapp/widgets/service_cards"
	"image/color"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type AccessoryCard struct {
	clickable *widgets.LongClickable

	acc *hkontroller.Accessory
	dev *hkontroller.Device
	th  *material.Theme

	// primary service
	primary       *hkontroller.ServiceDescription
	primaryWidget interface{ Layout(C) D }

	*application.App
}

func NewAccessoryCard(app *application.App, acc *hkontroller.Accessory, dev *hkontroller.Device, clickable *widgets.LongClickable) *AccessoryCard {

	var primary *hkontroller.ServiceDescription
	// find primary service
	for _, srv := range acc.Ss {
		// select primary
		if srv.Primary != nil {
			if *srv.Primary {
				primary = srv
			}
		}
	}
	// primary service was not selected
	// then we select first, but not accessory info
	if primary == nil {
		for _, srv := range acc.Ss {
			if srv.Type == hkontroller.SType_AccessoryInfo {
				continue
			}
			primary = srv
			break
		}
	}
	var primaryWidget interface{ Layout(C) D }
	if primary != nil {
		// TODO: GetQuickWidgetForService
		// 		 so widgets for full version and quick version differs
		w, err := service_cards.GetWidgetForService(app, acc, dev, primary)
		if err == nil {
			primaryWidget = w
		}
	}

	return &AccessoryCard{
		clickable:     clickable,
		acc:           acc,
		dev:           dev,
		primary:       primary,
		primaryWidget: primaryWidget,
		App:           app,
		th:            app.Theme,
	}
}

func (s *AccessoryCard) Layout(gtx C) D {
	label := "---"
	info := s.acc.GetService(hkontroller.SType_AccessoryInfo)
	if info != nil {
		cname := info.GetCharacteristic(hkontroller.CType_Name)
		if cname != nil {
			label, _ = cname.Value.(string)
		}
	}

	var cardWidgets []layout.FlexChild
	cardWidgets = append(cardWidgets,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Body1(s.th, label).Layout(gtx)
		}),
	)

	servicesStr := ""
	for _, srv := range s.acc.Ss {
		servicesStr += srv.Type.String() + ", "
	}
	cardWidgets = append(cardWidgets,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			srvLabel := material.Body2(s.th, servicesStr)
			return srvLabel.Layout(gtx)
		}))

	if s.primaryWidget != nil {
		cardWidgets = append(cardWidgets, layout.Rigid(s.primaryWidget.Layout))
	}

	content := func(gtx C) D {
		return widget.Border{
			Color: color.NRGBA{
				R: 0,
				G: 0,
				B: 0,
				A: 255,
			},
			Width:        unit.Dp(1),
			CornerRadius: unit.Dp(3),
		}.Layout(gtx, func(gtx C) D {
			return layout.Inset{
				Top:    4,
				Bottom: 4,
				Left:   4,
				Right:  4,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					cardWidgets...,
				)
			})
		})
	}

	return s.clickable.Layout(gtx, content)
}

func (s *AccessoryCard) SubscribeToEvents() {
	// dangerous:
	type evented interface {
		SubscribeToEvents()
		UnsubscribeFromEvents()
	}
	if pp, ok := s.primaryWidget.(evented); ok {
		pp.SubscribeToEvents()
	}
}
func (s *AccessoryCard) UnsubscribeFromEvents() {
	type evented interface {
		SubscribeToEvents()
		UnsubscribeFromEvents()
	}
	if pp, ok := s.primaryWidget.(evented); ok {
		pp.UnsubscribeFromEvents()
	}
}

func (s *AccessoryCard) QuickActionSupported() bool {

	primary := s.primary
	if primary == nil {
		return false
	}

	type withQuickAction interface {
		QuickAction()
	}
	_, ok := s.primaryWidget.(withQuickAction)
	return ok
}

func (s *AccessoryCard) TriggerQuickAction() {
	primary := s.primary
	if primary == nil {
		return
	}
	type withQuickAction interface {
		QuickAction()
	}
	q, ok := s.primaryWidget.(withQuickAction)
	if !ok {
		return
	}
	q.QuickAction()
}
