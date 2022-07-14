package main

import (
	"fyne.io/fyne/v2/test"
	"testing"
)

func TestGreeting(t *testing.T) {
	out, in := makeUI()
	if out.Text != "Hello World" {
		t.Errorf("Expected %q got %q", "Hello World", out.Text)
	}

	test.Type(in, "Andy")
	if out.Text != "Hello Andy" {
		t.Errorf("Expected %q got %q", "Hello Andy", out.Text)
	}
}
