package discover

import (
	"context"
	"fmt"
	"hkapp/application"
	"hkapp/icon"
	page "hkapp/pages"
	"image/color"
	"log"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/hkontrol/hkontroller"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Page holds the state for a page demonstrating the features of
// the AppBar component.
type Page struct {
	widget.List

	devs        []*hkontroller.Device
	devClicks   []widget.Clickable
	devSelected int
	pinInput    component.TextField
	btnPair     widget.Clickable
	btnVerify   widget.Clickable
	btnUnpair   widget.Clickable
	btnCancel   widget.Clickable
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
	return []component.OverflowAction{}
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

	// reset selected
	p.devSelected = -1
}

func (p *Page) Layout(gtx C, th *material.Theme) D {

	for i := range p.devClicks {
		if p.devClicks[i].Clicked() {
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
		go func() {
			err := dev.PairSetupAndVerify(context.TODO(), pin, 5*time.Second)
			if err != nil {
				log.Println("pairErr: ", err)
			}
			p.Update()
		}()
	}
	if p.btnVerify.Clicked() {
		dev := p.devs[p.devSelected]
		go func() {
			err := dev.PairSetupAndVerify(context.TODO(), "", 5*time.Second)
			if err != nil {
				log.Println("verifyErr: ", err)
			}
			p.Update()
		}()
	}
	if p.btnUnpair.Clicked() {
		dev := p.devs[p.devSelected]
		_ = dev.Unpair()
		p.Update()
	}
	if p.btnCancel.Clicked() {
		dev := p.devs[p.devSelected]
		dev.CancelPersistConnection()
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
						dev := p.devs[i]

						nameStr := fmt.Sprintf("%s", dev.Name)
						nameStyle := material.Label(th, unit.Sp(20), nameStr)
						idStr := fmt.Sprintf("%s", dev.Id)
						idStyle := material.Label(th, unit.Sp(16), idStr)
						idStyle.Font.Variant = "Mono"

						stateStr := ""
						if !dev.IsPaired() && dev.IsDiscovered() {
							stateStr = "discovered"
						} else if dev.IsPaired() && dev.IsVerified() {
							stateStr = "verified"
						} else if dev.IsPaired() && !dev.IsDiscovered() {
							stateStr = "paired, not discovered"
						} else if dev.IsPaired() && dev.IsDiscovered() && dev.IsVerifying() {
							stateStr = "establishing encrypted session"
						} else if dev.CloseReason() != nil {
							stateStr = dev.CloseReason().Error()
						}
						stateStyle := material.Label(th, unit.Sp(16), stateStr)

						return layout.UniformInset(unit.Dp(4)).Layout(gtx,
							func(gtx C) D {
								return widget.Border{
									Color: color.NRGBA{
										R: 0,
										G: 0,
										B: 0,
										A: 64,
									},
									Width:        unit.Dp(1),
									CornerRadius: unit.Dp(3),
								}.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(unit.Dp(8)).Layout(gtx,
										func(gtx C) D {
											if p.devSelected < 0 || i != p.devSelected {
												return material.Clickable(gtx, &p.devClicks[i],
													func(gtx layout.Context) layout.Dimensions {
														return layout.Flex{
															Axis: layout.Vertical,
														}.Layout(gtx,
															layout.Rigid(func(gtx layout.Context) layout.Dimensions {
																return nameStyle.Layout(gtx)
															}),
															layout.Rigid(func(gtx layout.Context) layout.Dimensions {
																return idStyle.Layout(gtx)
															}),
															layout.Rigid(func(gtx layout.Context) layout.Dimensions {
																return stateStyle.Layout(gtx)
															}),
														)
													})
											} else {
												if !dev.IsPaired() && dev.IsDiscovered() {
													return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], nameStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return p.pinInput.Layout(gtx, th, "pin")
														}),
														layout.Rigid(func(gtx C) D {
															return material.Button(th, &p.btnPair, "pair").Layout(gtx)
														}),
													)
												} else if dev.IsPaired() && dev.IsVerified() {
													return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], nameStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], idStyle.Layout)
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
												} else if dev.IsPaired() && !dev.IsDiscovered() {
													return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], nameStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], idStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Label(th, unit.Sp(16), "this one is paired but not discovered").Layout(gtx)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Button(th, &p.btnUnpair, "unpair, take care").Layout(gtx)
														}),
													)
												} else if dev.IsPaired() && dev.IsDiscovered() && dev.IsVerifying() && !dev.IsVerified() {

													return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], nameStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], idStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Label(th, unit.Sp(16), "trying to /pair-verify device").Layout(gtx)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Button(th, &p.btnCancel, "cancel").Layout(gtx)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Button(th, &p.btnUnpair, "unpair, take care").Layout(gtx)
														}),
													)
												} else if dev.CloseReason() != nil {
													return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], nameStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], idStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Label(th, unit.Sp(16), dev.CloseReason().Error()).Layout(gtx)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Button(th, &p.btnVerify, "pair-verify").Layout(gtx)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Button(th, &p.btnUnpair, "unpair, take care").Layout(gtx)
														}),
													)
												} else {
													log.Println("wtf is discovered?", dev.IsDiscovered())
													log.Println("wtf is paired?", dev.IsPaired())
													log.Println("wtf is verified?", dev.IsVerified())
													return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], nameStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Clickable(gtx, &p.devClicks[i], idStyle.Layout)
														}),
														layout.Rigid(func(gtx C) D {
															return material.Label(th, unit.Sp(16), "wtf?").Layout(gtx)
														}),
													)
												}
											}
										})
								})
							})
					})
				})
		}),
	)
}
