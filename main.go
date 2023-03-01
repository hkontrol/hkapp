package main

import (
	"flag"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/hkontrol/hkontroller"
	"hkapp/application"
	page "hkapp/pages"
	"hkapp/pages/accessories"
	"hkapp/pages/discover"
	"log"
	"os"
	"path"
	"sync"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func main() {
	flag.Parse()

	dd, err := app.DataDir()
	if err != nil {
		panic(err)
	}

	storePath := path.Join(dd, "hkapp", "controller")

	fmt.Println("store path: ", storePath)

	st := hkontroller.NewFsStore(storePath)

	hk, _ := hkontroller.NewController(
		st,
		"hkontroller",
	)
	_ = hk.LoadPairings()

	w := app.NewWindow(
		app.Title("hkontroller"),
		app.Size(unit.Dp(400), unit.Dp(600)),
	)

	router := page.NewRouter()

	myapp := application.NewApp(hk, w, router, dd)
	myapp.Window.Invalidate()

	discoverPage := discover.New(myapp)
	accessoriesPage := accessories.New(myapp)
	router.Register(0, accessoriesPage)
	router.Register(1, discoverPage)

	go func() {
		if err := myapp.Loop(); err != nil {
			panic(err)
		}
		os.Exit(0)
	}()

	updatePages := func() {
		discoverPage.Update()
		accessoriesPage.Update()
		w.Invalidate()
	}

	updatePages()

	mu := sync.Mutex{}
	devs := make(map[string]struct{}) // to store already discovered accs

	discoCh, lostCh := hk.StartDiscovery()

	go func() {
		for dev := range discoCh {
			fmt.Println("discovered: ", dev.Name)

			mu.Lock()
			_, ok := devs[dev.Name]
			mu.Unlock()

			if !ok { // first discover
				fmt.Println("first discover")
				mu.Lock()
				devs[dev.Name] = struct{}{}
				mu.Unlock()

				go func(d *hkontroller.Device) {
					for range dev.OnVerified() {
						fmt.Println("dev onverified ", dev.Name)
						err := dev.GetAccessories()
						if err == nil {
							updatePages()
							w.Invalidate()
						}
					}
					fmt.Println("dev. onverified channel gone?")
				}(dev)
				go func(d *hkontroller.Device) {
					for range dev.OnClose() {
						fmt.Println("dev onclose ", dev.Name)
						updatePages()
						w.Invalidate()
					}
				}(dev)
				go func(d *hkontroller.Device) {
					for range dev.OnLost() {
						fmt.Println("dev onlost ", dev.Name)
						updatePages()
						w.Invalidate()
					}
				}(dev)
				go func(d *hkontroller.Device) {
					for range dev.OnUnpaired() {
						fmt.Println("dev onunpaired ", dev.Name)
						updatePages()
						w.Invalidate()
					}
				}(dev)
			} else {
				fmt.Println("was discovered before")
			}

			if dev.IsPaired() {
				log.Println("already paired, establishing connection")
				go func(d *hkontroller.Device) {
					err := d.PairVerify()
					if err != nil {
						log.Println("pair-verify err: ", err)
						return
					}
				}(dev)
			}

			updatePages()
		}
	}()

	go func() {
		for dev := range lostCh {
			if !dev.IsPaired() {
				log.Println("lost and not paired, delete from discovered")
				mu.Lock()
				delete(devs, dev.Name)
				mu.Unlock()
			}
			updatePages()
		}
	}()

	app.Main()
}
