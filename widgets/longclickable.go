package widgets

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"time"
)

type LongClickable struct {
	widget.Clickable
	pressed      bool
	shortPressed bool
	longPressed  bool
	triggered    bool
	dur          time.Duration
	timer        *time.Timer
}

func NewLongClickable(dur time.Duration) LongClickable {
	return LongClickable{
		dur:       dur,
		triggered: true,
	}
}

func (b *LongClickable) update() {
	if b.Pressed() {
		if b.pressed == false {
			// first press
			b.triggered = false
			b.shortPressed = false
			b.longPressed = false
			b.timer = time.AfterFunc(b.dur, func() {
				fmt.Println("AfterFunc")
				b.triggered = true
				if b.pressed {
					b.longPressed = true
					b.shortPressed = false
				} else {
					b.shortPressed = true
					b.longPressed = false
				}
			})
		}
		b.pressed = true
	} else {
		b.pressed = false
		if b.timer != nil {
			b.timer.Stop()
		}
		if !b.triggered {
			b.shortPressed = true
			b.longPressed = false
			b.triggered = true
		}
	}
}

func (b *LongClickable) ShortClick() bool {
	res := b.shortPressed
	b.shortPressed = false
	return res
}

func (b *LongClickable) LongClick() bool {
	res := b.longPressed
	b.longPressed = false
	return res
}

func (b *LongClickable) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {

	b.update()

	return material.Clickable(gtx, &b.Clickable, w)
}
