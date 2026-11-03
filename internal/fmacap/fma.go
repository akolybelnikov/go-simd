// Package fmacap probes the selected SIMD runtime against binary32 FMA rules.
// It owns process capability facts independently of the operation kernels.
package fmacap

import (
	"math"
	"simd"

	"github.com/akolybelnikov/go-simd/internal/fma32"
)

var vectorFMA32 = probeVectorFMA32()

// Path reports the cached process gate, not a guarantee of instruction speed.
func Path() string {
	if vectorFMA32 {
		return fma32.VectorPath
	}
	return fma32.SoftwarePath
}

func probeVectorFMA32() bool {
	// This SDK's emulated MulAdd is ordinary x*y+z, without a fused contract.
	if simd.Emulated() {
		return false
	}
	u := float32(0x1p-23)
	probes := [][3]float32{
		{1 + u, 1 - u, -1}, {1 + u, 1 - u, 16777218},
		{float32(math.Ldexp(float64(1+u), -75)), float32(math.Ldexp(float64(1-u), -75)), math.Float32frombits(0x007fffff)},
		{float32(math.Ldexp(float64(1+u), 51)), float32(math.Ldexp(float64(1-u), 52)), math.MaxFloat32},
		{math.Float32frombits(0x80000000), 1, math.Float32frombits(0x80000000)},
	}
	for _, p := range probes {
		got := simd.BroadcastFloat32s(p[0]).MulAdd(simd.BroadcastFloat32s(p[1]), simd.BroadcastFloat32s(p[2]))
		var scratch [16]float32
		lanes := scratch[:]
		if got.Len() > len(lanes) {
			lanes = make([]float32, got.Len())
		}
		got.Store(lanes)
		want := math.Float32bits(fma32.FMA(p[0], p[1], p[2]))
		for _, v := range lanes[:got.Len()] {
			if math.Float32bits(v) != want {
				return false
			}
		}
	}
	return true
}

// VectorAvailable reports the cached gate without probing inside timed calls.
func VectorAvailable() bool { return vectorFMA32 }
