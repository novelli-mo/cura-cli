package main

import "testing"

func TestSomething(t *testing.T) {
	result := 2 + 2
	if result != 4 {
		t.Errorf("expected 4, got %d", result)
	}
}
