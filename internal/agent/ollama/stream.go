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
	pendingEvents []agent.ModelEvent
}

func (s *modelStream) Next() (agent.ModelEvent, error) {
	for {
		var resp chatResponse

		if len(s.pendingEvents) > 0 {
			e := s.pendingEvents[0]
			s.pendingEvents = s.pendingEvents[1:]
			return e, nil
		}

		if err := s.decoder.Decode(&resp); err != nil {
			return agent.ModelEvent{}, fmt.Errorf("error decoding body: %w", err)
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
					ID:        fmt.Sprintf("ollama-call-%d", toolCall.Function.Index),
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments.Clone(),
				},
			})
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
	return &modelStream{
		body:    body,
		decoder: json.NewDecoder(body),
	}
}
