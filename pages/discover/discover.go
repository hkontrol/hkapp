package discover

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/hkontrol/hkontroller"
	"hkapp/appmanager"
	"hkapp/icon"
	page "hkapp/pages"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Page holds the state for a page demonstrating the features of
// the AppBar component.
type Page struct {
	exampleOverflowState widget.Clickable

	widget.List

	devs        []*hkontroller.Device
	devClicks   []widget.Clickable
	devSelected int
	pinInput    widget.Editor
	btnPair     widget.Clickable
	btnUnpair   widget.Clickable
	btnVerify   widget.Clickable
	pairErr     error

	*page.Router
	*appmanager.AppManager
}

// New constructs a Page with the provided router.
func New(router *page.Router, app *appmanager.AppManager) *Page {
	return &Page{
		Router:      router,
		AppManager:  app,
		devSelected: -1,
	}
}

var _ page.Page = &Page{}

func (p *Page) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
}

func (p *Page) Overflow() []component.OverflowAction {
	return []component.OverflowAction{
		{
			Name: "Example 1",
			Tag:  &p.exampleOverflowState,
		},
		{
			Name: "Example 2",
			Tag:  &p.exampleOverflowState,
		},
	}
}

func (p *Page) NavItem() component.NavItem {
	return component.NavItem{
		Name: "helloriend",
		Icon: icon.HomeIcon,
	}
}

const (
	settingNameColumnWidth    = .3
	settingDetailsColumnWidth = 1 - settingNameColumnWidth
)

func (p *Page) Update() {
	p.devs = p.AppManager.GetDevices()
	p.devClicks = make([]widget.Clickable, len(p.devs))

	for _, d := range p.devs {
		fmt.Println(d.Id, " paired? ", d.IsPaired(), "; verified? ", d.IsVerified())
	}

	// reset selected
	p.devSelected = -1
}

func (p *Page) Layout(gtx C, th *material.Theme) D {

	for i := range p.devClicks {
		if p.devClicks[i].Clicked() {
			fmt.Println("clicked ", i, p.devs[i].Id)
			p.pinInput.SetText("")
			p.pairErr = nil
			if p.devSelected == i {
				p.devSelected = -1
			} else {
				p.devSelected = i
			}
		}
	}

	if p.btnPair.Clicked() {
		pin := p.pinInput.Text()
		devId := p.devs[p.devSelected].Id
		fmt.Println("btnPair: ", devId, pin)
		err := p.AppManager.PairSetupAndVerify(devId, pin)
		if err != nil {
			_ = p.AppManager.UnpairDevice(p.devs[p.devSelected])
		}
		fmt.Println("pairErr: ", err)
		p.Update()
	}
	if p.btnUnpair.Clicked() {
		dev := p.devs[p.devSelected]
		fmt.Println("btnUnpair: ", dev.Id)
		_ = p.AppManager.UnpairDevice(dev)
		p.Update()
	}
	if p.btnVerify.Clicked() {
		dev := p.devs[p.devSelected]
		fmt.Println("btnVerify: ", dev.Id)
		err := p.AppManager.PairVerify(dev.Id)
		if err != nil {
			_ = p.AppManager.UnpairDevice(dev)
		}
		p.Update()
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return (layout.Inset{Left: unit.Dp(6)}).Layout(gtx,
				func(gtx C) D {
					p.List.Axis = layout.Vertical
					listStyle := material.List(th, &p.List)
					return listStyle.Layout(gtx, len(p.devs), func(gtx C, i int) D {
						labelStyle := material.Label(th, unit.Sp(20), p.devs[i].Id)
						labelStyle.Font.Variant = "Mono"
						if p.devSelected < 0 {
							return material.Clickable(gtx, &p.devClicks[i], labelStyle.Layout)
						} else {
							if !p.devs[p.devSelected].IsPaired() && !p.devs[p.devSelected].IsVerified() {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return material.Clickable(gtx, &p.devClicks[i], labelStyle.Layout)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Editor(th, &p.pinInput, "enter pin here").Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Button(th, &p.btnPair, "pair").Layout(gtx)
									}),
								)
							} else if p.devs[p.devSelected].IsPaired() && p.devs[p.devSelected].IsVerified() {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return material.Clickable(gtx, &p.devClicks[i], labelStyle.Layout)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Label(th, unit.Sp(16), "this one paired and verified").Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Label(th, unit.Sp(16), "encrypted session established").Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Button(th, &p.btnUnpair, "unpair").Layout(gtx)
									}),
								)
							} else if p.devs[p.devSelected].IsPaired() && !p.devs[p.devSelected].IsVerified() {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return material.Clickable(gtx, &p.devClicks[i], labelStyle.Layout)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Label(th, unit.Sp(16), "this one is paired but unverified").Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Button(th, &p.btnVerify, "verify").Layout(gtx)
									}),
								)
							} else {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return material.Clickable(gtx, &p.devClicks[i], labelStyle.Layout)
									}),
									layout.Rigid(func(gtx C) D {
										return material.Label(th, unit.Sp(16), "wtf?").Layout(gtx)
									}),
								)
							}
						}
					})
				})
		}),
	)
}
