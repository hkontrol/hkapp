package accessories

import (
	"fmt"
	"hkapp/application"
	"hkapp/icon"
	page "hkapp/pages"
	"hkapp/widgets"
	"hkapp/widgets/accessory_card"
	"hkapp/widgets/accessory_page"
	"image/color"
	"sync"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gioui.org/x/outlay"
	"github.com/hkontrol/hkontroller"
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
	outlay.FlowWrap

	mu    sync.Mutex
	accs  []DeviceAccPair
	cards []*accessory_card.AccessoryCard

	selectedAccTags []string
	tagList         outlay.FlowWrap
	tagCtxAreas     []widgets.ContextArea
	tagCtxMenu      component.MenuState
	// last index for which context menu was open
	lastActiveTagIdx int

	// clickable elements for cards
	clickables    []widgets.LongClickable
	tagClickables []widget.Clickable
	tagInput      widget.Editor
	addTagClick   widget.Clickable
	tagSearchBtn  widget.Clickable
	tagRemoveBtn  widget.Clickable

	// for opened accessory
	closeSelectedAcc     widget.Clickable
	closeSelectedAccIcon widget.Clickable
	accSettingsIcon      widget.Clickable

	showSettings bool

	// index of selected accessory
	selectedAccIdx  int
	selectedAccPage interface {
		Layout(C) D
	}

	th *material.Theme
	*application.App
}

// New constructs a Page with the provided router.
func New(app *application.App) *Page {
	p := &Page{
		App:            app,
		mu:             sync.Mutex{},
		th:             app.Theme,
		selectedAccIdx: -1,
		FlowWrap: outlay.FlowWrap{
			Axis:      layout.Horizontal,
			Alignment: layout.End,
		},
	}
	p.tagCtxMenu = component.MenuState{
		Options: []func(gtx C) D{
			func(gtx C) D {
				item := component.MenuItem(app.Theme, &p.tagSearchBtn, "Search")
				item.Icon = icon.VisibilityIcon
				item.Hint = component.MenuHintText(app.Theme, "")
				return item.Layout(gtx)
			},
			func(gtx C) D {
				item := component.MenuItem(app.Theme, &p.tagRemoveBtn, "Remove")
				item.Icon = icon.EditIcon
				item.Hint = component.MenuHintText(app.Theme, "")
				return item.Layout(gtx)
			},
		},
	}
	return p
}

var _ page.Page = &Page{}

func (p *Page) Update() {
	devices := p.App.Manager.GetVerifiedDevices()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.accs = []DeviceAccPair{}
	p.clickables = []widgets.LongClickable{}
	p.cards = []*accessory_card.AccessoryCard{}

	for _, d := range devices {
		accs := d.Accessories()
		for _, a := range accs {
			p.accs = append(p.accs, DeviceAccPair{Device: d, Accessory: a})
		}
	}
	p.clickables = make([]widgets.LongClickable, len(p.accs))
	p.cards = make([]*accessory_card.AccessoryCard, len(p.accs))
	for i, accdev := range p.accs {
		a := accdev.Accessory
		d := accdev.Device
		p.clickables[i] = widgets.NewLongClickable(500 * time.Millisecond)
		p.cards[i] = accessory_card.NewAccessoryCard(p.App, a, d, &p.clickables[i])
		p.cards[i].SubscribeToEvents()
	}
}

