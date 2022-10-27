package service_cards

import (
	"fmt"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"hkapp/icon"
)

type LightBulb struct {
	click widget.Clickable

	aid   uint64
	label string
	th    *material.Theme
}

func NewLightBulb(aid uint64, label string, th *material.Theme) *LightBulb {
	return &LightBulb{
		aid:   aid,
		label: label,
		th:    th,
	}
}

func (l *LightBulb) Layout(gtx C) D {
	fmt.Println("lightBulb.Layout")
	// here we loop through all the events associated with this button.

	for range l.click.Clicks() {
		fmt.Println("click ", l.label)
	}

	material.Body1(l.th, l.label).Layout(gtx)
	return material.IconButton(l.th, &l.click, icon.LightBulbIcon, l.label).Layout(gtx)
}
