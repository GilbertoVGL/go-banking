package server

import (
	"strconv"
	"strings"
	"testing"
)

func TestNewAnyPort(t *testing.T) {
	port := 666
	srv, err := New(port)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(srv.Addr, strconv.Itoa(port)) {
		t.Fatalf("Mismatching addr, got: %s, expected: %d", srv.Addr, port)
	}
}
