package chain

import (
	"context"
	"testing"
)

func TestDialRejectsUnknownTransport(t *testing.T) {
	if _, err := Dial(context.Background(), "ftp://localhost:1"); err == nil {
		t.Fatal("Dial accepted a URL with no known transport")
	}
}
