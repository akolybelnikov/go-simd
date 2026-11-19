package ollama

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/akolybelnikov/go-simd/internal/agent"
)

type modelStream struct {
	body          io.ReadCloser
	decoder       *json.Decoder
	limitedBody   *io.LimitedReader
	pendingEvents []agent.ModelEvent
	nextCall      int
}

const maxResponseBytes = 1 << 20

func (s *modelStream) Next() (agent.ModelEvent, error) {
	for {
		var resp chatResponse

		if len(s.pendingEvents) > 0 {
			e := s.pendingEvents[0]
			s.pendingEvents = s.pendingEvents[1:]
			return e, nil
		}

		err := s.decoder.Decode(&resp)
		if s.limitedBody.N == 0 {
			return agent.ModelEvent{}, fmt.Errorf("Ollama response exceeds %d bytes", maxResponseBytes)
		}
		if err != nil {
			return agent.ModelEvent{}, fmt.Errorf("error decoding body: %w", err)
		}
		if resp.Error != "" {
			return agent.ModelEvent{}, fmt.Errorf("Ollama reported an error in the response stream")
		}

		if resp.Message.Content != "" {
			s.pendingEvents = append(s.pendingEvents, agent.ModelEvent{
				Kind: agent.EventTextDelta,
				Text: resp.Message.Content,
			})
		}

		for _, toolCall := range resp.Message.ToolCalls {
			s.pendingEvents = append(s.pendingEvents, agent.ModelEvent{
				Kind: agent.EventToolCall,
				ToolCall: &agent.ToolCall{
					ID:        fmt.Sprintf("ollama-call-%d", s.nextCall),
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments.Clone(),
				},
			})
			s.nextCall++
		}

		if resp.Done {
			s.pendingEvents = append(s.pendingEvents, agent.ModelEvent{Kind: agent.EventCompleted})
		}
	}
}

func (s *modelStream) Close() error {
	return s.body.Close()
}

func newModelStream(body io.ReadCloser) *modelStream {
	limited := &io.LimitedReader{R: body, N: maxResponseBytes + 1}
	return &modelStream{
		body:        body,
		decoder:     json.NewDecoder(limited),
		limitedBody: limited,
	}
}
