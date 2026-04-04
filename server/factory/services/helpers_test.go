package services

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello..."},
		{"empty", "", 5, ""},
		{"unicode", "Hello, \u4e16\u754c\uff01abc", 9, "Hello, \u4e16\u754c..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestParseDeps(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // expected length, -1 for nil
	}{
		{"empty string", "", -1},
		{"null", "null", -1},
		{"empty array", "[]", -1},
		{"one dep", `["abc"]`, 1},
		{"two deps", `["a","b"]`, 2},
		{"invalid json", `{bad`, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDeps(tt.input)
			if tt.want == -1 {
				if got != nil {
					t.Errorf("ParseDeps(%q) = %v, want nil", tt.input, got)
				}
			} else if len(got) != tt.want {
				t.Errorf("ParseDeps(%q) len = %d, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"strings", []string{"a", "b"}, `["a","b"]`},
		{"empty", []string{}, "[]"},
		{"nil", nil, "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToJSON(tt.input)
			if got != tt.want {
				t.Errorf("ToJSON(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContainsDep(t *testing.T) {
	tests := []struct {
		name    string
		deps    string
		taskID  string
		want    bool
	}{
		{"found", `["a","b","c"]`, "b", true},
		{"not found", `["a","b"]`, "c", false},
		{"empty", "[]", "a", false},
		{"null", "null", "a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsDep(tt.deps, tt.taskID)
			if got != tt.want {
				t.Errorf("ContainsDep(%q, %q) = %v, want %v", tt.deps, tt.taskID, got, tt.want)
			}
		})
	}
}
