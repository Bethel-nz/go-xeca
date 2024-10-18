package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func StringContains(t *testing.T, actual, expectedSubString string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubString) {
		t.Errorf("got: %q; expected to contain: %q", actual, expectedSubString)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v; expected: nil", actual)
	}
}

func NotNil(t *testing.T, actual interface{}) {
	t.Helper()

	if actual == nil {
		t.Error("expected non-nil value, got nil")
	}
}

func True(t *testing.T, actual bool) {
	t.Helper()

	if !actual {
		t.Error("expected true, got false")
	}
}

func False(t *testing.T, actual bool) {
	t.Helper()

	if actual {
		t.Error("expected false, got true")
	}
}
