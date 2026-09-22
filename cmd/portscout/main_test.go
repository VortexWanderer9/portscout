package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunArgsVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := runArgs([]string{"version"}, &stdout, &stderr); got != 0 {
		t.Fatalf("runArgs(version) = %d, want 0", got)
	}
	if want := "portscout " + version + "\n"; stdout.String() != want {
		t.Errorf("version output = %q, want %q", stdout.String(), want)
	}
}

func TestRunArgsHelpCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if got := runArgs([]string{"help"}, &stdout, &stderr); got != 0 {
		t.Fatalf("runArgs(help) = %d, want 0", got)
	}
	if !strings.Contains(stderr.String(), "portscout scan [flags] <host>") {
		t.Errorf("help output missing scan usage: %s", stderr.String())
	}
}

func TestRunArgsLongFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	got := runArgs([]string{"scan", "--ports", "invalid", "--workers", "1", "localhost"}, &stdout, &stderr)
	if got != 2 {
		t.Fatalf("runArgs(long flags) = %d, want 2", got)
	}
	if !strings.Contains(stderr.String(), "invalid port") {
		t.Errorf("error output = %q, want invalid-port error", stderr.String())
	}
}

func TestValidFormat(t *testing.T) {
	if !validFormat("json") || validFormat("yaml") {
		t.Error("validFormat returned an unexpected result")
	}
}

func TestBoolCount(t *testing.T) {
	if got := boolCount(true, false, true); got != 2 {
		t.Errorf("boolCount() = %d, want 2", got)
	}
}
