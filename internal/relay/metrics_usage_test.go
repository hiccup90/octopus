package relay

import "testing"

func TestIngestUsageFromRawJSON_OpenAI(t *testing.T) {
	m := NewRelayMetrics(1, "gpt", nil)
	m.ingestUsageFromRawJSON([]byte(`{"model":"gpt-4o","usage":{"prompt_tokens":11,"completion_tokens":22}}`))
	if m.Stats.InputToken != 11 || m.Stats.OutputToken != 22 {
		t.Fatalf("openai usage got in=%d out=%d", m.Stats.InputToken, m.Stats.OutputToken)
	}
	if m.ActualModel != "gpt-4o" {
		t.Fatalf("model=%q", m.ActualModel)
	}
}

func TestIngestUsageFromRawJSON_AnthropicNonStream(t *testing.T) {
	m := NewRelayMetrics(1, "claude", nil)
	m.ingestUsageFromRawJSON([]byte(`{"type":"message","model":"claude-sonnet","usage":{"input_tokens":100,"output_tokens":50}}`))
	if m.Stats.InputToken != 100 || m.Stats.OutputToken != 50 {
		t.Fatalf("anthropic usage got in=%d out=%d", m.Stats.InputToken, m.Stats.OutputToken)
	}
}

func TestIngestUsageFromRawJSON_AnthropicStream(t *testing.T) {
	m := NewRelayMetrics(1, "claude", nil)
	// message_start carries input
	m.ingestUsageFromRawJSON([]byte(`{"type":"message_start","message":{"model":"claude-x","usage":{"input_tokens":33,"output_tokens":0}}}`))
	// message_delta carries output; must not wipe input
	m.ingestUsageFromRawJSON([]byte(`{"type":"message_delta","usage":{"output_tokens":44}}`))
	if m.Stats.InputToken != 33 || m.Stats.OutputToken != 44 {
		t.Fatalf("stream usage got in=%d out=%d", m.Stats.InputToken, m.Stats.OutputToken)
	}
	if m.ActualModel != "claude-x" {
		t.Fatalf("model=%q", m.ActualModel)
	}
}
