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
	"os"
	"path"
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

	storePath := path.Join(dd, "hkstore")

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

	myapp := application.NewApp(hk, w, router)
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

	discoCh, lostCh := hk.StartDiscovering()

	go func() {
		for dev := range discoCh {
			go func(d *hkontroller.Device) {
				for range dev.OnVerified() {
					err := dev.GetAccessories()
					if err == nil {
						accessoriesPage.Update()
						w.Invalidate()
					}
				}
			}(dev)
			go func(d *hkontroller.Device) {
				for range dev.OnClose() {
					accessoriesPage.Update()
					w.Invalidate()
				}
			}(dev)
			go func(d *hkontroller.Device) {
				for range dev.OnLost() {
					accessoriesPage.Update()
					discoverPage.Update()
					w.Invalidate()
				}
			}(dev)
			go func(d *hkontroller.Device) {
				for range dev.OnUnpaired() {
					accessoriesPage.Update()
					discoverPage.Update()
					w.Invalidate()
				}
			}(dev)

			updatePages()
		}
	}()

	go func() {
		for range lostCh {
			updatePages()
		}
	}()

	app.Main()
}
