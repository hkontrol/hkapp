package service_cards

import (
	"errors"
	"fmt"
	"gioui.org/widget"
	"hkapp/application"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
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

	currentHeatCoolStateC *hkontroller.CharacteristicDescription
	targetHeatCoolStateC  *hkontroller.CharacteristicDescription
	currentTempC          *hkontroller.CharacteristicDescription
	targetTempC           *hkontroller.CharacteristicDescription
	tempDisplayUnitsC     *hkontroller.CharacteristicDescription

	coolingThresholdTempC   *hkontroller.CharacteristicDescription
	heatingThresholdTempC   *hkontroller.CharacteristicDescription
	currentRelativeHumidity *hkontroller.CharacteristicDescription
	targetRelativeHumidity  *hkontroller.CharacteristicDescription

	targetModeEnum  widget.Enum
	targetTempFloat widget.Float

	th *material.Theme

	*application.App
}

func NewThermostat(app *application.App, acc *hkontroller.Accessory, dev *hkontroller.Device) (*Thermostat, error) {
	t := &Thermostat{
		App: app,
		acc: acc,
		dev: dev,
		th:  app.Theme,
	}

	infoS := acc.GetService(hkontroller.SType_AccessoryInfo)
	if infoS == nil {
		return nil, errors.New("cannot get AccessoryInfo service")
	}
	labelC := infoS.GetCharacteristic(hkontroller.CType_Name)
	if labelC == nil {
		return nil, errors.New("cannot get characteristic Name")
	}
	label, ok := labelC.Value.(string)
	if !ok {
		return nil, errors.New("cannot extract accessory name")
	}
	t.label = label

	srv := acc.GetService(hkontroller.SType_Thermostat)
	if srv == nil {
		return nil, errors.New("cannot find thermostat service")
	}

	// required chars
	currentHeatCoolStateC := srv.GetCharacteristic(hkontroller.CType_CurrentHeatingCoolingState)
	if currentHeatCoolStateC == nil {
		return nil, errors.New("cannot find required characteristic")
	}
	targetHeatCoolStateC := srv.GetCharacteristic(hkontroller.CType_TargetHeatingCoolingState)
	if targetHeatCoolStateC == nil {
		return nil, errors.New("cannot find required characteristic")
	}
	currentTempC := srv.GetCharacteristic(hkontroller.CType_CurrentTemperature)
	if currentTempC == nil {
		return nil, errors.New("cannot find required characteristic")
	}
	targetTempC := srv.GetCharacteristic(hkontroller.CType_TargetTemperature)
	if targetTempC == nil {
		return nil, errors.New("cannot find required characteristic")
	}
	tempDisplayUnitsC := srv.GetCharacteristic(hkontroller.CType_TemperatureDisplayUnits)
	if tempDisplayUnitsC == nil {
		return nil, errors.New("cannot find required characteristic")
	}

	t.currentHeatCoolStateC = currentHeatCoolStateC
	t.targetHeatCoolStateC = targetHeatCoolStateC
	t.currentTempC = currentTempC
	t.targetTempC = targetTempC
	t.tempDisplayUnitsC = tempDisplayUnitsC

	// TODO hapEvents

	// optional chars
	t.coolingThresholdTempC = srv.GetCharacteristic(hkontroller.CType_CoolingThresholdTemperature)
	t.heatingThresholdTempC = srv.GetCharacteristic(hkontroller.CType_HeatingThresholdTemperature)
	t.currentRelativeHumidity = srv.GetCharacteristic(hkontroller.CType_CurrentRelativeHumidity)
	t.targetRelativeHumidity = srv.GetCharacteristic(hkontroller.CType_TargetRelativeHumidity)

	return t, nil
}

func (t *Thermostat) SubscribeToEvents() error {
	return nil
}

func (t *Thermostat) UnsubscribeFromEvents() error {
	return nil
}

func (t *Thermostat) Layout(gtx C) D {

	for t.targetTempFloat.Changed() {
		val := t.targetTempFloat.Value
		err := t.dev.PutCharacteristic(t.acc.Id, t.targetTempC.Iid, val)
		if err != nil {
			return D{}
		}
		t.App.EmitValueChange(t.dev.Id, t.acc.Id, t.targetTempC.Iid, val)
	}
	for t.targetModeEnum.Changed() {
		valStr := t.targetModeEnum.Value
		valInt := 0
		switch valStr {
		case "off":
			valInt = 0
		case "heat":
			valInt = 1
		case "cool":
			valInt = 2
		case "auto":
			valInt = 3
		}
		err := t.dev.PutCharacteristic(t.acc.Id, t.targetHeatCoolStateC.Iid, valInt)
		if err != nil {
			return D{}
		}
		t.App.EmitValueChange(t.dev.Id, t.acc.Id, t.targetHeatCoolStateC.Iid, valInt)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(material.Body1(t.th, fmt.Sprintf("Current t: %v | ", t.currentTempC.Value)).Layout),
				layout.Rigid(material.Body1(t.th, fmt.Sprintf("Target t: %v", t.targetTempC.Value)).Layout),
			)
		}),
		layout.Rigid(material.Slider(t.th, &t.targetTempFloat, 10, 38).Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(material.RadioButton(t.th, &t.targetModeEnum, "off", "Off").Layout),
				layout.Rigid(material.RadioButton(t.th, &t.targetModeEnum, "heat", "Heat").Layout),
				layout.Rigid(material.RadioButton(t.th, &t.targetModeEnum, "cool", "Cool").Layout),
				layout.Rigid(material.RadioButton(t.th, &t.targetModeEnum, "auto", "Auto").Layout),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(material.Body1(t.th, fmt.Sprintf("Current mode: %v | ", t.currentHeatCoolStateC.Value)).Layout),
				layout.Rigid(material.Body1(t.th, fmt.Sprintf("Target mode: %v", t.targetHeatCoolStateC.Value)).Layout),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx)
		}),
	)
}