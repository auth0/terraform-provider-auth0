---
page_title: Managing Forms & Flows with exported JSON
description: |-
	How to manage complex Forms and Flows resources in Terraform by exporting JSON from the Auth0 Dashboard
	and referencing it from concise auth0_form and auth0_flow resources.
---

# Managing Forms & Flows with exported JSON

Adopting Terraform for Auth0 Forms and Flows can feel cumbersome. These resources are rich, deeply nested
structures. Manually translating every dashboard change into Terraform often leads to drift, slow iteration,
and a reluctance to evolve existing user journeys.

This guide documents a pragmatic workflow: design and iterate in the Auth0 Dashboard, export Form JSON
snapshots, store them alongside Terraform, and reference their decoded structures inside lightweight
`auth0_form` and `auth0_flow` resources. Token placeholders (for flows and vault / connection references) are
replaced at plan/apply time, keeping Terraform authoritative without hand-maintaining large nested blocks.

~> This approach intentionally accepts duplication inside exported JSON files (repeated Flow and Vault
definitions). The duplication is benign; treat each export as a self‑contained snapshot.

## Pre-requisites

- Auth0 tenant access with permission to build Forms & Flows in the Dashboard UI.
- Terraform Auth0 Provider version that includes `auth0_form`, `auth0_flow`, and vault/connection resources.
- A working Terraform project structure (see the [Quickstart guide](./quickstart.md) if new to the provider).

## 1. Build Forms & Flows in the Dashboard

Use the Dashboard to iteratively design your Forms (screens, nodes, transitions) and Flows (actions, logic,
vault / connection usage). Iterate until the behavior is validated end‑to‑end. This keeps early experimentation
fast and visual.

## 2. Export Form JSON from Dashboard

From the Forms UI, export each Form. Each exported file contains:

- The `form` definition (nodes, start/ending, languages)
- All referenced `flows`
- All referenced `vault connections`

Placeholders (tokens) like `#FLOW-1#`, `#FLOW-2#`, `#CONN-1#`, etc. stand in for actual resource IDs.

~> Multiple exported Form files may repeat the same Flow or Connection definitions. This is expected; keep them
as-is. Avoid manual pruning—Terraform will only materialize what you explicitly declare as resources.

## 3. Store Exports in Module Directory

Create an `exported_forms` (or similar) folder co-located with your Terraform module and drop the raw JSON
files there.

Example:

```
actions/
├── actions.tf
├── exported_forms/
│   ├── verify_email.json
│   └── verify_phone.json
├── flows.tf
├── forms.tf
├── vaults.tf
└── ...
```

## 4. Inspect an Export (Annotated Example)

Below is an abbreviated export (comments added for illustration). Real exports are larger; keep them intact.

```json
{
	"version": "4.0.0",

  // Forms

	"form": {
		"name": "Verify Phone",
		"languages": { "primary": "en" },
		"nodes": [
			{ "id": "step_oEm5", "type": "STEP" },
			{
				"id": "flow_HtMY",
				"type": "FLOW",
				"config": { "flow_id": "#FLOW-1#", "next_node": "step_4lZR" }
			}
			// ...
		]
	},

  // Flows

	"flows": {
		"#FLOW-1#": {
			"name": "Send SMS OTP",
			"actions": [
				{
					"id": "send_sms_with_twilio",
						"type": "TWILIO",
						"action": "SEND_SMS",
						"params": {
							"connection_id": "#CONN-1#",
							"message": "Your verification code is {{actions.generate_otp.code}}."
						}
				}
			]
		}
	},

  // Vaults
  
	"connections": {
		"#CONN-1#": { "id": "ac_mRc2...", "app_id": "TWILIO", "name": "Twilio" }
	}
}
```

## 5. Decode JSON with Terraform Locals

Use `file()` + `jsondecode()` to hydrate only the fragments you need. Keep locals small and intent-revealing.

```hcl
locals {
	flow_send_email_otp_json   = jsondecode(file("${path.module}/exported_forms/verify_email.json"))["flows"]["#FLOW-1#"]
	flow_verify_email_otp_json = jsondecode(file("${path.module}/exported_forms/verify_email.json"))["flows"]["#FLOW-2#"]
	flow_send_sms_otp_json     = jsondecode(file("${path.module}/exported_forms/verify_phone.json"))["flows"]["#FLOW-1#"]
	flow_verify_sms_otp_json   = jsondecode(file("${path.module}/exported_forms/verify_phone.json"))["flows"]["#FLOW-2#"]
}
```

