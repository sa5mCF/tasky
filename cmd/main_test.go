package main

import (
	"path/filepath"
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

func TestResolveDataFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	databasePath, legacyJSONPath, err := resolveDataFiles()
	if err != nil {
		t.Fatalf("resolveDataFiles failed: %v", err)
	}

	if databasePath != filepath.Join(home, ".dataTodo.db") {
		t.Fatalf("unexpected database path: %q", databasePath)
	}
	if legacyJSONPath != filepath.Join(home, ".dataTodo.json") {
		t.Fatalf("unexpected legacy JSON path: %q", legacyJSONPath)
	}
}
