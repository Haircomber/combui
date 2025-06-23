package main

func max5(a, b, c, d, e uint) (o uint) {
	o = a
	if b > o {
		o = b
	}
	if c > o {
		o = c
	}
	if d > o {
		o = d
	}
	if e > o {
		o = e
	}
	return
}
func max2(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
func min2(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func ConvertPeriodically(csv [][]float64, period int64, mul float64, idx [5]uint) (out []PriceCandle) {
	var lastx int64
	for _, data := range csv {
		var maxidx = max5(idx[0], idx[1], idx[2], idx[3], idx[4])
		if uint(len(data)) <= maxidx {
			continue
		}
		var x = (int64(data[idx[0]]) / period) * period
		if x < lastx {
			return
		}
		if (len(out) > 0) && (lastx == x) {
			var top = &out[len(out)-1]
			top.High = max2(top.High, int64(data[idx[2]]*mul))
			top.Low = min2(top.Low, int64(data[idx[3]]*mul))
			top.Close = int64(data[idx[4]] * mul)
		} else {
			out = append(out, PriceCandle{
				X:     x,
				Open:  int64(data[idx[1]] * mul),
				High:  int64(data[idx[2]] * mul),
				Low:   int64(data[idx[3]] * mul),
				Close: int64(data[idx[4]] * mul),
			})
		}
		lastx = x
	}
	return
}

func ConvertCombDaily(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 144, 1, [5]uint{0, 2, 2, 2, 2})
}
func ConvertComb2Daily(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 2*144, 1, [5]uint{0, 2, 2, 2, 2})
}
func ConvertCombWeekly(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 7*144, 1, [5]uint{0, 2, 2, 2, 2})
}

func ConvertComb2Weekly(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 2*7*144, 1, [5]uint{0, 2, 2, 2, 2})
}

func ConvertWeekly(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 7*24*60*60, 100, [5]uint{0, 1, 2, 3, 4})
}

func Convert2Daily(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 2*24*60*60, 100, [5]uint{0, 1, 2, 3, 4})
}

func ConvertDaily(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 24*60*60, 100, [5]uint{0, 1, 2, 3, 4})
}

func ConvertHourly(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 60*60, 100, [5]uint{0, 1, 2, 3, 4})
}

func Convert15Minly(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 15*60, 100, [5]uint{0, 1, 2, 3, 4})
}

func Convert5Minly(csv [][]float64) (out []PriceCandle) {
	return ConvertPeriodically(csv, 5*60, 100, [5]uint{0, 1, 2, 3, 4})
}