package sanitize

import "testing"

func TestApplyRedactsKnownSecrets(t *testing.T) {
	s := New(LevelBalanced)
	in := "ghp_1234567890 token=abc Authorization: Bearer test"
	out, rep := s.Apply(in)
	if out == in {
		t.Fatalf("expected sanitization to modify input")
	}
	if rep.Level != LevelBalanced {
		t.Fatalf("unexpected level: %s", rep.Level)
	}
}
