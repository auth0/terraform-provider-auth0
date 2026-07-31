---
branch: tfp-bot-detection-entitlement-dxcdt-1688
base: main
---

# Surface Bot Detection/Captcha entitlement errors as warnings

<!--
❗ For general support or usage questions, use the Auth0 Community forums or raise a support ticket.

By submitting a pull request to this repository, you agree to the terms within the Auth0 Code of Conduct: https://github.com/auth0/open-source-template/blob/master/CODE-OF-CONDUCT.md.
-->

### 🔧 Changes

`auth0_attack_protection` gates two sub-features behind tenant entitlements — `bot_detection` and `captcha`, both served by the v2 Management API. On a tenant without the add-on, both `GET` and `PATCH` return `403` with `errorCode: "insufficient_entitlement"`. Previously the resource treated *any* v2 `403` whose message contained `insufficient_scope` as a silent skip (the temporary workaround from #1410) and let everything else fail the apply. That meant an unentitled tenant either got a hard apply failure, or — where the string happened to match — a skip with nothing but an `[INFO]` log line the user never sees.

This PR distinguishes the two `403` cases and makes the entitlement case visible without making it fatal.

**Added — `internal/error/api_error_v2.go`**

Typed error helpers for the v2 SDK, replacing `strings.Contains` on the error message:

- `v2ForbiddenErrorCode(err) string` (unexported) — unwraps via `errors.As` to `*managementv2.ForbiddenError` and reads `errorCode` out of the response `Body`; returns `""` for any non-`ForbiddenError`, non-map body, missing key, or non-string value.
- `IsInsufficientScope(err) bool` — token lacks the required scope.
- `IsInsufficientEntitlement(err) bool` — tenant lacks the required add-on entitlement.

**Changed — `internal/auth0/attackprotection/resource.go`**

`readAttackProtection` and `updateAttackProtection` now accumulate a `diag.Diagnostics` and branch on error code for each of the two gated calls:

- `IsInsufficientScope` → skip, `[INFO]` log (unchanged behavior, now matched on the error code rather than the message text).
- `IsInsufficientEntitlement` → append a `diag.Warning` naming the sub-feature and pointing at Auth0 support to enable the feature.
- anything else → fatal, as before (`diag.FromErr` on read, `multierror.Append` on update).

The warning text is built by a single helper, `entitlementWarning(feature, consequence)`, parameterized on both the sub-feature name and the consequence, so the message matches what actually happened. The two consequences are named constants: `entitlementUpdateConsequence` ("the configuration was not applied") and `entitlementReadConsequence` ("its current configuration could not be read") — a read failure hasn't attempted to write anything, so claiming otherwise was misleading.

Three consequences worth noting for reviewers:

- **No short-circuiting.** A gated sub-feature returning `403` no longer stops the remaining sub-features from being read or updated, and two gated failures in the same apply both produce warnings rather than the first one masking the second.
- **Warnings propagate through update → read.** `updateAttackProtection` merges the diagnostics returned by its trailing `readAttackProtection` call instead of returning them directly, so warnings raised during update survive the read.
- **Warnings survive a later fatal error.** Every fatal exit in both functions returns `append(diags, diag.FromErr(err)...)` rather than `diag.FromErr(err)`, so a warning collected from a gated sub-feature is still shown when an unrelated sub-feature subsequently fails hard. Previously the accumulated warnings were dropped on those paths.

`strings` is no longer imported by the resource.

No schema, public API, or user-facing configuration changes — this is error-surfacing behavior only.

### 📚 References

- Supersedes the temporary `insufficient_scope` string-matching workaround introduced in [#1410](https://github.com/auth0/terraform-provider-auth0/pull/1410)

### 🔬 Testing

**Unit — `internal/error/api_error_v2_test.go` (new, 279 lines)**

Table of cases for `v2ForbiddenErrorCode`, `IsInsufficientScope`, and `IsInsufficientEntitlement`, covering the happy path plus every way extraction can fail: non-`ForbiddenError`, `Body` that isn't a map, missing `errorCode`, and non-string `errorCode`. Also covers `errors.As` unwrapping through a wrapped error, and cross-checks that each predicate returns `false` for the other's error code.

**Unit — `internal/auth0/attackprotection/resource_internal_test.go` (new)**

Table over both sub-features × both consequences, asserting `entitlementWarning` produces `diag.Warning` with a summary naming the sub-feature and the exact expected detail string. Plus two guards: that the read consequence does *not* claim the configuration was not applied while the update one does, and that neither echoes the backend's own 403 `message` — which has a copy-paste bug where the captcha endpoint says "…to use bot detection".

**Acceptance — `internal/auth0/attackprotection/resource_test.go` (3 new tests, HTTP recordings included)**

Recorded against a non-entitled tenant, replayed from cassettes:

- `TestAccAttackProtectionBotDetectionInsufficientEntitlement` — `bot_detection` + `brute_force_protection`; asserts the apply succeeds, `bot_detection.#` is `0` (the 403 leaves it unpersisted), and `brute_force_protection` is still fully applied.
- `TestAccAttackProtectionCaptchaInsufficientEntitlement` — the same for `captcha`.
- `TestAccAttackProtectionMultipleInsufficientEntitlements` — both gated blocks 403 in one apply; asserts neither short-circuits the other and the ungated `suspicious_ip_throttling` block still applies.

All three set `ExpectNonEmptyPlan: true`: a gated block cannot be persisted to state, so the config/state difference legitimately remains after apply.

Existing recordings for `TestAccAttackProtectionBotDetection` and `TestAccAttackProtectionCaptcha` were re-recorded against an entitled tenant to keep the happy paths green.

Run with:

```bash
make test-unit
AUTH0_HTTP_RECORDINGS=on AUTH0_DOMAIN=terraform-provider-auth0-dev.eu.auth0.com make test-acc FILTER="TestAccAttackProtection"
```

Not covered by automated tests: the exact warning text as rendered by the Terraform CLI, since the acceptance test framework asserts on state rather than on diagnostic output. Verified manually against an unentitled tenant — `terraform apply` completes with `Warning: Bot Detection entitlement not available` / `Warning: Captcha entitlement not available` and a non-zero-diff plan.

### 📝 Checklist

- [x] All new/changed/fixed functionality is covered by tests (or N/A)
- [x] I have added documentation for all new/changed functionality (or N/A) — CHANGELOG entry added; no schema changes, so no `docs/` regeneration needed

<!--
❗ All the above items are required. Pull requests with an incomplete or missing checklist will be closed.
-->
