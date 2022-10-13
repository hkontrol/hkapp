package accessories

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/hkontrol/hkontroller"
	alo "hkapp/applayout"
	"hkapp/appmanager"

	"hkapp/icon"
	page "hkapp/pages"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type DeviceAccPair struct {
	device    *hkontroller.Device
	accessory *hkontroller.Accessory
}

// Page holds the state for a page demonstrating the features of
// the NavDrawer component.
type Page struct {
	widget.List

	accs         []DeviceAccPair
	accsSwitches []widget.Bool

	*page.Router
	*appmanager.AppManager
}

// New constructs a Page with the provided router.
func New(router *page.Router, manager *appmanager.AppManager) *Page {
	return &Page{
		Router:     router,
		AppManager: manager,
	}
}

var _ page.Page = &Page{}

func (p *Page) Update() {
	connections := p.AppManager.GetVerifiedDevices()
	p.accs = make([]DeviceAccPair, 0, len(connections))
	for _, c := range connections {
		err := c.DiscoverAccessories()
		if err != nil {
			fmt.Println("discover accs err: ", err)
			continue
		}
		accs := c.GetAccessories()
		for _, a := range accs {
			p.accs = append(p.accs, DeviceAccPair{device: c, accessory: a})
		}
	}
	p.accsSwitches = make([]widget.Bool, len(p.accs))
}

func (p *Page) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
}

func (p *Page) Overflow() []component.OverflowAction {
	return []component.OverflowAction{}
}

func (p *Page) NavItem() component.NavItem {
	return component.NavItem{
		Name: "whatsyourask",
		Icon: icon.RestaurantMenuIcon,
	}
}

func (p *Page) Layout(gtx C, th *material.Theme) D {

	for i := range p.accsSwitches {
		if p.accsSwitches[i].Changed() {
			val := p.accsSwitches[i].Value
			accdev := p.accs[i]
			acc := accdev.accessory
			dev := accdev.device

			var on *hkontroller.CharacteristicDescription
			lb := acc.GetService(hkontroller.SType_LightBulb)
			sw := acc.GetService(hkontroller.SType_Switch)
			if lb != nil {
				on = lb.GetCharacteristic(hkontroller.CType_On)
			} else if sw != nil {
				on = sw.GetCharacteristic(hkontroller.CType_On)
			}
			err := dev.PutCharacteristic(acc.Id, on.Iid, val)
			if err == nil {
				on.Value = val
			}
		}
	}

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
						acc := accdev.accessory
						sr := acc.GetService(hkontroller.SType_AccessoryInfo)
						if sr == nil {
							return D{}
						}
						ch := sr.GetCharacteristic(hkontroller.CType_Name)
						if ch == nil {
							return D{}
						}

						name, ok := ch.Value.(string)
						if !ok {
							name = "undefined"
						}

						var on *hkontroller.CharacteristicDescription
						lb := acc.GetService(hkontroller.SType_LightBulb)
						sw := acc.GetService(hkontroller.SType_Switch)
						if lb != nil {
							on = lb.GetCharacteristic(hkontroller.CType_On)
						} else if sw != nil {
							on = sw.GetCharacteristic(hkontroller.CType_On)
						}

						if on != nil {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return alo.DetailRow{PrimaryWidth: 0.75}.Layout(gtx,
										material.Body1(th, name).Layout,
										func(gtx C) D {
											val, ok := on.Value.(bool)
											if ok {
												p.accsSwitches[i].Value = val
											}
											return material.Switch(th, &p.accsSwitches[i], name).Layout(gtx)
										})
								}),
							)
						}

						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx C) D {
								return material.Body1(th, name).Layout(gtx)
							}),
						)
					})
				})
		}),
	)
}
