package metrics

import (
	"bytes"
	"strings"
	"testing"
)

func TestMetricsTrackingAndSnapshot(t *testing.T) {
	m := New()

	// Record requests of different statuses
	m.HTTPRequest("GET", 200)
	m.HTTPRequest("GET", 200)
	m.HTTPRequest("GET", 304)
	m.HTTPRequest("GET", 404)
	m.HTTPRequest("POST", 401)
	m.HTTPRequest("GET", 429)
	m.HTTPRequest("POST", 500)
	m.HTTPRequest("GET", 503)

	m.RequestDuration(10)
	m.RequestDuration(20)

	m.FetchCycle(false)
	m.FetchCycle(true)

	m.AICall(false)
	m.AICall(true)

	m.AITokens(120, 80, 200, "gpt-4o")
	m.AITokens(300, 150, 450, "gpt-4o")
	m.AITokens(50, 50, 100, "deepseek-chat")

	snap := m.Snapshot()

	if snap.HTTPRequestsTotal != 8 {
		t.Errorf("HTTPRequestsTotal = %d, want 8", snap.HTTPRequestsTotal)
	}
	if snap.HTTPErrorsTotal != 5 {
		t.Errorf("HTTPErrorsTotal = %d, want 5 (3 4xx + 2 5xx)", snap.HTTPErrorsTotal)
	}
	if snap.HTTP4xxTotal != 3 {
		t.Errorf("HTTP4xxTotal = %d, want 3", snap.HTTP4xxTotal)
	}
	if snap.HTTP5xxTotal != 2 {
		t.Errorf("HTTP5xxTotal = %d, want 2", snap.HTTP5xxTotal)
	}
	if snap.HTTPStatusCodes["200"] != 2 {
		t.Errorf("HTTPStatusCodes[200] = %d, want 2", snap.HTTPStatusCodes["200"])
	}
	if snap.HTTPStatusCodes["304"] != 1 {
		t.Errorf("HTTPStatusCodes[304] = %d, want 1", snap.HTTPStatusCodes["304"])
	}
	if snap.HTTPStatusCodes["404"] != 1 {
		t.Errorf("HTTPStatusCodes[404] = %d, want 1", snap.HTTPStatusCodes["404"])
	}
	if snap.HTTPStatusCodes["500"] != 1 {
		t.Errorf("HTTPStatusCodes[500] = %d, want 1", snap.HTTPStatusCodes["500"])
	}
	if snap.HTTPAvgLatencyMs != 15 {
		t.Errorf("HTTPAvgLatencyMs = %f, want 15.0", snap.HTTPAvgLatencyMs)
	}
	if snap.FetchCyclesTotal != 2 || snap.FetchCyclesFailed != 1 {
		t.Errorf("FetchCycles = %d/%d, want 2/1", snap.FetchCyclesTotal, snap.FetchCyclesFailed)
	}
	if snap.AICallsTotal != 2 || snap.AICallsFailed != 1 {
		t.Errorf("AICalls = %d/%d, want 2/1", snap.AICallsTotal, snap.AICallsFailed)
	}
	if snap.AIPromptTokens != 470 {
		t.Errorf("AIPromptTokens = %d, want 470", snap.AIPromptTokens)
	}
	if snap.AICompletionTokens != 280 {
		t.Errorf("AICompletionTokens = %d, want 280", snap.AICompletionTokens)
	}
	if snap.AITotalTokens != 750 {
		t.Errorf("AITotalTokens = %d, want 750", snap.AITotalTokens)
	}
	if snap.AITokensByModel["gpt-4o"] != 650 {
		t.Errorf("AITokensByModel[gpt-4o] = %d, want 650", snap.AITokensByModel["gpt-4o"])
	}
	if snap.AITokensByModel["deepseek-chat"] != 100 {
		t.Errorf("AITokensByModel[deepseek-chat] = %d, want 100", snap.AITokensByModel["deepseek-chat"])
	}

	var buf bytes.Buffer
	m.WritePrometheus(&buf)
	out := buf.String()

	if !strings.Contains(out, "neuralwire_http_4xx_errors_total 3") {
		t.Errorf("missing or incorrect 4xx errors in Prometheus output:\n%s", out)
	}
	if !strings.Contains(out, "neuralwire_http_5xx_errors_total 2") {
		t.Errorf("missing or incorrect 5xx errors in Prometheus output:\n%s", out)
	}
	if !strings.Contains(out, "neuralwire_http_errors_total 5") {
		t.Errorf("missing or incorrect total errors in Prometheus output:\n%s", out)
	}
	if !strings.Contains(out, "neuralwire_ai_tokens_total 750") {
		t.Errorf("missing or incorrect ai tokens in Prometheus output:\n%s", out)
	}
}

func TestMetricsNilSafety(t *testing.T) {
	var m *Metrics
	m.HTTPRequest("GET", 200)
	m.RequestDuration(5)
	m.FetchCycle(false)
	m.AICall(false)
	m.AITokens(10, 20, 30, "gpt-4o")

	snap := m.Snapshot()
	if snap.HTTPRequestsTotal != 0 {
		t.Errorf("snap.HTTPRequestsTotal = %d, want 0", snap.HTTPRequestsTotal)
	}

	var buf bytes.Buffer
	m.WritePrometheus(&buf)
	if buf.Len() != 0 {
		t.Errorf("expected empty Prometheus output for nil collector, got %d bytes", buf.Len())
	}
}
