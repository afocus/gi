// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package svg

import (
	"github.com/goki/gi/gi"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// Circle is a SVG circle
type Circle struct {
	NodeBase
	Pos    gi.Vec2D `xml:"{cx,cy}" desc:"position of the center of the circle"`
	Radius float32  `xml:"r" desc:"radius of the circle"`
}

var KiT_Circle = kit.Types.AddType(&Circle{}, nil)

// AddNewCircle adds a new button to given parent node, with given name, x,y pos, and radius.
func AddNewCircle(parent ki.Ki, name string, x, y, radius float32) *Circle {
	g := parent.AddNewChild(KiT_Circle, name).(*Circle)
	g.Pos.Set(x, y)
	g.Radius = radius
	return g
}

func (g *Circle) Render2D() {
	if g.Viewport == nil {
		g.This().(gi.Node2D).Init2D()
	}
	pc := &g.Pnt
	rs := &g.Viewport.Render
	rs.Lock()
	rs.PushXForm(pc.XForm)
	pc.DrawCircle(rs, g.Pos.X, g.Pos.Y, g.Radius)
	pc.FillStrokeClear(rs)
	rs.Unlock()

	g.ComputeBBoxSVG()
	g.Render2DChildren()

	rs.PopXFormLock()
}
