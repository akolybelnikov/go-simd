package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestBrowserChatEvaluation records real replies and the rubric for human grading.
// It exercises the same authenticated endpoint and context as the browser.
func TestBrowserChatEvaluation(t *testing.T) {
	model := os.Getenv("GO_SIMD_CHAT_EVAL_MODEL")
	if model == "" {
		t.Skip("set GO_SIMD_CHAT_EVAL_MODEL to an installed Ollama model")
	}
	app := testLocalUI(t)
	provider := newChatProvider("ollama:"+model, model, "")
	require.NoError(t, provider.requireLocalModel(t.Context()))
	app.chatProviders = []chatProvider{provider}

	capture, err := os.ReadFile(filepath.Join("testdata", "sum-cancellation.measurements.json"))
	require.NoError(t, err)
	opened := postUI(app, "/api/open", string(capture))
	require.Equal(t, http.StatusOK, opened.Code, opened.Body.String())
	var result uiResult
	require.NoError(t, json.Unmarshal(opened.Body.Bytes(), &result))
	batch := chatSelection{ResultRef: result.ChatRef, Size: result.Series[0].Size}

	cases := []struct {
		name, question, rubric string
		selection              chatSelection
		history                []chatMessage
	}{
		{"concept", "What is SIMD? Is it the same as using more cores?",
			"Explain one instruction operating on multiple values and distinguish CPU cores. No draft.", chatSelection{}, nil},
		{"matching-operation", "I have float32 sensor readings and want to clip them to [-1, 1]. Which operation fits?",
			"Choose float32 clamp and describe clamping. A comparison is optional.", chatSelection{}, nil},
		{"unsupported-type", "I need to sum float64 with that precision preserved. Which SIMD Lab operation should I run?",
			"Say float64 sum is unavailable. Do not substitute float32 sum or emit a draft.", chatSelection{}, nil},
		{"ambiguous-workload", "Which operation do I use to clean my measurements?",
			"Ask what cleaning means and what the data type is. No draft.", chatSelection{}, nil},
		{"correctness", "Does this sum run prove that ordinary and vectorized sums always return the same result?",
			"Say no. Explain that rounding and summation order can differ. No draft.", batch, nil},
		{"winner", "Which method had the lower median in this batch?",
			"Identify the vectorized sum as lower, consistent with the recorded 40% reduction. No draft.", batch, nil},
		{"variability", "Do these repeats prove a statistically significant win?",
			"Say no, explain two repeats per method are limited, and avoid a confidence claim. No draft.", batch, nil},
		{"bottleneck", "Can you prove I was limited by cache misses or memory bandwidth?",
			"Say the timing does not establish a bottleneck. Suggest counters, a controlled size sweep, or profiling without inventing counter readings. No draft.", batch, nil},
		{"different-host", "Will this get the same speedup on my laptop?",
			"Say this run describes its recorded host and a separate measurement is needed. No draft.", batch, nil},
		{"different-size", "Which method won at 80,000 values?",
			"Say 80,000 was not measured in the selected result. Do not reuse the current winner for that size. No draft.", batch, nil},
		{"requested-draft", "Prepare a comparison of ordinary sum and the one-accumulator vectorized sum at 80,000 values and 30 repeats.",
			"Return a valid sum/simd-striped draft with size 80000 and 30 repeats. Leave run approval to the user.", batch, nil},
		{"follow-up-draft", "Keep that operation, method, and batch size, but use 30 repeats.",
			"Return the previous addition/simd-partial draft at 4096 values with 30 repeats.", chatSelection{}, []chatMessage{
				{Role: "user", Content: "Suggest a comparison for float32 addition at 4096 values."},
				{Role: "assistant", Content: "Suggested comparison: {\"group\":\"addition\",\"method\":\"simd-partial\",\"size\":4096,\"repeats\":3}"},
			}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(uiChatRequest{
				Provider: provider.id, Selection: tc.selection, Question: tc.question, History: tc.history,
			})
			require.NoError(t, err)
			start := time.Now()
			reply := postUI(app, "/api/chat", string(payload))
			t.Logf("model=%s elapsed=%s status=%d\nrubric: %s\nreply: %s",
				model, time.Since(start).Round(time.Millisecond), reply.Code, tc.rubric, reply.Body.String())
			require.Nil(t, app.pending, "chat must never create a run approval")
		})
	}
}
