package reduction

import "testing"

func TestSoftwareFusedDotStripesAcrossWidths(t *testing.T) {
	for _, width := range []int{4, 8, 16, 32} {
		for _, count := range []int{1, 2, 4} {
			for _, layout := range []int{1, 2} {
				x, y := make([]float32, 8*width+1), make([]float32, 8*width+1)
				for i := range y {
					y[i] = 1
				}
				x[0], x[layout*width], x[2*layout*width] = 16777216, 1, -16777216
				want := float32(0)
				if count == 2 && layout == 1 || count == 4 && layout == 2 {
					want = 1
				}
				if got := dotFMASoftwareStriped(x, y, width, count); got != want {
					t.Fatal("software stripe assignment", width, count, layout, got)
				}
			}
		}
	}
}
