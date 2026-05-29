package enola

import "testing"

func TestMatchUsernameRegex_GitHub(t *testing.T) {
	pattern := `^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$`

	tests := []struct {
		username string
		want     bool
	}{
		{"amirrossein", true},
		{"blue", true},
		{"-invalid", false},
		{"invalid-", false},
	}

	for _, tt := range tests {
		got, err := matchUsernameRegex(pattern, tt.username)
		if err != nil {
			t.Fatalf("matchUsernameRegex(%q) error: %v", tt.username, err)
		}
		if got != tt.want {
			t.Fatalf("matchUsernameRegex(%q) = %v, want %v", tt.username, got, tt.want)
		}
	}
}

func TestMatchUsernameRegex_Stdlib(t *testing.T) {
	got, err := matchUsernameRegex(`^[a-z]+$`, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("expected match")
	}
}

func TestMatchUsernameRegex_Empty(t *testing.T) {
	got, err := matchUsernameRegex("", "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("empty pattern should match anything")
	}
}
