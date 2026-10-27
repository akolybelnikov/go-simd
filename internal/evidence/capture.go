// Package evidence encodes direct measurement files without SIMD or execution.
package evidence

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/akolybelnikov/go-simd/internal/result"
)

const Schema = "go-simd.measurements/v1"

type Config struct {
	Count     int    `json:"count"`
	Benchtime string `json:"benchtime"`
	Bench     string `json:"bench"`
}

type Machine struct {
	GoVersion      string           `json:"go_version"`
	OS             string           `json:"os"`
	Arch           string           `json:"arch"`
	CPU            string           `json:"cpu"`
	Bits           int              `json:"vector_bits"`
	Lanes          int              `json:"float32_lanes"`
	Procs          int              `json:"gomaxprocs"`
	Emulated       bool             `json:"emulated"`
	SourceRevision string           `json:"source_revision"`
	SourceModified string           `json:"source_modified"`
	Build          *debug.BuildInfo `json:"build,omitempty"`
}

// Capture is the minimal direct-result evidence file, not M4's worker manifest.
// Run fields retain the internal model's Go field names in this v1 encoding.
type Capture struct {
	Schema  string     `json:"schema"`
	Status  string     `json:"status"`
	Error   string     `json:"error,omitempty"`
	Group   string     `json:"group"`
	Config  Config     `json:"config"`
	Machine Machine    `json:"machine"`
	Run     result.Run `json:"run"`
}

func WriteCapture(w io.Writer, c Capture) error {
	if c.Schema != Schema || (c.Status != "complete" && c.Status != "failed") {
		return fmt.Errorf("invalid measurement capture header")
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}
