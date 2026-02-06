package main

import (
	"bytes"
	"os"
	"testing"
)

func TestBasementOutput(t *testing.T) {
	// Read the expected output from the existing test file
	expectedBytes, err := os.ReadFile("../test/demo-basement")
	if err != nil {
		t.Fatalf("Failed to read expected output file: %v", err)
	}
	expected := string(expectedBytes)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the demo function (which prints to stdout)
	demo()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	actual := buf.String()

	// Compare
	if actual != expected {
		// Write actual output to a file for debugging if they differ
		_ = os.WriteFile("actual_output.txt", []byte(actual), 0644)
		t.Errorf("Output mismatch.\nExpected length: %d\nActual length: %d\nSee actual_output.txt for details.", len(expected), len(actual))
	}
}
