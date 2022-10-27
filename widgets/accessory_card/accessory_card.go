package accessory_card

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"image"
	"image/color"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type AccessoryCard struct {
	clickable *widget.Clickable

	acc *hkontroller.Accessory
	th  *material.Theme
}

func NewAccessoryCard(th *material.Theme, clickable *widget.Clickable, accessory *hkontroller.Accessory) *AccessoryCard {
	return &AccessoryCard{
		clickable: clickable,
		acc:       accessory,
		th:        th,
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

	var serviceLabels []layout.FlexChild
	serviceLabels = append(serviceLabels,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Body1(s.th, label).Layout(gtx)
		}),
	)
	for i := range s.acc.Ss {
		srv := s.acc.Ss[i]
		fc := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Body2(s.th, string(srv.Type)).Layout(gtx)
		})
		serviceLabels = append(serviceLabels, fc)
	}

	content := func(gtx C) D {
		return widget.Border{
			Color: color.NRGBA{
				R: 0,
				G: 0,
				B: 0,
				A: 32,
			},
			Width:        unit.Dp(1),
			CornerRadius: unit.Dp(1),
		}.Layout(gtx, func(gtx C) D {
			max := image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.X)
			rect := image.Rectangle{
				Max: max,
			}
			paint.FillShape(gtx.Ops, color.NRGBA{
				R: 0,
				G: 128,
				B: 255,
				A: 64,
			}, clip.Rect(rect).Op())
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				serviceLabels...,
			)
		})
	}

	return material.Clickable(gtx, s.clickable, content)
}
