package main

import (
	"context"
	"flag"
	"fmt"
	"hkapp/application"
	page "hkapp/pages"
	"hkapp/pages/accessories"
	"hkapp/pages/discover"
	"log"
	"os"
	"path"
	"sync"
	"time"

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

	updatePages()

	mu := sync.Mutex{}
	devs := make(map[string]struct{}) // to store already discovered accs

	discoCh, lostCh := hk.StartDiscovering()

	go func() {
		for dev := range discoCh {
			fmt.Println("discovered: ", dev.Id)

			mu.Lock()
			_, ok := devs[dev.Id]
			mu.Unlock()

			if !ok { // first discover
				fmt.Println("first discover")
				mu.Lock()
				devs[dev.Id] = struct{}{}
				mu.Unlock()

				go func(d *hkontroller.Device) {
					for range dev.OnVerified() {
						err := dev.GetAccessories()
						if err == nil {
							updatePages()
							w.Invalidate()
						}
					}
				}(dev)
				go func(d *hkontroller.Device) {
					for range dev.OnClose() {
						updatePages()
						w.Invalidate()
					}
				}(dev)
				go func(d *hkontroller.Device) {
					for range dev.OnLost() {
						updatePages()
						w.Invalidate()
					}
				}(dev)
				go func(d *hkontroller.Device) {
					for range dev.OnUnpaired() {
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
					ctx := context.Background()
					err := d.PairSetupAndVerify(ctx, "---", 5*time.Second)
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
				delete(devs, dev.Id)
				mu.Unlock()
			}
			updatePages()
		}
	}()

	app.Main()
}
