package catalog

const PredicateGTPacked = "predicate-gt-packed-f32"

func packedPredicateOperation() Operation {
	op := predicateOperation()
	op.ID, op.Group, op.Title = PredicateGTPacked, "predicatepack", "Ordered predicate export (packed bits)"
	op.BenchmarkName = "BenchmarkPredicatePacked"
	op.ArithmeticID = "float32-predicate-gt-packed-lsb0-v1"
	op.Expression = "bit i%8 of out[i/8] = x[i] > threshold; unused high bits = 0"
	op.Output = Value{"out", Uint8, PackedBits, 1}
	op.NumericalPolicy = "Ordered strict comparison; equality or NaN yields bit 0, true yields bit 1. Input i maps to bit i%8 of byte i/8 (LSB first). Output length ceil(n/8); unused high bits zero; every byte overwritten. SIMD selects/converts explicit 0/1 and packs valid lanes, carrying partial bytes across vectors. Validation, conversion, packing and stores timed. Disjoint storage; invalid output length fails before writes. Empty succeeds. Logical bytes 4*n+ceil(n/8); no raw SDK mask encoding."
	return op
}
