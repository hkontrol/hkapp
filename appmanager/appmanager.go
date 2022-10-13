package appmanager

import (
	"fmt"
	"github.com/brutella/dnssd"
	"github.com/hkontrol/hkontroller"
	"sync"
)

type AppManager struct {
	controller *hkontroller.Controller
	store      hkontroller.Store

	discovered map[string]*hkontroller.Device
	mu         sync.Mutex

	discoverEvent chan interface{}

	verifiedEvent chan *hkontroller.Device
	closedEvent   chan *hkontroller.Device
}

func NewAppManager(controller *hkontroller.Controller, store hkontroller.Store) *AppManager {
	return &AppManager{
		controller:    controller,
		store:         store,
		discovered:    make(map[string]*hkontroller.Device),
		mu:            sync.Mutex{},
		discoverEvent: make(chan interface{}),
		verifiedEvent: make(chan *hkontroller.Device),
		closedEvent:   make(chan *hkontroller.Device),
	}
}

func (a *AppManager) StartDiscovering() {
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

func (a *AppManager) DiscoverEvents() chan interface{} {
	return a.discoverEvent
}

func (a *AppManager) VerifiedEvents() chan *hkontroller.Device {
	return a.verifiedEvent
}

func (a *AppManager) ClosedEvents() chan *hkontroller.Device {
	return a.closedEvent
}

func (a *AppManager) GetDevices() []*hkontroller.Device {
	var res []*hkontroller.Device

	a.mu.Lock()
	defer a.mu.Unlock()

	for _, d := range a.discovered {
		dev := a.controller.GetPairedDevice(d.Id)
		if dev != nil {
			res = append(res, dev)
		}
	}

	return res
}

func (a *AppManager) GetVerifiedDevices() []*hkontroller.Device {
	var res []*hkontroller.Device

	a.mu.Lock()
	defer a.mu.Unlock()

	for _, d := range a.discovered {
		dev := a.controller.GetPairedDevice(d.Id)
		if dev != nil {
			if dev.IsPaired() && dev.IsVerified() {
				res = append(res, dev)
			}
		}
	}

	return res
}

func (a *AppManager) PairSetupAndVerify(devId string, pin string) error {
	err := a.controller.PairSetup(devId, pin)
	if err != nil {
		return err
	}
	return a.PairVerify(devId)
}
func (a *AppManager) PairVerify(devId string) error {
	err := a.controller.PairVerify(devId)
	if err != nil {
		return err
	}

	go func() {
		dev := a.controller.GetPairedDevice(devId)
		if dev != nil {
			err := dev.DiscoverAccessories()
			fmt.Println("discover accs: ", err)
			a.verifiedEvent <- a.controller.GetPairedDevice(devId)
		}
	}()
	return nil
}

func (a *AppManager) UnpairDevice(dev *hkontroller.Device) error {
	defer func() {
		a.closedEvent <- dev
	}()
	err := a.controller.UnpairDevice(dev)
	return err
}
