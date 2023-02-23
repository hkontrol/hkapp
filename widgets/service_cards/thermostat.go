package service_cards

import (
	"errors"
	"fmt"
	"hkapp/application"
	"math"
	"time"

	"gioui.org/widget"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
)

const targetTempDragDelay = 300 * time.Millisecond

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
	quick bool // simplified version to display in list of accs

	label string

	acc *hkontroller.Accessory
	dev *hkontroller.Device

	hapEvents map[hkontroller.HapCharacteristicType]<-chan emitter.Event
	guiEvents map[hkontroller.HapCharacteristicType]<-chan emitter.Event

	/*
		currentHeatCoolStateC *hkontroller.CharacteristicDescription
		targetHeatCoolStateC  *hkontroller.CharacteristicDescription
		currentTempC          *hkontroller.CharacteristicDescription
		targetTempC           *hkontroller.CharacteristicDescription
		tempDisplayUnitsC     *hkontroller.CharacteristicDescription

		coolingThresholdTempC   *hkontroller.CharacteristicDescription
		heatingThresholdTempC   *hkontroller.CharacteristicDescription
		currentRelativeHumidity *hkontroller.CharacteristicDescription
		targetRelativeHumidity  *hkontroller.CharacteristicDescription
	*/

	chars map[hkontroller.HapCharacteristicType]*hkontroller.CharacteristicDescription

	targetModeEnum        widget.Enum
	targetTempFloatWidget widget.Float
	targetTempFloatValue  float32

	dragTimer *time.Timer

	th *material.Theme

	*application.App
}

