// Copyright (c) 2019, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi3d

import (
	"image"
	"image/draw"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/mat32"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// Text2D presents 2D rendered text on a vertically-oriented plane.
// Call RenderText to update if text changes.
// This automatically multiplies scale by size of rendered text in pixels, so
// a uniform scale multiplier should generally be set to achieve desired size (.005 default).
// Standard styling properties can be set on the node to set font size etc
// note that higher quality is achieved by using a larger font size (36 default).
// The margin property creates blank margin of the background color around the text (2 px default)
// and the background-color defaults to transparent but can be set to any color.
type Text2D struct {
	Object
	Text      string        `desc:"the text string to display"`
	Sty       gi.Style      `json:"-" xml:"-" desc:"styling settings for the text"`
	TxtPos    gi.Vec2D      `xml:"-" json:"-" desc:"position offset of start of text rendering relative to upper-left corner"`
	TxtRender gi.TextRender `view:"-" xml:"-" json:"-" desc:"render data for text label"`
	TxtTex    *TextureBase  `view:"-" xml:"-" json:"-" desc:"texture object for the text -- this is used directly instead of pointing to the Scene Texture resources"`
}

var KiT_Text2D = kit.Types.AddType(&Text2D{}, nil)

// AddNewText2D adds a new object of given name and text string to given parent
func AddNewText2D(sc *Scene, parent ki.Ki, name string, text string) *Text2D {
	txt := parent.AddNewChild(KiT_Text2D, name).(*Text2D)
	tm := sc.Text2DPlaneMesh()
	txt.SetMesh(sc, tm)
	txt.Defaults()
	txt.Text = text
	return txt
}

func (txt *Text2D) Defaults() {
	txt.Object.Defaults()
	txt.Pose.Scale.SetScalar(.005)
	txt.SetProp("font-size", units.NewPt(36))
	txt.SetProp("margin", units.NewPx(2))
	txt.SetProp("color", "black")
	txt.SetProp("background-color", gi.Color{0, 0, 0, 0})
	txt.Mat.Bright = 5 // this is key for making e.g., a white background show up as white..
}

func (txt *Text2D) Init3D(sc *Scene) {
	txt.RenderText(sc)
	err := txt.Validate(sc)
	if err != nil {
		txt.SetInvisible()
	}
	txt.Node3DBase.Init3D(sc)
}

// StyleText does basic 2D styling
func (txt *Text2D) StyleText(sc *Scene) {
	txt.Sty.Defaults()
	txt.Sty.SetStyleProps(nil, *txt.Properties(), sc.Viewport)
	pagg := txt.ParentCSSAgg()
	if pagg != nil {
		gi.AggCSS(&txt.CSSAgg, *pagg)
	} else {
		txt.CSSAgg = nil // restart
	}
	// css stuff only works for node2d
	// gi.AggCSS(&txt.CSSAgg, txt.CSS)
	// txt.Sty.StyleCSS(txt.This().(gi.Node2D), txt.CSSAgg, "", sc.Viewport)
	txt.Sty.SetUnitContext(sc.Viewport, gi.Vec2DZero)
}

func (txt *Text2D) RenderText(sc *Scene) {
	txt.StyleText(sc)
	txt.TxtRender.SetHTML(txt.Text, &txt.Sty.Font, &txt.Sty.Text, &txt.Sty.UnContext, txt.CSSAgg)
	sz := txt.TxtRender.Size
	txt.TxtRender.LayoutStdLR(&txt.Sty.Text, &txt.Sty.Font, &txt.Sty.UnContext, sz)
	if txt.TxtRender.Size != sz {
		sz = txt.TxtRender.Size
		txt.TxtRender.LayoutStdLR(&txt.Sty.Text, &txt.Sty.Font, &txt.Sty.UnContext, sz)
		if txt.TxtRender.Size != sz {
			sz = txt.TxtRender.Size
		}
	}
	marg := txt.Sty.Layout.Margin.Dots
	sz.SetAddVal(2 * marg)
	txt.TxtPos.SetVal(marg)
	szpt := sz.ToPoint()
	bounds := image.Rectangle{Max: szpt}
	var img *image.RGBA
	if txt.TxtTex == nil {
		txt.TxtTex = &TextureBase{Nm: txt.Nm}
		tx := txt.TxtTex.NewTex()
		img = image.NewRGBA(bounds)
		tx.SetImage(img)
	} else {
		img = txt.TxtTex.Tex.Image().(*image.RGBA)
	}
	txt.TxtTex.Tex.SetSize(szpt)
	rs := &sc.RenderState
	rs.Init(szpt.X, szpt.Y, img)
	rs.PushBounds(bounds)
	draw.Draw(rs.Image, bounds, &image.Uniform{txt.Sty.Font.BgColor.Color}, image.ZP, draw.Src)
	txt.TxtRender.Render(rs, txt.TxtPos)
	rs.PopBounds()
	rs.Image = nil
	txt.Mat.SetTexture(sc, txt.TxtTex)
	gi.SavePNG("text-test.png", img)
}

// Validate checks that object has valid mesh and texture settings, etc
func (txt *Text2D) Validate(sc *Scene) error {
	// todo: validate more stuff here
	return txt.Object.Validate(sc)
}

func (txt *Text2D) UpdateWorldMatrix(parWorld *mat32.Mat4) {
	if txt.TxtTex != nil {
		txt.Pose.Defaults()
		tsz := txt.TxtTex.Tex.Size()
		szsc := mat32.Vec3{float32(tsz.X), float32(tsz.Y), 1}.Mul(txt.Pose.Scale)
		txt.Pose.Matrix.SetTransform(txt.Pose.Pos, txt.Pose.Quat, szsc)
	} else {
		txt.Pose.UpdateMatrix()
	}
	txt.Pose.UpdateWorldMatrix(parWorld)
	txt.SetFlag(int(WorldMatrixUpdated))
}

func (txt *Text2D) IsTransparent() bool {
	if txt.Sty.Font.BgColor.Color.A < 255 {
		return true
	}
	return false
}

func (txt *Text2D) RenderClass() RenderClasses {
	if txt.IsTransparent() {
		return RClassTransTexture
	}
	return RClassOpaqueTexture
}
