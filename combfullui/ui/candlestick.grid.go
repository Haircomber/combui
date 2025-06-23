package main

import "image/color"

var DefaultPriceGridOptions = PriceGridOptions{
	Color: color.RGBA{77, 77, 77, 255},
}

type PriceGridOptions struct {
	Color color.Color
}

type PriceGrid struct {
	XTimeModulo  uint64
	YPriceModulo uint64
	Options      *PriceGridOptions
}