func NewThermostat(app *application.App,
	acc *hkontroller.Accessory,
	dev *hkontroller.Device,
	quickWidget bool) (*Thermostat, error) {
	t := &Thermostat{
		quick:     quickWidget,
		App:       app,
		acc:       acc,
		dev:       dev,
		th:        app.Theme,
		chars:     make(map[hkontroller.HapCharacteristicType]*hkontroller.CharacteristicDescription),
		hapEvents: make(map[hkontroller.HapCharacteristicType]<-chan emitter.Event),
		guiEvents: make(map[hkontroller.HapCharacteristicType]<-chan emitter.Event),
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

	t.chars[hkontroller.CType_CurrentHeatingCoolingState] = currentHeatCoolStateC
	t.chars[hkontroller.CType_TargetHeatingCoolingState] = targetHeatCoolStateC
	t.chars[hkontroller.CType_CurrentTemperature] = currentTempC
	t.chars[hkontroller.CType_TargetTemperature] = targetTempC
	t.chars[hkontroller.CType_TemperatureDisplayUnits] = tempDisplayUnitsC

	// optional chars
	coolingThresholdTempC := srv.GetCharacteristic(hkontroller.CType_CoolingThresholdTemperature)
	heatingThresholdTempC := srv.GetCharacteristic(hkontroller.CType_HeatingThresholdTemperature)
	currentRelativeHumidity := srv.GetCharacteristic(hkontroller.CType_CurrentRelativeHumidity)
	targetRelativeHumidity := srv.GetCharacteristic(hkontroller.CType_TargetRelativeHumidity)
	if coolingThresholdTempC != nil {
		t.chars[hkontroller.CType_CoolingThresholdTemperature] = coolingThresholdTempC
	}
	if heatingThresholdTempC != nil {
		t.chars[hkontroller.CType_HeatingThresholdTemperature] = heatingThresholdTempC
	}
	if currentRelativeHumidity != nil {
		t.chars[hkontroller.CType_CurrentRelativeHumidity] = currentRelativeHumidity
	}
	if targetRelativeHumidity != nil {
		t.chars[hkontroller.CType_TargetRelativeHumidity] = targetRelativeHumidity
	}

	for ctype, cdescr := range t.chars {
		// TODO: for such a case one should collect multiple chars and Get them
		//        should be implemented in core hkontroller package as well
		//         think also about saving in hkontroller.Device instance
		descr, err := t.dev.GetCharacteristic(t.acc.Id, cdescr.Iid)
		if err != nil {
			fmt.Println("err: ", err)
			continue
		}
		t.chars[ctype] = &descr
		t.onValue(descr.Value, ctype)
	}
	t.App.Window.Invalidate()

	return t, nil
}

func (t *Thermostat) onValue(value interface{},
	ctype hkontroller.HapCharacteristicType) {

	t.chars[ctype].Value = value
	if ctype == hkontroller.CType_TargetTemperature {
		if val32, ok := value.(float32); ok {
			t.targetTempFloatWidget.Value = val32
			t.targetTempFloatValue = val32
		}
		if val64, ok := value.(float64); ok {
			t.targetTempFloatWidget.Value = float32(val64)
			t.targetTempFloatValue = float32(val64)
		}
	}
	if ctype == hkontroller.CType_TargetHeatingCoolingState {
		var valNum float64
		if vf, ok := value.(float64); ok {
			valNum = vf
		} else if vi, ok := value.(int); ok {
			valNum = float64(vi)
		} else {
			return
		}

		t.targetModeEnum.Value = targetMode2enum(valNum)
	}
}

func (t *Thermostat) SubscribeToEvents() {

	var err error
	var ev <-chan emitter.Event
	devId := t.dev.Name
	aid := t.acc.Id
	onEvent := func(e emitter.Event, ctype hkontroller.HapCharacteristicType) {
		value := e.Args[2]
		t.onValue(value, ctype)
		t.App.Window.Invalidate()
	}
	for ctype, cdescr := range t.chars {
		iid := cdescr.Iid
		ev, err = t.dev.SubscribeToEvents(aid, iid)
		if err != nil {
			fmt.Println("err subscribing: ", cdescr.Type.String(), err)
			continue
		}
		go func(evs <-chan emitter.Event, ct hkontroller.HapCharacteristicType) {
			for e := range evs {
				onEvent(e, ct)
			}
		}(ev, ctype)
		t.hapEvents[ctype] = ev

		ev = t.App.OnValueChange(devId, aid, iid)
		t.guiEvents[ctype] = ev
		go func(evs <-chan emitter.Event, ct hkontroller.HapCharacteristicType) {
			for e := range evs {
				onEvent(e, ct)
			}
		}(ev, ctype)
	}

	return
}

func (t *Thermostat) UnsubscribeFromEvents() {
	aid := t.acc.Id
	devId := t.dev.Name
	for ctype, ee := range t.hapEvents {
		iid := t.chars[ctype].Iid
		err := t.dev.UnsubscribeFromEvents(aid, iid, ee)
		if err != nil {
			continue
		}
		delete(t.hapEvents, ctype)
	}
	for ctype, ee := range t.guiEvents {
		iid := t.chars[ctype].Iid
		t.App.OffValueChange(devId, aid, iid, ee)
		delete(t.hapEvents, ctype)
	}
	return
}

func (t *Thermostat) Layout(gtx C) D {

	for t.targetTempFloatWidget.Changed() {
		if t.targetTempFloatValue == 0 { // should trigger when widget is created
			t.targetTempFloatValue = t.targetTempFloatWidget.Value
			continue
		}

		if t.dragTimer != nil {
			t.dragTimer.Stop()
		}

		// timer to prevent change on drag
		t.dragTimer = time.AfterFunc(targetTempDragDelay, func() {
			val := float64(t.targetTempFloatWidget.Value)
			// one digit after point
			val = math.Floor(val*10) / 10
			ctype := hkontroller.CType_TargetTemperature
			go func() {
				err := t.dev.PutCharacteristic(t.acc.Id, t.chars[ctype].Iid, float32(val))
				if err != nil {
					return
				}
				t.App.EmitValueChange(t.dev.Name, t.acc.Id, t.chars[ctype].Iid, float32(val))
			}()
		})
	}
	for t.targetModeEnum.Changed() {
		valStr := t.targetModeEnum.Value
		valNum := targetMode2num(valStr)
		ctype := hkontroller.CType_TargetHeatingCoolingState
		go func() {
			err := t.dev.PutCharacteristic(t.acc.Id, t.chars[ctype].Iid, valNum)
			if err != nil {
				return
			}
			t.App.EmitValueChange(t.dev.Name, t.acc.Id, t.chars[ctype].Iid, valNum)
		}()
	}

	ctemp := t.chars[hkontroller.CType_CurrentTemperature].Value
	ttemp := t.chars[hkontroller.CType_TargetTemperature].Value
	cmode := t.chars[hkontroller.CType_CurrentHeatingCoolingState].Value
	tmode := t.chars[hkontroller.CType_TargetHeatingCoolingState].Value

	cmodeStr := currentMode2str(cmode)
	tmodeStr := targetMode2enum(tmode)

	// TODO optional chars

	if t.quick {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Temp: %v | ", ctemp)).Layout),
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Target: %v", ttemp)).Layout),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Mode: %v | ", cmodeStr)).Layout),
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Target: %v", tmodeStr)).Layout),
				)
			}),
		)
	} else {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Temp: %v | ", ctemp)).Layout),
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Target: %v", ttemp)).Layout),
				)
			}),
			layout.Rigid(material.Slider(t.th, &t.targetTempFloatWidget, 10, 38).Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(material.RadioButton(t.th,
						&t.targetModeEnum, "off", "Off").Layout),
					layout.Rigid(material.RadioButton(t.th,
						&t.targetModeEnum, "heat", "Heat").Layout),
					layout.Rigid(material.RadioButton(t.th,
						&t.targetModeEnum, "cool", "Cool").Layout),
					layout.Rigid(material.RadioButton(t.th,
						&t.targetModeEnum, "auto", "Auto").Layout),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Mode: %v | ", cmodeStr)).Layout),
					layout.Rigid(material.Body1(t.th,
						fmt.Sprintf("Target: %v", tmodeStr)).Layout),
				)
			}),
		)
	}
}

// util
// ------------------
func currentMode2str(v interface{}) string {
	valStr := "unknown"
	if m, ok := v.(float64); ok {
		switch m {
		case 0:
			valStr = "off"
		case 1:
			valStr = "heat"
		case 2:
			valStr = "cool"
		}
	}
	return valStr
}
func targetMode2enum(v interface{}) string {
	valStr := ""
	if m, ok := v.(float64); ok {
		switch m {
		case 0:
			valStr = "off"
		case 1:
			valStr = "heat"
		case 2:
			valStr = "cool"
		case 3:
			valStr = "auto"
		}
	}
	return valStr
}
func targetMode2num(v interface{}) float64 {
	valNum := float64(0)
	if valStr, ok := v.(string); ok {
		switch valStr {
		case "off":
			valNum = 0
		case "heat":
			valNum = 1
		case "cool":
			valNum = 2
		case "auto":
			valNum = 3
		}
	}
	return valNum
}
