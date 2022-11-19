package application

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	page "hkapp/pages"
)

type App struct {
	Manager *hkontroller.Controller
	Window  *app.Window
	Router  *page.Router

	*material.Theme
}

func NewApp(controller *hkontroller.Controller, window *app.Window, router *page.Router) *App {
	return &App{
		Manager: controller,
		Window:  window,
		Router:  router,
		Theme:   material.NewTheme(gofont.Collection()),
	}
}

func (a *App) Loop() error {
	return a.loop()
}

func (a *App) loop() error {
	w := a.Window
	router := a.Router
	th := a.Theme

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
