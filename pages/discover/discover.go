package discover

import (
	"context"
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
	"time"
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
	pinInput    component.TextField
	btnPair     widget.Clickable
	btnUnpair   widget.Clickable
	btnVerify   widget.Clickable
	pairErr     error

	*application.App
}

// New constructs a Page with the provided router.
func New(app *application.App) *Page {
	return &Page{
		App:         app,
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
		Name: "pairings",
		Icon: icon.HomeIcon,
	}
}

const (
	settingNameColumnWidth    = .3
	settingDetailsColumnWidth = 1 - settingNameColumnWidth
)

func (p *Page) Update() {
	p.devs = p.App.Manager.GetAllDevices()
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
		dev := p.devs[p.devSelected]
		fmt.Println("btnPair: ", dev.Id, pin)
		err := dev.PairSetupAndVerify(context.TODO(), pin, 5*time.Second)
		if err != nil {
			fmt.Println("pairErr: ", err)
			_ = dev.Unpair()
		}
		p.Update()
	}
	if p.btnUnpair.Clicked() {
		dev := p.devs[p.devSelected]
		fmt.Println("btnUnpair: ", dev.Id)
		_ = dev.Unpair()
		p.Update()
	}
	if p.btnVerify.Clicked() {
		dev := p.devs[p.devSelected]
		fmt.Println("btnVerify: ", dev.Id)
		err := dev.PairVerify()
		if err != nil {
			_ = dev.Unpair()
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
						if p.devSelected < 0 || i != p.devSelected {
							return material.Clickable(gtx, &p.devClicks[i], labelStyle.Layout)
						} else {
							if !p.devs[p.devSelected].IsPaired() && !p.devs[p.devSelected].IsVerified() {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return material.Clickable(gtx, &p.devClicks[i], labelStyle.Layout)
									}),
									layout.Rigid(func(gtx C) D {
										return p.pinInput.Layout(gtx, th, "pin")
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
