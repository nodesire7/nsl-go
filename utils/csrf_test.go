package utils

import "testing"

func TestGenerateCSRFToken(t *testing.T) {
	t1, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken error: %v", err)
	}
	t2, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken error: %v", err)
	}
	if t1 == "" || t2 == "" {
		t.Fatalf("token should not be empty")
	}
	if t1 == t2 {
		t.Fatalf("tokens should be different")
	}
	// 32 bytes hex => 64 chars
	if len(t1) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(t1))
	}
}


