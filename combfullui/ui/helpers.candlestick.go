package main

import "image/png"
import "fmt"
import "bytes"
import "encoding/base64"

func CandlestickChart(data [][2]uint64, width uint64, minx, maxx, miny, maxy, zoom int64) string {

	if width < 100 {
		width = 100
	}
	if zoom < 0 {
		zoom = 0
	}
	var zooms = [][2]uint64{
		{1, 30 * 144},
		{3, 30 * 144},
		{1, 7 * 144},
		{3, 7 * 144},
		{6, 7 * 144},
		{1, 144},
		{3, 144},
		{1, 36},
		{3, 36},
		{1, 6},
		{3, 6},
		{6, 6},
		{3, 2},
		{6, 2},
		{12, 2},
		{24, 2},
	}
	if zoom >= int64(len(zooms)) {
		zoom = int64(len(zooms) - 1)
	}

	csv := [][]float64{}
	for _, dat := range data {
		var d = []float64{float64(dat[0]), float64(dat[1]) * 100000000 / float64(Coinbase(dat[0]))}
		d = append(d, d[1], d[1], d[1])
		csv = append(csv, d)

	}

	c, _ := NewChart(width, 600, 0, 0, width-50, 580, zooms[zoom][0],
		minx, maxx, miny, maxy, nil)

	c.DrawGrid(nil)

	for _, candle := range ConvertPeriodically(csv, int64(zooms[zoom][1]), 1, [5]uint{0, 1, 1, 1, 1}) {
		c.Draw(&candle)
	}

	c.DrawAxisY(int64(width-25), func(y int64) string {
		return fmt.Sprintf("%.0f", float64(y))
	})
	c.DrawAxisX(590, func(x int64) string {
		return fmt.Sprintf("%.0f", float64(x))
	})

	buf := bytes.NewBufferString("data:image/png;base64,")
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	png.Encode(encoder, c.Image())
	// Must close the encoder when finished to flush any partial blocks.
	// If you comment out the following line, the last partial block "r"
	// won't be encoded.
	encoder.Close()
	return buf.String()

}
