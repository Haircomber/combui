package main

import "image"

type PriceCandle struct {
	X     int64
	High  int64
	Low   int64
	Open  int64
	Close int64
}

func (pc *PriceCandle) Draw(c PriceCandleCanvas) {
	opt := c.Options
	if opt == nil {
		opt = &DefaultPriceCandleCanvasOptions
	}
	var clear = opt.ClearColor
	var brush = opt.DecreaseColor
	var miny, maxy = pc.Open, pc.Close
	var fill = opt.DecreaseFill
	// the following check is reversed because higher price is lower y
	if pc.Open > pc.Close {
		miny, maxy = pc.Close, pc.Open
		brush = opt.IncreaseColor
		fill = opt.IncreaseFill
	}
	for y := miny; y <= maxy; y++ {
		var onBorderY = y == miny || y == maxy
		for x := pc.X - c.Zoom; x <= pc.X+c.Zoom; x++ {
			var onBorderX = x == pc.X-c.Zoom || x == pc.X+c.Zoom
			if (image.Point{int(x), int(y)}).In(c.Rect) {
				if onBorderY || onBorderX || fill {
					c.Img.Set(int(x), int(y), brush)
				} else {
					c.Img.Set(int(x), int(y), clear)
				}
			}
		}
	}
	for y := pc.High; y < miny; y++ {
		if (image.Point{int(pc.X), int(y)}).In(c.Rect) {
			c.Img.Set(int(pc.X), int(y), brush)
		}
	}
	for y := pc.Low; y > maxy; y-- {
		if (image.Point{int(pc.X), int(y)}).In(c.Rect) {
			c.Img.Set(int(pc.X), int(y), brush)
		}
	}
}