func (p *Page) Actions() []component.AppBarAction {
	if p.selectedAccIdx > -1 {
		return []component.AppBarAction{
			{
				OverflowAction: component.OverflowAction{
					Name: "Settigns",
					Tag:  &p.accSettingsIcon,
				},
				Layout: func(gtx layout.Context, bg, fg color.NRGBA) layout.Dimensions {
					for p.accSettingsIcon.Clicked() {
						p.showSettings = !p.showSettings
					}
					btn := component.SimpleIconButton(bg, fg,
						&p.accSettingsIcon, icon.SettingsIcon)
					btn.Background = bg
					if p.showSettings {
						btn.Color = color.NRGBA{R: 200, A: 128}
					} else {
						btn.Color = fg
					}
					return btn.Layout(gtx)
				},
			},
			component.SimpleIconAction(
				&p.closeSelectedAccIcon,
				icon.ExitIcon,
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

func (p *Page) openAccPage(i int) {
	p.selectedAccIdx = i
	accdev := p.accs[i]
	acc := accdev.Accessory
	dev := accdev.Device

	meta, err := p.App.Load(dev.Id, acc.Id)
	if err == nil {
		if tt, ok := meta["tags"]; ok {
			p.selectedAccTags = tt
		} else {
			p.selectedAccTags = []string{}
		}
	} else {
		p.selectedAccTags = []string{}
	}
	p.tagClickables = make([]widget.Clickable, len(p.selectedAccTags))
	p.tagCtxAreas = make([]widgets.ContextArea, len(p.selectedAccTags))
	for i := range p.tagCtxAreas {
		p.tagCtxAreas[i].LongPressDuration = 500 * time.Millisecond
	}

	p.selectedAccPage = accessory_page.NewAccessoryPage(p.App, acc, dev, p.th)
	if ap, ok := p.selectedAccPage.(*accessory_page.AccessoryPage); ok {
		ap.SubscribeToEvents()
	}
	p.showSettings = false

	p.App.Router.AppBar.SetActions(p.Actions(), p.Overflow())
	p.App.Window.Invalidate()
}

func (p *Page) Layout(gtx C, th *material.Theme) D {

	p.mu.Lock()
	for i := range p.clickables {

		for p.clickables[i].ShortClick() {
			card := p.cards[i]
			if card.QuickActionSupported() {
				card.TriggerQuickAction()
			} else {
				p.openAccPage(i)
			}
		}

		for p.clickables[i].LongClick() {
			p.openAccPage(i)
		}

	}
	p.mu.Unlock()

	for p.closeSelectedAcc.Clicked() || p.closeSelectedAccIcon.Clicked() {
		p.selectedAccIdx = -1
		if ap, ok := p.selectedAccPage.(*accessory_page.AccessoryPage); ok {
			ap.UnsubscribeFromEvents()
		}
		p.selectedAccPage = nil
		p.App.Router.AppBar.SetActions(p.Actions(), p.Overflow())
		p.App.Window.Invalidate()
	}

	for p.addTagClick.Clicked() {
		t := p.tagInput.Text()
		p.tagInput.SetText("")
		if t == "" {
			continue
		}

		found := false
		for _, tt := range p.selectedAccTags {
			if tt == t {
				found = true
				break
			}
		}
		if found {
			continue
		}

		p.selectedAccTags = append(p.selectedAccTags, t)
		p.tagClickables = make([]widget.Clickable, len(p.selectedAccTags))
		p.tagCtxAreas = make([]widgets.ContextArea, len(p.selectedAccTags))
		for i := range p.tagCtxAreas {
			p.tagCtxAreas[i].LongPressDuration = 500 * time.Millisecond
		}

		meta := make(map[string][]string)
		meta["tags"] = p.selectedAccTags
		accdev := p.accs[p.selectedAccIdx]
		acc := accdev.Accessory
		dev := accdev.Device
		p.App.Save(dev.Id, acc.Id, meta)
	}

	for i, tc := range p.tagClickables {
		for tc.Clicked() {
			fmt.Println("clicked tag: ", p.selectedAccTags[i])
		}
	}
	for p.tagRemoveBtn.Clicked() {
		i := p.lastActiveTagIdx

		if len(p.selectedAccTags) == 1 {
			p.selectedAccTags = []string{}
		} else {
			p.selectedAccTags = append(p.selectedAccTags[:i], p.selectedAccTags[i+1:]...)
		}
		p.tagClickables = make([]widget.Clickable, len(p.selectedAccTags))
		p.tagCtxAreas = make([]widgets.ContextArea, len(p.selectedAccTags))
		for i := range p.tagCtxAreas {
			p.tagCtxAreas[i].LongPressDuration = 500 * time.Millisecond
		}

		meta := make(map[string][]string)
		meta["tags"] = p.selectedAccTags
		accdev := p.accs[p.selectedAccIdx]
		acc := accdev.Accessory
		dev := accdev.Device
		p.App.Save(dev.Id, acc.Id, meta)
	}

	for p.tagRemoveBtn.Clicked() {
		i := p.lastActiveTagIdx
		if i >= len(p.selectedAccTags) && i < 0 {
			continue
		}
		tag := p.selectedAccTags[i]
		_ = tag // TODO search
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
						listStyle := material.List(p.th, &p.List)

						p.mu.Lock()
						defer p.mu.Unlock()
						return listStyle.Layout(gtx, 1, func(gtx C, i int) D {
							return p.FlowWrap.Layout(gtx, len(p.accs), func(gtx C, i int) D {
								if i >= len(p.accs) {
									return D{}
								}

								var children []layout.Widget
								w := p.cards[i]
								children = append(children, w.Layout)

								var flexChildren []layout.FlexChild
								for _, w := range children {
									flexChildren = append(flexChildren, layout.Rigid(w))
								}

								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									flexChildren...,
								)
							})
						})
					})
			}))
	} else {
		// if accessory selected

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.UniformInset(unit.Dp(6)).Layout(gtx,
					func(gtx C) D {
						p.List.Axis = layout.Vertical
						listStyle := material.List(p.th, &p.List)

						p.tagList.Axis = layout.Horizontal

						var content []layout.Widget

						if p.showSettings {
							getTagList := func(gtx C) D {
								return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx C) D {
									return p.tagList.Layout(gtx, len(p.selectedAccTags), func(gtx C, j int) D {
										state := &p.tagCtxAreas[j]

										return layout.Stack{}.Layout(gtx,
											layout.Stacked(func(gtx C) D {
												return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx C) D {
													return p.tagClickables[j].Layout(gtx, func(gtx C) D {
														return material.Button(p.th,
															&p.tagClickables[j], p.selectedAccTags[j]).Layout(gtx)
													})
												})
											}),
											layout.Expanded(func(gtx C) D {
												return state.Layout(gtx, func(gtx C) D {
													gtx.Constraints.Min.X = 0
													return component.Menu(th, &p.tagCtxMenu).Layout(gtx)
												})
											}),
										)
									})
								})
							}

							getAddTagBtn := func(gtx C) D {
								return widget.Border{
									Color: color.NRGBA{A: 64},
									Width: unit.Dp(1),
								}.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx C) D {
										return layout.Flex{
											Axis: layout.Horizontal,
										}.Layout(gtx,
											layout.Flexed(1, func(gtx C) D {
												return material.Editor(p.th, &p.tagInput, "new tag").Layout(gtx)
											}),
											layout.Rigid(func(gtx C) D {
												return material.Button(p.th, &p.addTagClick, "+tag").Layout(gtx)
											}))
									})
								})
							}

							getTagSettingsWidget := func(gtx C) D {
								return widget.Border{
									Color: color.NRGBA{A: 64},
									Width: unit.Dp(1),
								}.Layout(gtx, func(gtx C) D {
									return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx C) D {
										return layout.Flex{
											Axis: layout.Vertical,
										}.Layout(gtx,
											layout.Rigid(material.Body2(p.th, "tags").Layout),
											layout.Rigid(getTagList),
											layout.Rigid(getAddTagBtn),
										)
									})
								})
							}
							content = append(content, getTagSettingsWidget)
						} else {
							content = append(content, p.selectedAccPage.Layout)
						}

						content = append(content, layout.Spacer{Height: unit.Dp(25)}.Layout)
						content = append(content, material.Button(p.th, &p.closeSelectedAcc, "close").Layout)

						return listStyle.Layout(gtx, len(content), func(gtx C, i int) D {
							return content[i](gtx)
						})
					})
			}),
		)
	}
}
