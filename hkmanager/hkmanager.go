package hkmanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/hkontrol/hkontroller"
	"sync"
	"time"
)

type DeviceAccPair struct {
	Device    *hkontroller.Device
	Accessory *hkontroller.Accessory
}

type HomeKitManager struct {
	controller *hkontroller.Controller
	store      hkontroller.Store

	discovered map[string]*hkontroller.Device
	mu         sync.Mutex

	discoverEvent chan *hkontroller.Device
	lostEvent     chan *hkontroller.Device
	pairedEvent   chan *hkontroller.Device
	verifiedEvent chan *hkontroller.Device
	closeEvent    chan *hkontroller.Device

	//emitter.Emitter
}

func NewAppManager(controller *hkontroller.Controller, store hkontroller.Store) *HomeKitManager {
	return &HomeKitManager{
		controller: controller,
		store:      store,
		discovered: make(map[string]*hkontroller.Device),
		mu:         sync.Mutex{},

		discoverEvent: make(chan *hkontroller.Device),
		lostEvent:     make(chan *hkontroller.Device),
		pairedEvent:   make(chan *hkontroller.Device),
		verifiedEvent: make(chan *hkontroller.Device),
		closeEvent:    make(chan *hkontroller.Device),
	}
}

func (a *HomeKitManager) EventDeviceDiscover() chan *hkontroller.Device {
	return a.discoverEvent
}
func (a *HomeKitManager) EventDeviceLost() chan *hkontroller.Device {
	return a.lostEvent
}
func (a *HomeKitManager) EventDevicePaired() chan *hkontroller.Device {
	return a.pairedEvent
}
func (a *HomeKitManager) EventDeviceVerified() chan *hkontroller.Device {
	return a.verifiedEvent
}
func (a *HomeKitManager) EventDeviceClose() chan *hkontroller.Device {
	return a.closeEvent
}

func (a *HomeKitManager) StartDiscovering() {
	_ = a.controller.LoadPairings()
	discoCh, lostCh := a.controller.StartDiscovering()

	go func() {
		for device := range discoCh {
			a.mu.Lock()
			defer a.mu.Unlock()

			id := device.Id
			a.discovered[id] = device

			fmt.Println("on discovered: ", id)

			a.discoverEvent <- device

			if device.IsPaired() && !device.IsVerified() {
				err := a.PairVerify(id)
				if err != nil {
					_ = a.UnpairDevice(device)
				}
			}
		}
	}()

	go func() {
		for device := range lostCh {
			a.mu.Lock()
			defer a.mu.Unlock()
			id := device.Id
			fmt.Println("lost discovered: ", id)

			delete(a.discovered, id)
			a.lostEvent <- device
		}
	}()
}

func (a *HomeKitManager) GetDevices() []*hkontroller.Device {
	var res []*hkontroller.Device

	for _, d := range a.discovered {
		dev := a.controller.GetDevice(d.Id)
		if dev != nil {
			res = append(res, dev)
		}
	}

	return res
}

func (a *HomeKitManager) GetVerifiedDevices() []*hkontroller.Device {
	var res []*hkontroller.Device

	for _, d := range a.discovered {
		dev := a.controller.GetDevice(d.Id)
		if dev != nil {
			if dev.IsPaired() && dev.IsVerified() {
				res = append(res, dev)
			}
		}
	}

	return res
}

func (a *HomeKitManager) PairSetupAndVerify(devId string, pin string) error {
	d := a.controller.GetDevice(devId)
	if d == nil {
		return errors.New("no device found")
	}
	err := d.PairSetup(pin)
	if err != nil {
		return err
	}
	return a.PairVerify(devId)
}

func (a *HomeKitManager) PairVerify(devId string) error {
	d := a.controller.GetDevice(devId)
	if d == nil {
		return errors.New("no device found")
	}
	err := d.PairSetupAndVerify(context.Background(), "xz", 5*time.Second)

	err = d.GetAccessories()
	fmt.Println("discovered accs: ", len(d.Accessories()), err)
	a.verifiedEvent <- d // TODO replace
	return nil
}

func (a *HomeKitManager) UnpairDevice(dev *hkontroller.Device) error {
	err := dev.Unpair()
	a.lostEvent <- dev // TODO replace
	return err
}
