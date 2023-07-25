package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestActionMain(t *testing.T) {
	// Redirect stdout to capture the printed message

	buf := executeFunction(main)

	// Check if the printed message is correct
	expected := "Hello from my Golang GitHub Action!\n"
	if buf.String() != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, buf.String())
	}
}

func executeFunction(testFunction func()) bytes.Buffer {

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	testFunction()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf
}
