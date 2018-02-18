package cmd

import (
	"testing"
)

func TestEmpty(t *testing.T) {
	var tests = []struct {
		port  Port
		empty bool
	}{
		{Port{}, true},
		{Port{80, ""}, true},
		{Port{0, "HTTP"}, true},
		{Port{80, "HTTP"}, false},
	}

	for _, test := range tests {
		if test.port.Empty() != test.empty {
			t.Errorf("expected port %s empty == %t, got %t", test.port, test.empty, test.port.Empty())
		}
	}
}

func TestString(t *testing.T) {
	var tests = []struct {
		port Port
		out  string
	}{
		{Port{}, ""},
		{Port{80, ""}, ""},
		{Port{0, "HTTP"}, ""},
		{Port{80, "HTTP"}, "HTTP:80"},
		{Port{25, "TCP"}, "TCP:25"},
	}

	for _, test := range tests {
		if test.port.String() != test.out {
			t.Errorf("expected port %s == %s, got %s", test.port, test.out, test.port.String())
		}
	}
}
