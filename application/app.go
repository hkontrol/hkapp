package application

import (
	"fmt"
	page "hkapp/pages"
	"path"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/hkontrol/hkontroller"
	"github.com/olebedev/emitter"
)

type App struct {
	Manager *hkontroller.Controller
	Window  *app.Window
	Router  *page.Router

	*AccessoryMetadataStore

	*material.Theme

	// to be able to emit value changes
	// to share state between widgets
	ee emitter.Emitter
}

func NewApp(controller *hkontroller.Controller, window *app.Window, router *page.Router, settingsDir string) *App {
	return &App{
		Manager:                controller,
		Window:                 window,
		Router:                 router,
		Theme:                  material.NewTheme(gofont.Collection()),
		AccessoryMetadataStore: NewAccessoryMetadataStore(path.Join(settingsDir, "accmeta")),

		ee: emitter.Emitter{},
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

func (a *App) OnValueChange(deviceId string, aid uint64, iid uint64) <-chan emitter.Event {
	topic := fmt.Sprintf("value_%s_%d_%d", deviceId, aid, iid)
	return a.ee.On(topic)
}

func (a *App) OffValueChange(deviceId string, aid uint64, iid uint64, ch <-chan emitter.Event) {
	topic := fmt.Sprintf("value_%s_%d_%d", deviceId, aid, iid)
	a.ee.Off(topic, ch)
}

func (a *App) EmitValueChange(deviceId string, aid uint64, iid uint64, value interface{}) <-chan struct{} {
	topic := fmt.Sprintf("value_%s_%d_%d", deviceId, aid, iid)
	return a.ee.Emit(topic, aid, iid, value)
}
