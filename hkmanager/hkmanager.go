package hkmanager

import (
	"fmt"
	"github.com/brutella/dnssd"
	"github.com/hkontrol/hkontroller"
	"sync"
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

	discoverEvent chan interface{}

	verifiedEvent chan *hkontroller.Device
	closedEvent   chan *hkontroller.Device
}

func NewAppManager(controller *hkontroller.Controller, store hkontroller.Store) *HomeKitManager {
	return &HomeKitManager{
		controller:    controller,
		store:         store,
		discovered:    make(map[string]*hkontroller.Device),
		mu:            sync.Mutex{},
		discoverEvent: make(chan interface{}),
		verifiedEvent: make(chan *hkontroller.Device),
		closedEvent:   make(chan *hkontroller.Device),
	}
}

func (a *HomeKitManager) StartDiscovering() {
	_ = a.controller.LoadPairings()
	a.controller.StartDiscovering(
		func(entry *dnssd.BrowseEntry, device *hkontroller.Device) {
			a.mu.Lock()
			defer a.mu.Unlock()
			id := device.Id
			a.discovered[id] = device

			if device.IsPaired() && !device.IsVerified() {
				err := a.PairVerify(id)
				if err != nil {
					_ = a.UnpairDevice(device)
				}
			}
			a.discoverEvent <- 1
		},
		func(entry *dnssd.BrowseEntry, device *hkontroller.Device) {
			a.mu.Lock()
			defer a.mu.Unlock()
			id := device.Id
			delete(a.discovered, id)
			a.closedEvent <- device
			a.discoverEvent <- 0
		},
	)
}

func (a *HomeKitManager) DiscoverEvents() chan interface{} {
	return a.discoverEvent
}

func (a *HomeKitManager) VerifiedEvents() chan *hkontroller.Device {
	return a.verifiedEvent
}

func (a *HomeKitManager) ClosedEvents() chan *hkontroller.Device {
	return a.closedEvent
}

func (a *HomeKitManager) GetDevices() []*hkontroller.Device {
	var res []*hkontroller.Device

	a.mu.Lock()
	defer a.mu.Unlock()

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

	a.mu.Lock()
	defer a.mu.Unlock()

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
	err := a.controller.PairSetup(devId, pin)
	if err != nil {
		return err
	}
	return a.PairVerify(devId)
}

func (a *HomeKitManager) PairVerify(devId string) error {
	err := a.controller.PairVerify(devId)
	if err != nil {
		return err
	}

	go func() {
		dev := a.controller.GetDevice(devId)
		if dev != nil {
			err := dev.DiscoverAccessories()
			fmt.Println("discover accs: ", err)
			a.verifiedEvent <- a.controller.GetDevice(devId)
		}
	}()
	return nil
}

func (a *HomeKitManager) UnpairDevice(dev *hkontroller.Device) error {
	defer func() {
		a.closedEvent <- dev
	}()
	err := a.controller.UnpairDevice(dev.Id)
	return err
}
