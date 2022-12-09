package service_cards

import (
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
	"hkapp/application"
)

/*
Required Characteristics
9.32 Current Heating Cooling State : 0-Off 1-Heat 2-Cool
9.119 Target Heating Cooling State : 0-Off 1-Heat 2-Cool 3-Auto
9.35 Current Temperature
9.121 Target Temperature : 10.0-38.0
9.122 Temperature Display Units : 0-Celsius 1-Fahrenheit

Optional Characteristics
9.20 Cooling Threshold Temperature : 10-35
9.34 Current Relative Humidity : 0-100
9.42 Heating Threshold Temperature : 0-25
9.62 Name (page 188)
9.120 Target Relative Humidity : 0-100
*/

type Thermostat struct {
	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device

	events <-chan emitter.Event

	th *material.Theme

	*application.App
}

func NewThermostat() (*Thermostat, error) {
	return &Thermostat{}, nil
}

func (t *Thermostat) SubscribeToEvents() error {
	return nil
}

func (t *Thermostat) UnsubscribeFromEvents() error {
	return nil
}

func (t *Thermostat) Layout(gtx C) D {
	return D{}
}
