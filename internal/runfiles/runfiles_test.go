package runfiles

import (
	"strings"
	"testing"
)

func TestSafeMetadata(t *testing.T) {
	if got := Sanitize("../A:B\\C?* á"); got != "___A_B_C____" {
		t.Fatalf("sanitize: %q", got)
	}
	for _, raw := range []string{"mqtt://user:secret@localhost:1883", "mqtt://user:secret@localhost", "mqtt://localhost:1883"} {
		if strings.Contains(RedactURL(raw), "secret") {
			t.Fatal("URL password leaked")
		}
	}
}
