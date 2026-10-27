package cli

import (
	"fmt"
	"math"
	"strconv"
)

type float32Flag struct{ bits *uint32 }

func (f float32Flag) String() string {
	if f.bits == nil {
		return "0"
	}
	return strconv.FormatFloat(float64(math.Float32frombits(*f.bits)), 'g', -1, 32)
}
func (f float32Flag) Set(value string) error {
	parsed, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return fmt.Errorf("invalid float32: %w", err)
	}
	*f.bits = math.Float32bits(float32(parsed))
	return nil
}

func InputSummary(opParameters string, opGenerator string) string {
	return "parameters: " + opParameters + "; fixture: " + opGenerator
}
