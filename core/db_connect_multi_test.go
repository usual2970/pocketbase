package core

import "testing"

func TestVersionAtLeast(t *testing.T) {
	cases := []struct {
		v    string
		min  int
		want bool
	}{
		{"8.0.36", 8, true},
		{"5.7.42", 8, false},
		{"13.10", 13, true},
		{"12.9", 13, false},
		{"v14.1", 13, true},
		{"", 8, false},
		{"abc", 8, false},
	}
	for _, c := range cases {
		if got := versionAtLeast(c.v, c.min); got != c.want {
			t.Fatalf("versionAtLeast(%q,%d)=%v, want %v", c.v, c.min, got, c.want)
		}
	}
}

func TestErrUnsupportedDBVersion_Error(t *testing.T) {
	err := ErrUnsupportedDBVersion("mysql", "8.0+", "5.7.42")
	s := err.Error()
	if s == "" {
		t.Fatal("expected non-empty error string")
	}
	if want := "mysql"; !contains(s, want) {
		t.Fatalf("expected error to mention %q; got %q", want, s)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	// simple substring search to avoid importing strings in test helper
	n, m := len(s), len(sub)
	if m == 0 {
		return 0
	}
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