## 6. Materialize Flow Resources

Replace token placeholders with Terraform-managed resource IDs (for example, vault connection IDs). Nest
`replace()` calls if multiple tokens occur.

```hcl
resource "auth0_flow" "send_email_otp" {
	name    = "Send Email OTP"
	actions = replace(
		jsonencode(local.flow_send_email_otp_json["actions"]),
		"#CONN-1#", auth0_flow_vault_connection.sendgrid_vault.id
	)
}

resource "auth0_flow" "verify_email_otp" {
	name    = "Verify Email OTP"
	actions = replace(
		jsonencode(local.flow_verify_email_otp_json["actions"]),
		"#CONN-2#", auth0_flow_vault_connection.auth0_management_api_vault.id
	)
}
```

## 7. Materialize Form Resources

Similarly, hydrate Form primitives (`start`, `ending`, `nodes`). Perform placeholder substitution for Flow
references inside the `nodes` structure.

```hcl
locals {
	form_verify_email_json = jsondecode(file("${path.module}/exported_forms/verify_email.json"))["form"]
	form_verify_phone_json = jsondecode(file("${path.module}/exported_forms/verify_phone.json"))["form"]
}

resource "auth0_form" "verify_email" {
	name = "Verify Email"
	languages { primary = "en" }

	start  = jsonencode(local.form_verify_email_json["start"])
	ending = jsonencode(local.form_verify_email_json["ending"])
	nodes  = replace(
		replace(
			jsonencode(local.form_verify_email_json["nodes"]),
			"#FLOW-1#", auth0_flow.send_email_otp.id
		),
		"#FLOW-2#", auth0_flow.verify_email_otp.id
	)
}

resource "auth0_form" "verify_phone" {
	name = "Verify Phone"
	languages { primary = "en" }

	start  = jsonencode(local.form_verify_phone_json["start"])
	ending = jsonencode(local.form_verify_phone_json["ending"])
	nodes  = replace(
		replace(
			replace(
				jsonencode(local.form_verify_phone_json["nodes"]),
				"#FLOW-1#", auth0_flow.send_sms_otp.id
			),
			"#FLOW-2#", auth0_flow.verify_sms_otp.id
		),
		"#FLOW-3#", auth0_flow.create_or_replace_phone_mfa.id
	)
}
```

## 8. Update Workflow

1. Modify a Form or Flow in the Dashboard.
2. Re-export the Form JSON.
3. Replace the corresponding file under `exported_forms/`.
4. Run `terraform plan` to view drift (primarily token-resolved diffs).
5. Apply when satisfied.

Because Terraform only sees concise resource definitions (actions arrays, node lists), large structural edits
become low-friction: no manual reformatting, fewer merge conflicts, clearer reviews.

## 9. Tips & Considerations

- Duplication Trade-off: Prefer duplication in exports over premature normalization that obscures provenance.
- Secret Material: Exports won't include sensitive secret values beyond what the Dashboard reveals—manage secrets via dedicated resources or vault configuration.

## Alternative: Auth0 CLI Auto-generation (v1.6.1+)

For a simpler workflow focused on Forms / Flows, an alternative is to leverage the Auth0 CLI (version >= 1.6.1)
to generate Terraform configuration directly:

```sh
auth0 tf generate \
	--tf-version 1.28.0 \
	--output-dir tmp-auth0-tf \
	--resources auth0_flow,auth0_flow_vault_connection,auth0_form
```

This limits generation to the relevant resource types and can be useful for an initial bootstrap or periodic
regeneration snapshot.

Note: Replace exported IDs in the generated output with Terraform variables or interpolations where appropriate
to avoid hard-coding environment-specific identifiers.

## Summary

By externalizing the rich, nested structure of Forms and Flows into exported JSON snapshots, Terraform
configuration remains concise, reviewable, and easy to evolve. This pattern lowers friction for iterative
improvements while preserving Terraform as the source of truth for final deployed resource wiring.

