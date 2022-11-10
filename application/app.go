package application

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"hkapp/hkmanager"
	page "hkapp/pages"
)

type App struct {
	Manager *hkmanager.HomeKitManager
	Window  *app.Window
	Router  *page.Router
}

func (a *App) Loop() error {
	return loop(a.Window, a.Router, a.Manager)
}

func loop(w *app.Window, router *page.Router, mgr *hkmanager.HomeKitManager) error {
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
