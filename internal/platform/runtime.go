// Package platform captures facts from this executable, without invoking Go tools.
package platform

import (
	"runtime"
	"runtime/debug"
	"simd"

	"github.com/akolybelnikov/go-simd/internal/evidence"
)

type Info = evidence.Machine

func Current() Info {
	var vector simd.Float32s
	i := Info{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: "unknown", Bits: simd.VectorBitSize(), Lanes: vector.Len(), Procs: runtime.GOMAXPROCS(0), Emulated: simd.Emulated(), SourceRevision: "unknown", SourceModified: "unknown"}
	if build, ok := debug.ReadBuildInfo(); ok {
		i.Build = build
		for _, setting := range build.Settings {
			switch setting.Key {
			case "vcs.revision":
				i.SourceRevision = setting.Value
			case "vcs.modified":
				i.SourceModified = setting.Value
			}
		}
	}
	return i
}
