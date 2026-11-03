package reduction

import "simd"

// SumSIMD2 assigns full vectors round-robin to 2 independent accumulators.
// Merge accumulators lane-wise left to right, fold lanes low to high from +0,
// then process the scalar tail. Each addition rounds to float32.
func SumSIMD2(x []float32) float32 {
	if len(x) == 0 {
		return 0
	}
	var a0, a1 simd.Float32s
	w := a0.Len()
	i := 0
	for ; i+2*w <= len(x); i += 2 * w {
		a0 = a0.Add(simd.LoadFloat32s(x[i:]))
		a1 = a1.Add(simd.LoadFloat32s(x[i+w:]))
	}
	if i+w <= len(x) {
		a0 = a0.Add(simd.LoadFloat32s(x[i:]))
		i += w
	}
	merged := a0.Add(a1)
	// Extraction and tail stay in this kernel; current widths use stack
	// scratch. A larger future width exposes its allocation in counters.
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	merged.Store(lanes)
	var total float32
	for _, value := range lanes[:w] {
		total = float32(total + value)
	}
	for ; i < len(x); i++ {
		total = float32(total + x[i])
	}
	return total
}

// SumSIMD4 assigns full vectors round-robin to 4 independent accumulators.
// Merge accumulators lane-wise left to right, fold lanes low to high from +0,
// then process the scalar tail. Each addition rounds to float32.
func SumSIMD4(x []float32) float32 {
	if len(x) == 0 {
		return 0
	}
	var a0, a1, a2, a3 simd.Float32s
	w := a0.Len()
	i := 0
	for ; i+4*w <= len(x); i += 4 * w {
		a0 = a0.Add(simd.LoadFloat32s(x[i:]))
		a1 = a1.Add(simd.LoadFloat32s(x[i+w:]))
		a2 = a2.Add(simd.LoadFloat32s(x[i+2*w:]))
		a3 = a3.Add(simd.LoadFloat32s(x[i+3*w:]))
	}
	if i+w <= len(x) {
		a0 = a0.Add(simd.LoadFloat32s(x[i:]))
		i += w
	}
	if i+w <= len(x) {
		a1 = a1.Add(simd.LoadFloat32s(x[i:]))
		i += w
	}
	if i+w <= len(x) {
		a2 = a2.Add(simd.LoadFloat32s(x[i:]))
		i += w
	}
	merged := a0.Add(a1).Add(a2).Add(a3)
	// Extraction and tail stay in this kernel; current widths use stack
	// scratch. A larger future width exposes its allocation in counters.
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	merged.Store(lanes)
	var total float32
	for _, value := range lanes[:w] {
		total = float32(total + value)
	}
	for ; i < len(x); i++ {
		total = float32(total + x[i])
	}
	return total
}

// DotSIMD2 assigns full vectors round-robin to 2 independent accumulators.
// Merge accumulators lane-wise left to right, fold lanes low to high from +0,
// then process the scalar tail. Products round before addition; no FMA.
func DotSIMD2(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	if len(x) == 0 {
		return 0, nil
	}
	var a0, a1 simd.Float32s
	w := a0.Len()
	i := 0
	for ; i+2*w <= len(x); i += 2 * w {
		p0 := simd.LoadFloat32s(x[i:]).Mul(simd.LoadFloat32s(y[i:]))
		a0 = a0.Add(p0)
		p1 := simd.LoadFloat32s(x[i+w:]).Mul(simd.LoadFloat32s(y[i+w:]))
		a1 = a1.Add(p1)
	}
	if i+w <= len(x) {
		product := simd.LoadFloat32s(x[i:]).Mul(simd.LoadFloat32s(y[i:]))
		a0 = a0.Add(product)
		i += w
	}
	merged := a0.Add(a1)
	// Extraction and tail stay in this kernel; current widths use stack
	// scratch. A larger future width exposes its allocation in counters.
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	merged.Store(lanes)
	var total float32
	for _, value := range lanes[:w] {
		total = float32(total + value)
	}
	for ; i < len(x); i++ {
		product := float32(x[i] * y[i])
		total = float32(total + product)
	}
	return total, nil
}

// DotSIMD4 assigns full vectors round-robin to 4 independent accumulators.
// Merge accumulators lane-wise left to right, fold lanes low to high from +0,
// then process the scalar tail. Products round before addition; no FMA.
func DotSIMD4(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	if len(x) == 0 {
		return 0, nil
	}
	var a0, a1, a2, a3 simd.Float32s
	w := a0.Len()
	i := 0
	for ; i+4*w <= len(x); i += 4 * w {
		p0 := simd.LoadFloat32s(x[i:]).Mul(simd.LoadFloat32s(y[i:]))
		a0 = a0.Add(p0)
		p1 := simd.LoadFloat32s(x[i+w:]).Mul(simd.LoadFloat32s(y[i+w:]))
		a1 = a1.Add(p1)
		p2 := simd.LoadFloat32s(x[i+2*w:]).Mul(simd.LoadFloat32s(y[i+2*w:]))
		a2 = a2.Add(p2)
		p3 := simd.LoadFloat32s(x[i+3*w:]).Mul(simd.LoadFloat32s(y[i+3*w:]))
		a3 = a3.Add(p3)
	}
	if i+w <= len(x) {
		product := simd.LoadFloat32s(x[i:]).Mul(simd.LoadFloat32s(y[i:]))
		a0 = a0.Add(product)
		i += w
	}
	if i+w <= len(x) {
		product := simd.LoadFloat32s(x[i:]).Mul(simd.LoadFloat32s(y[i:]))
		a1 = a1.Add(product)
		i += w
	}
	if i+w <= len(x) {
		product := simd.LoadFloat32s(x[i:]).Mul(simd.LoadFloat32s(y[i:]))
		a2 = a2.Add(product)
		i += w
	}
	merged := a0.Add(a1).Add(a2).Add(a3)
	// Extraction and tail stay in this kernel; current widths use stack
	// scratch. A larger future width exposes its allocation in counters.
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	merged.Store(lanes)
	var total float32
	for _, value := range lanes[:w] {
		total = float32(total + value)
	}
	for ; i < len(x); i++ {
		product := float32(x[i] * y[i])
		total = float32(total + product)
	}
	return total, nil
}
