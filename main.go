package main

import (
	"flag"
	"fmt"
	"hkapp/application"
	"hkapp/hkmanager"
	page "hkapp/pages"
	"hkapp/pages/accessories"
	"hkapp/pages/discover"
	"os"
	"path"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/hkontrol/hkontroller"
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

	go func() {
		for _ = range mgr.DiscoverEvents() {
			fmt.Println("discover event")
			discoverPage.Update()
			w.Invalidate()
		}
	}()

	go func() {
		for _ = range mgr.VerifiedEvents() {
			fmt.Println("verified event")
			accessoriesPage.Update()
			w.Invalidate()
		}
	}()
	go func() {
		for _ = range mgr.ClosedEvents() {
			fmt.Println("closed event")
			accessoriesPage.Update()
			w.Invalidate()
		}
	}()
	go func() {
		mgr.StartDiscovering()
	}()

	app.Main()
}
