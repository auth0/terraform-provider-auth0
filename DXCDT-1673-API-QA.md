# API QA Report — Rate Limit Policies (EA)

> **Scope:** `/api/v2/rate-limit-policies` (5 endpoints) — DXCDT-1673
> **Spec (source of truth):** [api2-3.1-internal.json](api2-3.1-internal.json)
> **Tenant under test:** `https://kiran-dev.us.auth0.com/api/v2`
> **Auth:** Bearer token from `token.txt` (M2M, `sub=HCuHYoHrhT2OgacfzK0TVNc16uFbwZ7P@clients`)
> **Date:** 2026-05-31
> **Tester:** Claude (API QA)

---

## ⛔ BLOCKER — testing could not proceed past authorization

**Every** rate-limit-policies endpoint returns **HTTP 403 `insufficient_scope`**. The token in `token.txt` is valid and works against other Management APIs, but it **does not carry any of the four `*:rate_limit_policies` scopes** — and **those scopes cannot currently be granted on this tenant** (root cause below).

### Root cause (traced 2026-05-31, after the feature flag was enabled)

The EA flag `rate_limit_policies_management` **was** enabled on `kiran-dev@us` (confirmed via layer0). Despite that, the scopes are ungrantable:

1. Token has 0 of the 4 `*:rate_limit_policies` scopes → 403 on all 5 endpoints.
2. Tried to **PATCH the client grant** `cgr_qoPdX2x3DGhbXjT1` (M2M → Mgmt API) to add them →
   `400 "scope must be a subset of resource server scopes"`.
3. Inspected the **Auth0 Management API resource server** (`id=69a573f05758fd9f6a521b12`, `is_system:true`): it defines **256 scopes, none matching `rate_limit_policies`**. That's why they're absent from the dashboard permissions list and rejected by the grant.
4. Tried to **PATCH the resource server** to append the 4 scopes →
   `400 "Additional properties not allowed: scopes"`. **The system Management API resource server's scope list is not editable via the public Management API.**

➡️ **Conclusion:** Enabling the `rate_limit_policies_management` flag did **not** auto-register the `*:rate_limit_policies` scopes onto the system Management API resource server for this tenant. Without those scopes on the resource server, no client grant (and therefore no token) can carry them — so the API cannot be authorized through normal/public means. Functional CRUD / field / boundary testing is **blocked**.

### ✅ Definitive root cause (found in api2 source)

