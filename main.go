package main

import (
	"flag"
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/hkontrol/hkontroller"
	"hkapp/application"
	"hkapp/hkmanager"
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

	mgr := hkmanager.NewAppManager(hk, st)

	w := app.NewWindow(
		app.Title("hkontroller"),
		app.Size(unit.Dp(400), unit.Dp(600)),
	)

	myapp := &application.App{
		Manager: mgr,
		//controller: hk,
		Window: w,
	}
	myapp.Window.Invalidate()

	router := page.NewRouter()
	discoverPage := discover.New(myapp)
	accessoriesPage := accessories.New(myapp)
	router.Register(0, accessoriesPage)
	router.Register(1, discoverPage)
	myapp.Router = router

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
	go func() {
		for range mgr.EventDeviceDiscover() {
			updatePages()
		}
	}()

	go func() {
		for range mgr.EventDeviceLost() {
			updatePages()
		}
	}()

	go func() {
		for range mgr.EventDeviceVerified() {
			updatePages()
		}
	}()
	go func() {
		for range mgr.EventDeviceClose() {
			updatePages()
		}
	}()

	go func() {
		mgr.StartDiscovering()
	}()

	app.Main()
}
