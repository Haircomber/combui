package main

import "image/color"
import "image"

var DefaultPriceCandleCanvasOptions = PriceCandleCanvasOptions{
	DecreaseColor: color.RGBA{217, 88, 99, 255},
	IncreaseColor: color.RGBA{38, 174, 117, 255},
	ClearColor:    color.RGBA{0, 0, 0, 0},
	DecreaseFill:  true,
	IncreaseFill:  false,
}

type PriceCandleCanvasOptions struct {
	DecreaseColor, IncreaseColor, ClearColor color.Color
	DecreaseFill, IncreaseFill               bool
}

type PriceCandleCanvas struct {
	Options *PriceCandleCanvasOptions
	Zoom    int64
	Rect    image.Rectangle
	Img     *image.RGBA
}

func (c *PriceCandleCanvas) VisibleX(x int64, x0, x1 uint64) bool {

	if int64(x0)-c.Zoom > x {
		return false
	}
	if int64(x1)+c.Zoom < x {
		return false
	}
	return true
}
