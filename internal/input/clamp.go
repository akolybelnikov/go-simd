package input

import "math"

// Clamp fixtures count ordered outside-range values as clipped. Non-clipped
// values include endpoints, zeros where valid, and unchanged NaN payloads.
func (o Options) clampValues() (inside, below, above float32) {
	lo, hi := o.Float("lower"), o.Float("upper")
	inside = lo
	if lo < hi {
		switch {
		case math.IsInf(float64(lo), -1) && math.IsInf(float64(hi), 1):
			inside = 0
		case math.IsInf(float64(lo), -1):
			inside = math.Nextafter32(hi, lo)
		case math.IsInf(float64(hi), 1):
			inside = math.Nextafter32(lo, hi)
		default:
			inside = float32((float64(lo) + float64(hi)) / 2)
		}
	}
	return inside, math.Nextafter32(lo, float32(math.Inf(-1))), math.Nextafter32(hi, float32(math.Inf(1)))
}

func (o Options) clampValue(i, n int) float32 {
	lo, hi := o.Float("lower"), o.Float("upper")
	inside, below, above := o.clampValues()
	if o.trueAt(i, n) {
		if below < lo && ((i/2)%2 == 0 || !(above > hi)) {
			return below
		}
		return above
	}
	switch i % 8 {
	case 1:
		return lo
	case 2:
		return hi
	case 3, 4:
		zero := float32(0)
		if i%8 == 3 {
			zero = math.Float32frombits(0x80000000)
		}
		if zero >= lo && zero <= hi {
			return zero
		}
	case 5:
		return math.Float32frombits(0x7fc12345)
	case 6:
		return math.Float32frombits(0x7f812345)
	}
	return inside
}
