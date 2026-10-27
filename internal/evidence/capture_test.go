package evidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/result"
)

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }

func TestCaptureEncodingAndWriterFailure(t *testing.T) {
	duration := int64(2)
	capture := Capture{Schema: Schema, Status: "complete", Run: result.Run{ID: "run", Observations: []result.Observation{{Iterations: 5, DurationNS: &duration, NanosecondsPerOp: 0.4}}}}
	var output bytes.Buffer
	if err := WriteCapture(&output, capture); err != nil {
		t.Fatal(err)
	}
	var decoded Capture
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil || decoded.Schema != Schema || *decoded.Run.Observations[0].DurationNS != 2 || decoded.Run.Observations[0].NanosecondsPerOp != 0.4 {
		t.Fatalf("precision lost: %v %s", err, &output)
	}
	want := errors.New("disk full")
	if err := WriteCapture(failedWriter{want}, capture); !errors.Is(err, want) {
		t.Fatalf("writer error=%v", err)
	}
	capture.Schema = "unknown"
	if err := WriteCapture(&output, capture); err == nil {
		t.Fatal("unknown schema accepted")
	}
	capture.Schema = Schema
	capture.Status = "interrupted"
	if err := WriteCapture(&output, capture); err == nil {
		t.Fatal("unknown status accepted")
	}
}