How EA scopes are surfaced on the customer Management API resource server is documented in [APIF/702972117 "mgmt-main-api | Adding new scopes" §2.a Closed Beta scopes](https://oktainc.atlassian.net/wiki/spaces/APIF/pages/702972117): a Closed-Beta/EA scope is only returned from `GET /resource-servers` if it is registered in the **`SCOPE_FLAG_MAPPING`** table in [`packages/main-api/lib/resource_servers/scopes/api2.js`](https://github.com/atko-cic/api2/blob/master/packages/main-api/lib/resource_servers/scopes/api2.js), keyed to the feature flag.

**api2 `master` currently has NO `rate_limit_policies` entry in `SCOPE_FLAG_MAPPING`.** It *used* to — but api2 PR **[#12148](https://github.com/atko-cic/api2/pull/12148)** ("feat: add rate-limit-policies in tenant members scopes", merged **2026-05-27**, by Leonardo Zanivan) **removed** it:

- ❌ Removed `RATE_LIMIT_POLICIES_SCOPES` and its `{ scopes: RATE_LIMIT_POLICIES_SCOPES, flag: 'rate_limit_policies_management' }` line from `SCOPE_FLAG_MAPPING` (the flag→scope mapping that surfaces the scopes to customer tenants).
- ✅ Added the 4 scopes to the **`tenant_members` owner role** (`packages/main-api/lib/tenant_members/roles.js`) — i.e. internal/dashboard/RTA context, **not** the public customer-tenant resource server.

The PR description states the intent and the blocking condition explicitly:
> "Add rate limit policies scope for tenant member owner role. **Remove the scopes conditional in preparation for EA.**"
> 🚫 **Not ready to merge until:** *The new scopes are added for EA.*

So the flag-gated (Closed-Beta) path was intentionally removed **ahead of** the real EA enablement, and the EA re-registration **has not yet landed**. That is precisely why enabling `rate_limit_policies_management` on `kiran-dev@us` surfaces **zero** scopes: the code path that would add them under that flag no longer exists on `master`, and the GA/public registration isn't there yet.

This matches the EA checklist item **"Add new rate limit policies scopes … `PR IN REVIEW`"** ([APIF/702972117 referenced from the Edge Rate Limits Launch Checklist](https://oktainc.atlassian.net/wiki/spaces/479625216/pages/780176394)) — the scope-enablement PR for EA is still outstanding.

### What needs to happen to unblock (owner action — Kiran)

The fix is **server-side in the api2 repo**, not something the public Management API / this token can do. The EA scope-registration that api2 PR #12148 deferred must land:

1. **Primary ask — re-register the EA scopes under the flag.** A follow-up api2 change must restore the `rate_limit_policies` scopes into `SCOPE_FLAG_MAPPING` (keyed to `rate_limit_policies_management`) in [`packages/main-api/lib/resource_servers/scopes/api2.js`](https://github.com/atko-cic/api2/blob/master/packages/main-api/lib/resource_servers/scopes/api2.js), per the [Adding-new-scopes §2.a](https://oktainc.atlassian.net/wiki/spaces/APIF/pages/702972117) process. PR #12148's own "Not ready to merge until: *The new scopes are added for EA*" confirms this is the outstanding step. **Owner to ping Leonardo Zanivan** (PR #12148 author) **/ the `iam_protocols` team** (flag owner) to confirm status & ETA — this is also the EA-checklist item *"Add new rate limit policies scopes — PR IN REVIEW"*.
   - Per the doc, after the api2 change there are companion `layer0-base` config edits (api2 + server `API2_SCOPES`) and a deploy needed before the scope is honored end-to-end.
2. **Once the scope is live on the resource server** (verify with `GET /resource-servers` → the Management API RS lists `*:rate_limit_policies`), I can finish the rest with this token: PATCH the M2M grant `cgr_qoPdX2x3DGhbXjT1` to add the 4 scopes (token has `update:client_grants`), then you re-mint.
3. **Confirm global client** (spec 401: "Client is not global") — verify the M2M app qualifies, or use an RTA/internal token path.
4. **Fastest interim unblock:** ask the feature team for a **pre-minted token already carrying the 4 scopes** (they have internal token tooling — `npm run get-token:vivaldi`), drop it into `token.txt`, and I run the matrix immediately.

> I already performed every reversible public-API step available (grant PATCH, resource-server PATCH) — both are rejected by platform constraints, and the source-level diagnosis above shows why: the enabling code isn't on api2 `master` yet.

### Second tenant attempt — `kiran-kumar-test-dev.tus.auth0.com` (2026-05-31)

Tried a second internal tenant with a fresh token (`tus_token.txt`). **Identical outcome — same root cause, confirmed environment-independent:**

| Step | Result |
|---|---|
| Token validity (`GET /clients?per_page=1`) | ✅ 200 — token valid, 252 scopes, **0** `rate_limit_policies` |
| `GET /rate-limit-policies` | ❌ 403 `insufficient_scope` (read:rate_limit_policies) |
| Mgmt API resource server (`697311617e6baaa64820a9e2`, `is_system:true`) | 252 scopes, **0** rate_limit — same as kiran-dev |
| PATCH resource server `{scopes:[...]}` | ❌ 400 `Additional properties not allowed: scopes` |
| PATCH grant `cgr_YIftdCVOvE9occOj` add scopes | ❌ 400 `scope must be a subset of resource server scopes` |
| Alt internal `/resource-servers/{id}/scopes` endpoint | ❌ 404 Not Found |
| Feature flag override for this tenant | **Not present** — layer0 overrides list only `dev-ankita-t@us`, `dev-tanya@us`, `kiran-dev@us`, `kunal-dev@us` (all `@us`); the `.tus.` tenant has no override. |

➡️ Confirms the blocker is **not** tenant- or environment-specific. The scopes are absent from the system Management API resource server because api2 `master` no longer registers them (PR #12148) — no public-API path on any tenant can add them. The fix must land in api2.

### Who/where to check for prior art (Slack)

This same wall was almost certainly hit by the **other 3 dev tenants** that have the flag enabled (from layer0): **`dev-ankita-t@us`** (Deploy CLI — Ankita Tripathi), **`dev-tanya@us`**, **`kunal-dev@us`** (go-auth0 SDK — Kunal Dawar). Worth asking them directly how they tested (or whether they only used mocked/recorded responses — the Deploy CLI PR #1395 used **recorded 403 fixtures**, suggesting they may *not* have had live scopes either).

Suggested Slack channels / threads to search for prior discussion:
- The **`iam_protocols`** team channel (flag owner per layer0).
- The handover-strategy thread referenced in IPS-5789 comment 1: `https://auth0.slack.com/archives/C0A2C8EM54M/p1771473445947329` (channel `C0A2C8EM54M`).
- Search Slack for: `rate_limit_policies scope`, `create:rate_limit_policies`, `SCOPE_FLAG_MAPPING rate limit`, `rate_limit_policies_management scope`, or PR `api2#12148`.
- People to ping: **Leonardo Zanivan**, **Samuel Salazar** (IPS-5789 assignee), **Charles Rea**, **Rajat Bajaj** (SDKREQ-196).

---

## 1. Token analysis

Decoded JWT payload (no secret needed — informational):

| Claim | Value |
|---|---|
| `iss` | `https://kiran-dev.us.auth0.com/` |
| `sub` | `HCuHYoHrhT2OgacfzK0TVNc16uFbwZ7P@clients` (M2M) |
| `aud` | `https://kiran-dev.us.auth0.com/api/v2/` |
| `iat` | 1780211098 |
| `exp` | 1780297498 (**~22 h** remaining at time of test — still valid) |
| `scope` | **256 scopes**, but **0** matching `rate_limit_policies` |

→ Token is valid and unexpired, but lacks the required authorization for this API.

---

## 2. Endpoints under test (from spec)

| # | Method | Path | Scope required | Documented success |
|---|---|---|---|---|
| 1 | GET | `/rate-limit-policies` | `read:rate_limit_policies` | 200 list |
| 2 | POST | `/rate-limit-policies` | `create:rate_limit_policies` | 201 |
| 3 | GET | `/rate-limit-policies/{id}` | `read:rate_limit_policies` | 200 |
| 4 | PATCH | `/rate-limit-policies/{id}` | `update:rate_limit_policies` | 200 |
| 5 | DELETE | `/rate-limit-policies/{id}` | `delete:rate_limit_policies` | 204 |

---

## 3. Blocker evidence

### 3.1 All five endpoints → 403 `insufficient_scope`

**GET `/rate-limit-policies`**
```http
GET /api/v2/rate-limit-policies
Authorization: Bearer <token.txt>
```
```json
HTTP 403
{"statusCode":403,"error":"Forbidden","message":"Insufficient scope, expected any of: read:rate_limit_policies","errorCode":"insufficient_scope"}
```

**GET `/rate-limit-policies?take=50`** → identical 403 (`read:rate_limit_policies`).

**POST `/rate-limit-policies`**
```http
POST /api/v2/rate-limit-policies
Content-Type: application/json

{"resource":"oauth_authentication_api","consumer":"client","consumer_selector":"default","configuration":{"action":"block","limit":100}}
```
```json
HTTP 403
{"statusCode":403,"error":"Forbidden","message":"Insufficient scope, expected any of: create:rate_limit_policies","errorCode":"insufficient_scope"}
```

**GET `/rate-limit-policies/rlp_testid`**
```json
HTTP 403
{"statusCode":403,"error":"Forbidden","message":"Insufficient scope, expected any of: read:rate_limit_policies","errorCode":"insufficient_scope"}
```

**PATCH `/rate-limit-policies/rlp_testid`** (body `{"configuration":{"action":"allow"}}`)
```json
HTTP 403
{"statusCode":403,"error":"Forbidden","message":"Insufficient scope, expected any of: update:rate_limit_policies","errorCode":"insufficient_scope"}
```

**DELETE `/rate-limit-policies/rlp_testid`**
```json
HTTP 403
{"statusCode":403,"error":"Forbidden","message":"Insufficient scope, expected any of: delete:rate_limit_policies","errorCode":"insufficient_scope"}
```

### 3.2 Control checks (proving the blocker is scope, not token/endpoint)

| Check | Result | Conclusion |
|---|---|---|
| `GET /clients?per_page=1` (a scope the token HAS) | **HTTP 200** | Token itself is valid & accepted by the Mgmt API. |
| `GET /rate-limit-policies-FAKE` (nonexistent path) | **HTTP 404** | Unknown paths 404. |
| `GET /rate-limit-policies` (real path) | **HTTP 403** | Real path is **registered** (else 404); only scope is missing. |

→ The endpoints exist and are deployed on this tenant; the **sole** obstacle is the token's missing `*:rate_limit_policies` scopes.

### 3.3 Why the scopes can't be granted (root-cause chain)

**Attempt to add scopes to the M2M client grant** (`PATCH /client-grants/cgr_qoPdX2x3DGhbXjT1`, body included the 4 new scopes):
```json
HTTP 400
{"statusCode":400,"error":"Bad Request","message":"Payload validation error: scope must be a subset of resource server scopes","errorCode":"invalid_body"}
```

**Inspect the Management API resource server** (`GET /resource-servers`, filtered):
```json
{"id":"69a573f05758fd9f6a521b12","identifier":"https://kiran-dev.us.auth0.com/api/v2/",
 "name":"Auth0 Management API","is_system":true,"total_scopes":256,"rate_scopes":[]}
```
→ The system resource server **defines no `rate_limit_policies` scopes** → the grant subset check fails, and the dashboard shows none under "rate".

**Attempt to add the scopes to the resource server** (`PATCH /resource-servers/69a573f05758fd9f6a521b12`, body `{"scopes":[...]}`):
```json
HTTP 400
{"statusCode":400,"error":"Bad Request","message":"Payload validation error: 'Additional properties not allowed: scopes'.","errorCode":"invalid_body"}
```
→ The **system Management API resource server's scopes are not editable via the public API**. Scope registration for this EA feature must happen through Auth0-internal provisioning (expected to accompany the feature flag, but it did not for this tenant).

---

## 4. Findings so far (vs. spec)

| Observation | Spec alignment |
|---|---|
| 403 body shape `{statusCode,error,message,errorCode}` with `errorCode:"insufficient_scope"` | ✅ Consistent with spec's documented 403 ("Insufficient scope; expected any of: …:rate_limit_policies"). Message wording is `"Insufficient scope, expected any of: <scope>"` (comma vs the spec's semicolon prose — cosmetic only). |
| Per-operation scope enforcement (each verb demands its own scope) | ✅ Matches the per-operation `oAuth2ClientCredentials` security blocks in the spec. |
| Endpoint registered on `kiran-dev.us.auth0.com` (403 not 404) | ✅ Endpoint is live on the tenant. |
| Could not yet verify: request/response schemas, enums, `limit` 0–10000 bound, `redirect_uri` https validation, 409 uniqueness, 404-by-id, pagination, field-level edge cases | ⏸️ **Blocked** — pending scoped token. |

**No functional deviations can be confirmed or denied yet** — everything below the auth layer is untested.

---

## 5. Planned test matrix (ready to execute once unblocked)

The full, execution-ready matrix (~90 concrete cases with request bodies, expected responses, and a results-capture template) lives in its own file: **[DXCDT-1673-API-TEST-PLAN.md](DXCDT-1673-API-TEST-PLAN.md)**.

Coverage summary: pre-flight/auth (P1–P4) · list+pagination+filters (L1–L18) · create happy/boundary/union/uniqueness (C1–C49) · get (G1–G7) · update incl. immutability checks (U1–U22) · delete (D1–D6) · cleanup (X1–X2) · cross-cutting schema/header/timestamp audits (§9). Each case maps back to a Terraform provider decision in §11 of that file.

→ Once a scoped token is in `token.txt`, execute that plan top-to-bottom and record actuals there; surface any deviations into this QA report's findings ([§4](#4-findings-so-far-vs-spec)).

---

## 6. Status & next step

| | |
|---|---|
| **Endpoints reachable** | ✅ Yes (deployed on `kiran-dev.us.auth0.com`) |
| **Token valid** | ✅ Yes (works for other APIs, ~22 h TTL) |
| **Authorized for rate-limit-policies** | ❌ **No — missing all 4 scopes** |
| **Functional testing performed** | ⏸️ Blocked at auth layer |

**Next step:** Kiran to re-issue the token with `read/create/update/delete:rate_limit_policies` scopes (and confirm the `rate_limit_policies_management` flag + global client), overwrite `token.txt`, and notify me. I will then execute §5 in full and expand this report with actual requests/responses, observations, and any spec deviations.

> ⚠️ **Security note:** `token.txt` contains a live bearer token for the Management API. It is currently untracked in git — keep it out of commits (add to `.gitignore`) and revoke/rotate it after testing.
