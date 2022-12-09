package service_cards

import (
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
	"hkapp/application"
)

type Dummy struct {
	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device

	events <-chan emitter.Event

	th *material.Theme

	*application.App
}

func NewDummy() (*Dummy, error) {
	return &Dummy{}, nil
}

func (d *Dummy) SubscribeToEvents() error {
	return nil
}

func (d *Dummy) UnsubscribeFromEvents() error {
	return nil
}

func (d *Dummy) Layout(gtx C) D {
	return D{}
}
