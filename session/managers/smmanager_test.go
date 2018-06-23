package managers

import (
	"testing"
)

func TestGenerateSessionID(t *testing.T) {
	sid := generateSessionID()
	if len(sid) < sessionIDLength {
		t.Errorf("Expected generated session id to be at least length: %v - found: %v", sessionIDLength, len(sid))
	}
}
