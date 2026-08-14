package organization

import (
	"context"
	"testing"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func associatedClient(clientID string, useForMemberAccess bool) *managementv3.OrganizationClient {
	return &managementv3.OrganizationClient{
		ClientID:           clientID,
		UseForMemberAccess: useForMemberAccess,
	}
}

func desiredClient(clientID string, useForMemberAccess bool) *managementv3.CreateOrganizationClientRequestItem {
	return &managementv3.CreateOrganizationClientRequestItem{
		ClientID:           clientID,
		UseForMemberAccess: useForMemberAccess,
	}
}

// TestDiffOrganizationClients covers reconciling the configured associations against what the
// API reports, including where the organization was changed outside Terraform.
func TestDiffOrganizationClients(t *testing.T) {
	var testCases = []struct {
		name     string
		current  []*managementv3.OrganizationClient
		desired  []*managementv3.CreateOrganizationClientRequestItem
		toAdd    []*managementv3.CreateOrganizationClientRequestItem
		toUpdate []*managementv3.CreateOrganizationClientRequestItem
		toRemove []string
	}{
		{
			name:    "associates every client when the organization has none",
			current: nil,
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-b", true),
				desiredClient("client-a", false),
			},
			toAdd: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
				desiredClient("client-b", true),
			},
		},
		{
			name: "does nothing when the organization already matches",
			current: []*managementv3.OrganizationClient{
				associatedClient("client-a", false),
				associatedClient("client-b", true),
			},
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
				desiredClient("client-b", true),
			},
		},
		{
			name: "disassociates a client the configuration no longer lists",
			current: []*managementv3.OrganizationClient{
				associatedClient("client-a", false),
				associatedClient("client-b", true),
			},
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
			},
			toRemove: []string{"client-b"},
		},
		{
			name: "patches a client whose use_for_member_access differs",
			current: []*managementv3.OrganizationClient{
				associatedClient("client-a", false),
			},
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", true),
			},
			toUpdate: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", true),
			},
		},
		{
			// Re-created rather than patched, which a diff against prior state would get wrong.
			name:    "re-associates a client that was disassociated outside terraform",
			current: nil,
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", true),
			},
			toAdd: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", true),
			},
		},
		{
			// Removed rather than left in place, so the state after the apply matches the plan.
			name: "disassociates a client that was associated outside terraform",
			current: []*managementv3.OrganizationClient{
				associatedClient("client-a", false),
				associatedClient("client-rogue", true),
			},
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
			},
			toRemove: []string{"client-rogue"},
		},
		{
			// Patched back, never a disassociate plus re-associate, which would be a 409.
			name: "patches a client whose flag was flipped outside terraform",
			current: []*managementv3.OrganizationClient{
				associatedClient("client-a", true),
			},
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
			},
			toUpdate: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
			},
		},
		{
			name: "combines an association, a patch and a disassociation",
			current: []*managementv3.OrganizationClient{
				associatedClient("client-keep", false),
				associatedClient("client-flip", false),
				associatedClient("client-drop", true),
			},
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-keep", false),
				desiredClient("client-flip", true),
				desiredClient("client-new", true),
			},
			toAdd: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-new", true),
			},
			toUpdate: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-flip", true),
			},
			toRemove: []string{"client-drop"},
		},
		{
			name: "orders the calls by client id so that an apply is reproducible",
			current: []*managementv3.OrganizationClient{
				associatedClient("client-9", false),
				associatedClient("client-1", false),
			},
			desired: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-c", false),
				desiredClient("client-a", false),
				desiredClient("client-b", false),
			},
			toAdd: []*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
				desiredClient("client-b", false),
				desiredClient("client-c", false),
			},
			toRemove: []string{"client-1", "client-9"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			diff := diffOrganizationClients(testCase.current, testCase.desired)

			assert.Equal(t, testCase.toAdd, diff.toAdd)
			assert.Equal(t, testCase.toUpdate, diff.toUpdate)
			assert.Equal(t, testCase.toRemove, diff.toRemove)
		})
	}
}

// organizationClientsDiffError runs the plan time validation over the given `clients`
// configuration through the real diff machinery, so values arrive as Terraform delivers them.
func organizationClientsDiffError(t *testing.T, clients []cty.Value) error {
	t.Helper()

	resource := NewClientsResource()

	rawConfig := cty.ObjectVal(map[string]cty.Value{
		"id":              cty.UnknownVal(cty.String),
		"organization_id": cty.StringVal("org_1234567890123456"),
		"clients":         cty.SetVal(clients),
	})

	// RawConfig goes on the state, not the config: ResourceDiff.GetRawConfig reads what the
	// SDK recorded on the diff, and that is copied over from the state.
	state := &terraform.InstanceState{
		Attributes: map[string]string{},
		RawConfig:  rawConfig,
	}

	_, err := resource.Diff(
		context.Background(),
		state,
		terraform.NewResourceConfigShimmed(rawConfig, resource.CoreConfigSchema()),
		nil,
	)

	return err
}

func configuredClient(clientID cty.Value, useForMemberAccess bool) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"client_id":             clientID,
		"use_for_member_access": cty.BoolVal(useForMemberAccess),
	})
}

// TestValidateOrganizationClientsDiff covers the plan time duplicate check. IDs only known
// after apply must be skipped: the SDK hands them over as one and the same placeholder, so
// treating them as ordinary values rejects clients created in the same apply.
func TestValidateOrganizationClientsDiff(t *testing.T) {
	t.Run("accepts clients whose IDs are only known after apply", func(t *testing.T) {
		assert.NoError(t, organizationClientsDiffError(t, []cty.Value{
			configuredClient(cty.UnknownVal(cty.String), true),
			configuredClient(cty.UnknownVal(cty.String), false),
		}))
	})

	t.Run("accepts a known client alongside one known only after apply", func(t *testing.T) {
		assert.NoError(t, organizationClientsDiffError(t, []cty.Value{
			configuredClient(cty.StringVal("client-a"), true),
			configuredClient(cty.UnknownVal(cty.String), false),
		}))
	})

	t.Run("accepts distinct known clients", func(t *testing.T) {
		assert.NoError(t, organizationClientsDiffError(t, []cty.Value{
			configuredClient(cty.StringVal("client-a"), true),
			configuredClient(cty.StringVal("client-b"), false),
		}))
	})

	t.Run("rejects the same known client listed twice", func(t *testing.T) {
		assert.ErrorContains(t, organizationClientsDiffError(t, []cty.Value{
			configuredClient(cty.StringVal("client-a"), true),
			configuredClient(cty.StringVal("client-a"), false),
		}), `client_id "client-a" is listed more than once`)
	})
}

func TestValidateOrganizationClientsAreUnique(t *testing.T) {
	t.Run("accepts a set that names every client once", func(t *testing.T) {
		assert.NoError(t, validateOrganizationClientsAreUnique(
			[]*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
				desiredClient("client-b", true),
			},
		))
	})

	t.Run("rejects a set that names the same client twice", func(t *testing.T) {
		err := validateOrganizationClientsAreUnique(
			[]*managementv3.CreateOrganizationClientRequestItem{
				desiredClient("client-a", false),
				desiredClient("client-a", true),
			},
		)

		assert.ErrorContains(t, err, `client_id "client-a" is listed more than once`)
	})
}
