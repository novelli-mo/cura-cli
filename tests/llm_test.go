package tests

import (
	"errors"
	"testing"

	"github.com/novelli-mo/cura/llm"
)

type mockCaller struct {
	response string
	err      error
}

func (m mockCaller) Call(prompt string) (string, error) {
	return m.response, m.err
}

func TestCaller_ReturnsResponse(t *testing.T) {
	caller := mockCaller{response: `{"skills": ["go-backend", "cli-tool"]}`}

	got, err := caller.Call(llm.BuildPrompt("some summary"))
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"skills": ["go-backend", "cli-tool"]}` {
		t.Errorf("unexpected response: %s", got)
	}
}

func TestCaller_PropagatesError(t *testing.T) {
	caller := mockCaller{err: errors.New("provider unavailable")}

	_, err := caller.Call(llm.BuildPrompt("some summary"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
