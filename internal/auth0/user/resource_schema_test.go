package user

import (
	"testing"
)

// Changing the connection of an existing user is not supported by the Auth0 Management API,
// so the attribute must force a replacement instead of an in-place update.
func TestUserConnectionNameForcesReplacement(t *testing.T) {
	connectionName := NewResource().Schema["connection_name"]

	if !connectionName.ForceNew {
		t.Fatal("expected connection_name to be ForceNew so that a change plans a destroy and create")
	}
}
