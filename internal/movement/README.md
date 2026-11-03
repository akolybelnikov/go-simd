# Float32 data movement

CopyScalar, CopyBuiltin, CopySIMD and CopySIMDVectorTail accept
`(x,out []float32) error` and preserve exact bits, including signed zero and
NaN payloads. Unequal lengths return ErrLengthMismatch before writes. Equal
empty/nil succeeds. Disjoint output and exact in-place alias are supported;
partial overlap is outside the common contract.

The copy group compares a source loop, Go's built-in copy, partial-vector
loads/stores and full vectors with a scalar tail. Scalar/GoCopy labels do not
prove machine instruction selection; GoCopy uses compiler/runtime optimization.
Measured buffers are disjoint. All calls include validation and movement/tails,
with zero per-call allocation. Logical work is `8*n` bytes.

Typed unary bindings allocate one input and one output before timing. Dedicated
copy harnesses retain the complete output after timing and record SHA-256 over
all little-endian float32 bits, alongside input identity, SDK lanes and bit policy.
SDK emulation does not disable Go's independently selected built-in-copy code.
Percentages use the source-loop Scalar baseline; inspect GoCopy as another control.

```powershell
go run ./examples/copy
go run ./cmd/simdreport -list -group copy
go run ./cmd/simdreport -group copy -bench '^BenchmarkCopy$/^n=64$' -count 3 -benchtime 100ms -format md
```

Installed commands and Make use the same options. See the
[feature record](../../docs/features/F004-copy.md) and
[Wiki guide](https://github.com/akolybelnikov/go-simd/wiki/21-Copy-and-data-movement).
