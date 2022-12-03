package accessories

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/hkontrol/hkontroller"
	"hkapp/application"
	"hkapp/icon"
	page "hkapp/pages"
	"hkapp/widgets"
	"hkapp/widgets/accessory_card"
	"hkapp/widgets/accessory_page"
	"time"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type DeviceAccPair struct {
	*hkontroller.Device
	*hkontroller.Accessory
}

// Page holds the state for a page demonstrating the features of
// the NavDrawer component.
type Page struct {
	widget.List

	accs []DeviceAccPair

	// clickable elements for cards
	//clickables []widget.Clickable
	clickables []widgets.LongClickable

	// for opened accessory
	closeSelectedAcc     widget.Clickable
	closeSelectedAccIcon widget.Clickable

	// index of selected accessory
	selectedAccIdx  int
	selectedAccPage interface {
		Layout(C) D
	}

	*application.App
}

// New constructs a Page with the provided router.
func New(app *application.App) *Page {
	return &Page{
		App:            app,
		selectedAccIdx: -1,
	}
}

var _ page.Page = &Page{}

func (p *Page) Update() {
	connections := p.App.Manager.GetVerifiedDevices()
	fmt.Println("connections: ", len(connections))
	p.accs = []DeviceAccPair{}
	for _, c := range connections {
		err := c.GetAccessories()
		if err != nil {
			fmt.Println("discover accs err: ", err)
			continue
		}

		accs := c.Accessories()
		for _, a := range accs {
			p.accs = append(p.accs, DeviceAccPair{Device: c, Accessory: a})
		}
	}
	p.clickables = make([]widgets.LongClickable, len(p.accs))
	for i := range p.clickables {
		p.clickables[i] = widgets.NewLongClickable(time.Second)
	}
}

func (p *Page) Actions() []component.AppBarAction {
	if p.selectedAccIdx > -1 {
		return []component.AppBarAction{
			component.SimpleIconAction(&p.closeSelectedAccIcon, icon.ExitIcon,
				component.OverflowAction{
					Name: "Close",
					Tag:  &p.closeSelectedAccIcon,
				},
			),
		}
	} else {
		return []component.AppBarAction{}
	}
}

func (p *Page) Overflow() []component.OverflowAction {
	return []component.OverflowAction{}
}

func (p *Page) NavItem() component.NavItem {
	return component.NavItem{
		Name: "accessories",
		Icon: icon.LightBulbIcon,
	}
}

func (p *Page) Layout(gtx C, th *material.Theme) D {

	for i := range p.clickables {
		for p.clickables[i].ShortClick() {
			fmt.Println("short click ", i)
		}
		for p.clickables[i].LongClick() {
			fmt.Println("long click ", i)
			p.selectedAccIdx = i
			accdev := p.accs[i]
			acc := accdev.Accessory
			dev := accdev.Device
			p.selectedAccPage = accessory_page.NewAccessoryPage(p.App, acc, dev, th)
			if ap, ok := p.selectedAccPage.(*accessory_page.AccessoryPage); ok {
				ap.SubscribeToEvents()
			}
			p.App.Router.AppBar.SetActions(p.Actions(), p.Overflow())
			p.App.Window.Invalidate()
		}
	}
	for p.closeSelectedAcc.Clicked() || p.closeSelectedAccIcon.Clicked() {
		fmt.Println("close selected acc")
		p.selectedAccIdx = -1
		if ap, ok := p.selectedAccPage.(*accessory_page.AccessoryPage); ok {
			ap.UnsubscribeFromEvents()
		}
		p.selectedAccPage = nil
		p.App.Router.AppBar.SetActions(p.Actions(), p.Overflow())
		p.App.Window.Invalidate()
	}

	if p.selectedAccIdx < 0 {
		// all accessories
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return (layout.Inset{Left: unit.Dp(6)}).Layout(gtx,
					func(gtx C) D {
						p.List.Axis = layout.Vertical

						listStyle := material.List(th, &p.List)

						return listStyle.Layout(gtx, len(p.accs), func(gtx C, i int) D {

							accdev := p.accs[i]
							acc := accdev.Accessory

							var children []layout.Widget
							w := accessory_card.NewAccessoryCard(th, &p.clickables[i], acc).Layout
							children = append(children, w)

							var flexChildren []layout.FlexChild
							for _, w := range children {
								flexChildren = append(flexChildren, layout.Rigid(w))
							}

							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								flexChildren...,
							)
						})
					})
			}))
	} else {
		// if accessory selected

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return (layout.Inset{Left: unit.Dp(6)}).Layout(gtx,
					func(gtx C) D {
						p.List.Axis = layout.Vertical

						listStyle := material.List(th, &p.List)

						return listStyle.Layout(gtx, 3, func(gtx C, i int) D {
							if i == 0 {
								return p.selectedAccPage.Layout(gtx)
							} else if i == 1 {
								return layout.Spacer{Height: unit.Dp(25)}.Layout(gtx)
							} else {
								return material.Button(th, &p.closeSelectedAcc, "close").Layout(gtx)
							}
						})
					})
			}),
		)
	}
}
