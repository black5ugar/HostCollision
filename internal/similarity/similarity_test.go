package similarity

import "testing"

// TestScoreIdentical verifies that identical inputs produce a perfect score.
func TestScoreIdentical(t *testing.T) {
	a := []byte("Hello, world!")
	b := []byte("Hello, world!")

	score := Score(a, b)
	if score != 100 {
		t.Fatalf("expected 100, got %d", score)
	}
}

// TestScoreEmpty verifies the behavior for empty inputs.
func TestScoreEmpty(t *testing.T) {
	if score := Score(nil, nil); score != 100 {
		t.Fatalf("expected 100 for two empty inputs, got %d", score)
	}

	if score := Score([]byte("data"), nil); score != 0 {
		t.Fatalf("expected 0 when only one side is empty, got %d", score)
	}
}

// TestScoreDifferent verifies that clearly different inputs result in a lower score.
func TestScoreDifferent(t *testing.T) {
	a := []byte("aaaaaaaaaaaaaa")
	b := []byte("bbbbbbbbbbbbbb")

	score := Score(a, b)
	if score >= 80 {
		t.Fatalf("expected score < 80 for very different inputs, got %d", score)
	}
}
