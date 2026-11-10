package tcpportscan

import (
	"reflect"
	"testing"
)

func TestPorts(t *testing.T) {
	got, err := Ports(80, 443, 80, Range(80, 82))
	if err != nil {
		t.Fatal(err)
	}

	want := []uint16{80, 443, 81, 82}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPostInvalid(t *testing.T) {
	tests := []struct {
		name   string
		values []any
	}{
		{"zero", []any{0}},
		{"negative", []any{-1}},
		{"too large", []any{65536}},
		{"reversed range", []any{Range(100, 10)}},
		{"invalid range", []any{Range(1, 65536)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Ports(tt.values...)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestPortsInvalidRanges(t *testing.T) {
	tests := []struct {
		name  string
		input PortRange
	}{
		{"zero start", Range(0, 10)},
		{"zero end", Range(1, 0)},
		{"port above limit", Range(1, 65536)},
		{"both ports above limit", Range(70000, 80000)},
		{"negative start", Range(-5, 10)},
		{"reversed range", Range(10, 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Ports(tt.input)

			if err == nil {
				t.Fatalf("expected error for range %d-%d", tt.input.From, tt.input.To)
			}
		})
	}
}
