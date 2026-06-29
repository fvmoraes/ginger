package handlers

import "testing"

func TestExistingUserHandlerIsPreserved(t *testing.T) {
	// This pre-existing test intentionally occupies user_test.go. A scan plan
	// must report it as skipped rather than replacing it.
}
