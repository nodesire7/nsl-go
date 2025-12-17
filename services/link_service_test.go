package services

import (
	"testing"
)

func TestGenerateRandomCodeLength(t *testing.T) {
	s := NewLinkService()
	code := s.GenerateRandomCode(6)
	if len(code) != 6 {
		t.Fatalf("expected length 6, got %d", len(code))
	}
}


