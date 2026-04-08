package helpers

import "testing"

func TestTruncate_Short(t *testing.T) {
	if got := Truncate("hello", 10); got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}

func TestTruncate_Exact(t *testing.T) {
	if got := Truncate("hello", 5); got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}

func TestTruncate_Long(t *testing.T) {
	if got := Truncate("hello world", 5); got != "hello..." {
		t.Fatalf("expected hello..., got %s", got)
	}
}

func TestTruncate_UTF8(t *testing.T) {
	input := "🎉🎊🎈🎁🎂"
	got := Truncate(input, 3)
	want := "🎉🎊🎈..."
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestTruncate_Empty(t *testing.T) {
	if got := Truncate("", 5); got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

func TestParseDeps_Valid(t *testing.T) {
	deps := ParseDeps(`["a","b"]`)
	if len(deps) != 2 || deps[0] != "a" || deps[1] != "b" {
		t.Fatalf("unexpected: %v", deps)
	}
}

func TestParseDeps_Empty(t *testing.T) {
	for _, input := range []string{"", "null", "[]"} {
		if deps := ParseDeps(input); deps != nil {
			t.Fatalf("expected nil for %q, got %v", input, deps)
		}
	}
}

func TestParseDeps_InvalidJSON(t *testing.T) {
	if deps := ParseDeps("{broken"); deps != nil {
		t.Fatalf("expected nil for invalid JSON, got %v", deps)
	}
}

func TestToJSON(t *testing.T) {
	got := ToJSON([]string{"a", "b"})
	if got != `["a","b"]` {
		t.Fatalf("expected [\"a\",\"b\"], got %s", got)
	}
}

func TestContainsDep(t *testing.T) {
	json := `["task-1","task-2"]`
	if !ContainsDep(json, "task-1") {
		t.Fatal("expected true")
	}
	if ContainsDep(json, "task-3") {
		t.Fatal("expected false")
	}
}
