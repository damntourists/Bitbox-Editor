package util

import (
	"errors"
	"testing"
)

func TestPanicOnError(t *testing.T) {
	defer func() { recover() }()

	err := errors.New("some sort of panic")
	PanicOnError(err)

	t.Errorf("did not panic")
}
