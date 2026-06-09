package test

import (
	"testing"
)

func Expect[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("Expected %v; got %v", want, got)
	}
}

func ExpectFatal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("Expected %v; got %v", want, got)
	}
}

