package main

import (
	"flag"
	"fmt"
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"hkapp/appmanager"
	page "hkapp/pages"
	"hkapp/pages/accessories"
	"hkapp/pages/discover"
	"os"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type App struct {
	manager    *appmanager.AppManager
	controller *hkontroller.Controller
	window     *app.Window
	router     *page.Router
}

func main() {
	flag.Parse()

	st := hkontroller.NewFsStore("./.store")

	hk, _ := hkontroller.NewController(
		st,
		"hkontroller",
	)
	_ = hk.LoadPairings()

	mgr := appmanager.NewAppManager(hk, st)

	w := app.NewWindow(
		app.Title("hkontroller"),
		app.Size(unit.Dp(400), unit.Dp(600)),
	)

	myapp := App{
		manager:    mgr,
		controller: hk,
		window:     w,
	}

	router := page.NewRouter()
	discoverPage := discover.New(router, mgr)
	accessoriesPage := accessories.New(router, mgr)
	router.Register(0, discoverPage)
	router.Register(1, accessoriesPage)
	myapp.router = router

	go func() {
		if err := myapp.loop(); err != nil {
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

func (a *App) loop() error {
	return loop(a.window, a.router, a.manager)
}

func loop(w *app.Window, router *page.Router, mgr *appmanager.AppManager) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				router.Layout(gtx, th)
				e.Frame(gtx.Ops)
			}
		}
	}
}
