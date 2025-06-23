package main

import "image"
import "errors"

type Chart struct {
	pcc                                  PriceCandleCanvas
	mintime, maxtime, minprice, maxprice int64
	zoom, minx, miny, maxx, maxy         uint64
}

var ErrZeroSize = errors.New("zero size")
var ErrMaxXOutOfBounds = errors.New("max x out of bounds")
var ErrMaxYOutOfBounds = errors.New("max y out of bounds")
var ErrMinXOutOfBounds = errors.New("min x out of bounds")
var ErrMinYOutOfBounds = errors.New("min y out of bounds")
var ErrPriceOutOfBounds = errors.New("price out of bounds")
var ErrTimeOutOfBounds = errors.New("time out of bounds")

func NewChart(width, height, minx, miny, maxx, maxy, zoom uint64,
	mintime, maxtime, minprice, maxprice int64,
	options *PriceCandleCanvasOptions) (*Chart, error) {
	if width == 0 || height == 0 {
		return nil, ErrZeroSize
	}
	if maxx > width {
		return nil, ErrMaxXOutOfBounds
	}
	if maxy > height {
		return nil, ErrMaxYOutOfBounds
	}
	if minx > maxx {
		return nil, ErrMinXOutOfBounds
	}
	if miny > maxy {
		return nil, ErrMinYOutOfBounds
	}
	if minprice >= maxprice {
		return nil, ErrPriceOutOfBounds
	}
	if mintime >= maxtime {
		return nil, ErrTimeOutOfBounds
	}
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(width), int(height)}})

	return &Chart{
		pcc: PriceCandleCanvas{
			Img:     img,
			Rect:    image.Rectangle{image.Point{int(minx), int(miny)}, image.Point{int(maxx), int(maxy)}},
			Zoom:    int64(zoom),
			Options: options,
		},
		mintime: mintime, maxtime: maxtime, minprice: minprice, maxprice: maxprice,
		zoom: zoom, minx: minx, miny: miny, maxx: maxx, maxy: maxy,
	}, nil
}

func (c *Chart) DrawNumberX(n string, x, y int64) {
	if c == nil {
		return
	}
	var time = c.maxtime - c.mintime
	var timerange = int64(c.maxx) - int64(c.minx)

	var X = int64(c.minx) + timerange*(x-c.mintime)/time
	var Y = y
	var l = int64(len(n))
	for i := int64(0); i < l; i++ {
		var ch = byte(255)
		for j := range characters {
			if n[i] == characters[j] {
				ch = byte(j)
				break
			}
		}
		if ch == 255 {
			continue
		}
		(&Digit{Character: ch, X: X + 5*i + 3 - 5*l/2, Y: Y}).Draw(c.pcc)
	}
}
func (c *Chart) DrawNumberY(n string, x, y int64) {
	if c == nil {
		return
	}
	var price = c.maxprice - c.minprice
	var pricerange = int64(c.maxy) - int64(c.miny)

	var X = x
	var Y = int64(c.maxy) - pricerange*(y-c.minprice)/price
	var l = int64(len(n))
	for i := int64(0); i < l; i++ {
		var ch = byte(255)
		for j := range characters {
			if n[i] == characters[j] {
				ch = byte(j)
				break
			}
		}
		if ch == 255 {
			continue
		}
		(&Digit{Character: ch, X: X + 5*i + 3 - 5*l/2, Y: Y}).Draw(c.pcc)
	}
}

func (c *Chart) DrawAxisY(x int64, axisFunc func(int64) string) {
	if c == nil {
		return
	}
	var modulo = uint64(c.maxprice-c.minprice) / 16
	var price = c.maxprice - c.minprice
	var pricerange = int64(c.maxy) - int64(c.miny)
	if price != 0 {
		var pstep = pricerange / (256 * price)
		if pstep < 1 {
			pstep = 1
		}
		for p := c.minprice; p <= c.maxprice; p += pstep {
			if modulo > 0 {
				if p%int64(modulo) != 0 {
					continue
				}
			}
			c.DrawNumberY(axisFunc(p), x, p)
		}
	}
}

