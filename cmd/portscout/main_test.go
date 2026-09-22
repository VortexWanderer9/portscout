package main

import "testing"

func TestBoolCount(t *testing.T) {
	if got := boolCount(true, false, true); got != 2 {
		t.Errorf("boolCount() = %d, want 2", got)
	}
}
