package service_cards

import (
	"fmt"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"hkapp/applayout"
)

type Switch struct {
	widget.Bool

	aid   uint64
	label string
	th    *material.Theme
}

func NewSwitch(aid uint64, label string, th *material.Theme) *Switch {
	return &Switch{
		aid:   aid,
		label: label,
		th:    th,
	}
}

func (s *Switch) Layout(gtx C) D {
	fmt.Println("switch.Layout ", s.label)

	if s.Bool.Changed() {
		fmt.Println("changed ", s.Bool.Value)
	}

	return applayout.DetailRow{}.Layout(gtx,
		material.Body1(s.th, s.label).Layout,
		material.Switch(s.th, &s.Bool, s.label).Layout,
	)
}
