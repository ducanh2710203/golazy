package golazy

import "testing"

func TestStaticValue(t *testing.T) {
	s := Static(123)
	v, err := s.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 123 {
		t.Fatalf("expected 123, got %v", v)
	}

	// Clear should be a no-op
	s.Clear()
	v2, _ := s.Value()
	if v2 != 123 {
		t.Fatalf("expected 123 after Clear, got %v", v2)
	}
}