func (c *Chart) DrawAxisX(y int64, axisFunc func(int64) string) {
	if c == nil {
		return
	}
	var modulo = uint64(c.maxtime-c.mintime) / 4
	var time = c.maxtime - c.mintime
	var timerange = int64(c.maxx) - int64(c.minx)
	if time != 0 {
		var tstep = timerange / (256 * time)
		if tstep < 1 {
			tstep = 1
		}
		for t := c.mintime; t <= c.maxtime; t += tstep {
			if modulo > 0 {
				if t%int64(modulo) != 0 {
					continue
				}
			}
			c.DrawNumberX(axisFunc(t), t, y)
		}
	}
}

func (c *Chart) Draw(candle *PriceCandle) {
	if candle == nil {
		return
	}
	if c == nil {
		return
	}
	var transformed PriceCandle

	var time = c.maxtime - c.mintime
	var timerange = int64(c.maxx) - int64(c.minx)

	transformed.X = int64(c.minx) + timerange*(candle.X-c.mintime)/time

	if !c.pcc.VisibleX(transformed.X, c.minx, c.maxx) {
		return
	}

	var price = c.maxprice - c.minprice
	var pricerange = int64(c.maxy) - int64(c.miny)

	transformed.High = int64(c.maxy) - pricerange*(candle.High-c.minprice)/price
	transformed.Low = int64(c.maxy) - pricerange*(candle.Low-c.minprice)/price
	transformed.Open = int64(c.maxy) - pricerange*(candle.Open-c.minprice)/price
	transformed.Close = int64(c.maxy) - pricerange*(candle.Close-c.minprice)/price

	transformed.Draw(c.pcc)
}

func (c *Chart) DrawGrid(g *PriceGrid) {
	if g == nil {
		g = &PriceGrid{
			uint64(c.maxtime-c.mintime) / 16, uint64(c.maxprice-c.minprice) / 16, nil,
		}
	}
	if c == nil {
		return
	}

	opts := g.Options
	if opts == nil {
		opts = &DefaultPriceGridOptions
	}

	var time = c.maxtime - c.mintime
	var timerange = int64(c.maxx) - int64(c.minx)
	var price = c.maxprice - c.minprice
	var pricerange = int64(c.maxy) - int64(c.miny)

	var xline = func(y int64) {
		for x := c.minx; x <= c.maxx; x++ {
			c.pcc.Img.Set(int(x), int(y), opts.Color)
		}
	}
	var yline = func(x int64) {
		for y := c.miny; y <= c.maxy; y++ {
			c.pcc.Img.Set(int(x), int(y), opts.Color)
		}
	}
	xline(int64(c.miny))
	xline(int64(c.maxy))
	yline(int64(c.minx))
	yline(int64(c.maxx))
	if price != 0 {
		var pstep = pricerange / (256 * price)
		if pstep < 1 {
			pstep = 1
		}
		for p := c.minprice; p <= c.maxprice; p += pstep {
			if g.YPriceModulo > 0 {
				if p%int64(g.YPriceModulo) != 0 {
					continue
				}
			}
			var transformedP = int64(c.maxy) - pricerange*(p-c.minprice)/price
			xline(transformedP)
		}
	}
	if time != 0 {
		var tstep = timerange / (256 * time)
		if tstep < 1 {
			tstep = 1
		}
		for t := c.mintime; t <= c.maxtime; t += tstep {
			if g.XTimeModulo > 0 {
				if t%int64(g.XTimeModulo) != 0 {
					continue
				}
			}
			var transformedT = int64(c.minx) + timerange*(t-c.mintime)/time
			yline(transformedT)
		}
	}
}

func (c *Chart) Image() image.Image {
	return c.pcc.Img
}
