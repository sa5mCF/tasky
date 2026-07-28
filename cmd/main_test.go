package main

import (
	"strings"
	"testing"
)

func TestGetInputFromArgs(t *testing.T) {
	got, err := getInput(strings.NewReader("ignored"), "buy", "milk")
	if err != nil {
		t.Fatalf("getInput failed: %v", err)
	}

	if got != "buy milk" {
		t.Fatalf("expected %q, got %q", "buy milk", got)
	}
}

func TestGetInputFromStdin(t *testing.T) {
	got, err := getInput(strings.NewReader("   clean inbox   \n"))
	if err != nil {
		t.Fatalf("getInput failed: %v", err)
	}

	if got != "clean inbox" {
		t.Fatalf("expected %q, got %q", "clean inbox", got)
	}
}

func TestGetInputRejectsEmptyText(t *testing.T) {
	if _, err := getInput(strings.NewReader("\n")); err == nil {
		t.Fatal("expected empty stdin input to fail")
	}

	if _, err := getInput(strings.NewReader("ignored"), "   "); err == nil {
		t.Fatal("expected whitespace-only args to fail")
	}
}
