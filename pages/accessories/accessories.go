package accessories

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"hkapp/appmanager"
	"hkapp/icon"
	page "hkapp/pages"
	"hkapp/widgets/accessory_card"
	"hkapp/widgets/accessory_page"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Page holds the state for a page demonstrating the features of
// the NavDrawer component.
type Page struct {
	widget.List

	accs []appmanager.DeviceAccPair

	// clickable elements for cards
	clickables []widget.Clickable

	// for opened accessory
	closeSelectedAcc widget.Clickable

	// index of selected accessory
	selectedAccIdx  int
	selectedAccPage interface {
		Layout(C) D
	}

	*page.Router
	*appmanager.AppManager
}

// New constructs a Page with the provided router.
func New(router *page.Router, manager *appmanager.AppManager) *Page {
	return &Page{
		Router:         router,
		AppManager:     manager,
		selectedAccIdx: -1,
	}
}

var _ page.Page = &Page{}

func (p *Page) Update() {
	connections := p.AppManager.GetVerifiedDevices()
	p.accs = make([]appmanager.DeviceAccPair, 0, len(connections))
	for _, c := range connections {
		err := c.DiscoverAccessories()
		if err != nil {
			fmt.Println("discover accs err: ", err)
			continue
		}
		accs := c.GetAccessories()
		for _, a := range accs {
			p.accs = append(p.accs, appmanager.DeviceAccPair{Device: c, Accessory: a})
		}
	}
	p.clickables = make([]widget.Clickable, len(p.accs))
}

func (p *Page) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
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
		for p.clickables[i].Clicked() {
			fmt.Println("clicked ", i)
			p.selectedAccIdx = i
			accdev := p.accs[i]
			acc := accdev.Accessory
			dev := accdev.Device
			// TODO replace
			p.selectedAccPage = accessory_page.NewAccessoryPage(p.AppManager, acc, dev, th)
		}
	}
	for p.closeSelectedAcc.Clicked() {
		fmt.Println("close selected acc")
		p.selectedAccIdx = -1
		p.selectedAccPage = nil
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
				return p.selectedAccPage.Layout(gtx)
			}),
			layout.Rigid(
				// The height of the spacer is 25 Device independent pixels
				layout.Spacer{Height: unit.Dp(25)}.Layout,
			),
			layout.Rigid(func(gtx C) D {
				return material.Button(th, &p.closeSelectedAcc, "close").Layout(gtx)
			}),
		)
	}
}
