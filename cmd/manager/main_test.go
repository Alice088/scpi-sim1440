package main

import (
	"context"
	"testing"
)

func TestRunRejectsUnknownAction(t *testing.T) {
	if err := run(context.Background(), "bogus"); err == nil {
		t.Fatal("run(bogus) error = nil, want error")
	}
}
