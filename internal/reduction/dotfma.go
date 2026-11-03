package reduction

import (
	"simd"

	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
)

// DotFMAScalar rounds each x*y+total once to binary32, left to right from +0.
func DotFMAScalar(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	var total float32
	for i, v := range x {
		total = fma32.FMA(v, y[i], total)
	}
	return total, nil
}

// DotFMASIMD uses 1 striped FMA accumulator, a left-to-right
// lane-wise merge and low-to-high lane fold, then software FMA tail updates.
// Software execution preserves this order rather than substituting scalar order.
func DotFMASIMD(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	if len(x) == 0 {
		return 0, nil
	}
	var a0 simd.Float32s
	w := a0.Len()
	if !fmacap.VectorAvailable() {
		return dotFMASoftwareStriped(x, y, w, 1), nil
	}
	i := 0
	for ; i+1*w <= len(x); i += 1 * w {
		a0 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), a0)
	}
	merged := a0
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	merged.Store(lanes)
	var total float32
	for _, v := range lanes[:w] {
		total = float32(total + v)
	}
	for ; i < len(x); i++ {
		total = fma32.FMA(x[i], y[i], total)
	}
	return total, nil
}

// DotFMASIMD2 uses 2 striped FMA accumulators, a left-to-right
// lane-wise merge and low-to-high lane fold, then software FMA tail updates.
// Software execution preserves this order rather than substituting scalar order.
func DotFMASIMD2(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	if len(x) == 0 {
		return 0, nil
	}
	var a0, a1 simd.Float32s
	w := a0.Len()
	if !fmacap.VectorAvailable() {
		return dotFMASoftwareStriped(x, y, w, 2), nil
	}
	i := 0
	for ; i+2*w <= len(x); i += 2 * w {
		a0 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), a0)
		a1 = simd.LoadFloat32s(x[i+w:]).MulAdd(simd.LoadFloat32s(y[i+w:]), a1)
	}
	if i+w <= len(x) {
		a0 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), a0)
		i += w
	}
	merged := a0.Add(a1)
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	merged.Store(lanes)
	var total float32
	for _, v := range lanes[:w] {
		total = float32(total + v)
	}
	for ; i < len(x); i++ {
		total = fma32.FMA(x[i], y[i], total)
	}
	return total, nil
}

// DotFMASIMD4 uses 4 striped FMA accumulators, a left-to-right
// lane-wise merge and low-to-high lane fold, then software FMA tail updates.
// Software execution preserves this order rather than substituting scalar order.
func DotFMASIMD4(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	if len(x) == 0 {
		return 0, nil
	}
	var a0, a1, a2, a3 simd.Float32s
	w := a0.Len()
	if !fmacap.VectorAvailable() {
		return dotFMASoftwareStriped(x, y, w, 4), nil
	}
	i := 0
	for ; i+4*w <= len(x); i += 4 * w {
		a0 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), a0)
		a1 = simd.LoadFloat32s(x[i+w:]).MulAdd(simd.LoadFloat32s(y[i+w:]), a1)
		a2 = simd.LoadFloat32s(x[i+2*w:]).MulAdd(simd.LoadFloat32s(y[i+2*w:]), a2)
		a3 = simd.LoadFloat32s(x[i+3*w:]).MulAdd(simd.LoadFloat32s(y[i+3*w:]), a3)
	}
	if i+w <= len(x) {
		a0 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), a0)
		i += w
	}
	if i+w <= len(x) {
		a1 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), a1)
		i += w
	}
	if i+w <= len(x) {
		a2 = simd.LoadFloat32s(x[i:]).MulAdd(simd.LoadFloat32s(y[i:]), a2)
		i += w
	}
	merged := a0.Add(a1).Add(a2).Add(a3)
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	merged.Store(lanes)
	var total float32
	for _, v := range lanes[:w] {
		total = float32(total + v)
	}
	for ; i < len(x); i++ {
		total = fma32.FMA(x[i], y[i], total)
	}
	return total, nil
}

// Fixed current widths/counts use stack storage; a larger future width exposes
// its scratch allocation. No invalid input lanes participate in the reduction.
func dotFMASoftwareStriped(x, y []float32, width, count int) float32 {
	var scratch [64]float32
	accumulators := scratch[:]
	needed := width * count
	if needed > len(accumulators) {
		accumulators = make([]float32, needed)
	}
	accumulators = accumulators[:needed]
	full := len(x) / width
	for v := 0; v < full; v++ {
		base := (v % count) * width
		for lane := 0; lane < width; lane++ {
			i := v*width + lane
			accumulators[base+lane] = fma32.FMA(x[i], y[i], accumulators[base+lane])
		}
	}
	var total float32
	for lane := 0; lane < width; lane++ {
		merged := accumulators[lane]
		for a := 1; a < count; a++ {
			merged = float32(merged + accumulators[a*width+lane])
		}
		total = float32(total + merged)
	}
	for i := full * width; i < len(x); i++ {
		total = fma32.FMA(x[i], y[i], total)
	}
	return total
}
