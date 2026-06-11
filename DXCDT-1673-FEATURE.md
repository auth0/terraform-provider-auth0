# Feature Spec — Terraform Provider support for 3rd Party / Application Rate Limits (Authentication API, EA)

> **Audience:** CDT (Customer Developer Tooling) team
> **Primary ticket:** [DXCDT-1673](https://auth0team.atlassian.net/browse/DXCDT-1673) — *Terraform-provider support for 3rd Party / Application Rate Limits for the Authentication API (EA)*
> **Status (as of 2026-05-31):** In Progress · Assignee: Kiran Kumar
> **Repo:** `auth0/terraform-provider-auth0`

---

## 1. Ticket & Document Map

This feature spans several pillars. The Jira/Confluence hierarchy I traversed:

| Level | Ticket / Page | Summary | Status |
|---|---|---|---|
| Product epic (grandparent) | [ROAD-6548](https://auth0team.atlassian.net/browse/ROAD-6548) | Custom Rate Limits for the Authentication API (EA) | Release Readiness · **DPR 🔴 Red** |
| CDT parent | [DXCDT-1672](https://auth0team.atlassian.net/browse/DXCDT-1672) | CDT support for 3rd Party / Application Rate Limits | In Progress |
| **This ticket** | **[DXCDT-1673](https://auth0team.atlassian.net/browse/DXCDT-1673)** | **Terraform-provider support** | **In Progress** |
| Sibling | [DXCDT-1674](https://auth0team.atlassian.net/browse/DXCDT-1674) | Deploy CLI support | Internal Review |
| SDK/enablement | [SDKREQ-196](https://auth0team.atlassian.net/browse/SDKREQ-196) | New Mgmt API endpoint + CDT tooling changes | In Progress |
| Product epic parent | [CIC-29](https://auth0team.atlassian.net/browse/CIC-29) | Protocols, Clients & API Management Improvements | — |

**Related product/infra tickets (from ROAD-6548):** AOB-1622, AOB-1635 (rate-limit policy mgmt/enforcement), PDOPPS-6913 (per-client-ID limiting), PDOPPS-4169 (RPS per tenant), ROAD-2371 / ROAD-8534 (**Replace ingress-nginx — blocker**), ROAD-6991 (Partner Portal EA), ROAD-5837 (3rd Party Apps GA).

**Confluence:**
- [Rate Limit Policies EA Checklist](https://oktainc.atlassian.net/wiki/spaces/PSS/pages/796956029) (PSS)
- [Edge Rate Limits Launch Checklist](https://oktainc.atlassian.net/wiki/spaces/479625216/pages/780176394) — master EA/GA task tracker
- [SDK & CDT Enablement Process](https://oktainc.atlassian.net/wiki/spaces/DXSDK/pages/673551497) — the process we follow

**Key engineering artifacts:**
- OAS PR (merged 2026-05-21): [atko-cic/api2#12200](https://github.com/atko-cic/api2/pull/12200)
- **Source of truth for API structure & params:** [api2-3.1-internal.json](api2-3.1-internal.json) (OpenAPI 3.1, internal spec) — all field constraints below are taken verbatim from it.
- Server feature flag: **`rate_limit_policies_management`**
- New Mgmt API base path: **`/api/v2/rate-limit-policies`**
- Release lifecycle in spec: **`x-release-lifecycle: beta`**, **`x-internal: true`** (confirms EA/internal-only status)
- OAuth2 scopes: `create:` / `read:` / `update:` / `delete:rate_limit_policies`
- **Requires a *global* client** (spec 401 note: "Client is not global").

> ⚠️ **The product itself is DPR-Red.** ROAD-6548's May-31 launch is blocked by the ingress-nginx → Istio infra migration (tracking ~end-of-June). CDT work can proceed in parallel, but **GA/public exposure is gated on the infra rollout and the EA feature flag.**

---

## 2. What — The Feature

Add a new Terraform resource (and likely a data source) to the Auth0 provider that lets administrators manage **Rate Limit Policies** for the Authentication API as code.

A Rate Limit Policy expresses **application-level (client-ID-based) throttling** for the Authentication API, covering both interactive user flows (Authorization Code, Refresh Token) and M2M `client_credentials` flows.

Proposed resource: **`auth0_rate_limit_policy`** (full CRUD), mapping 1:1 to the `/api/v2/rate-limit-policies` endpoints.

---

## 3. Why — Problem & Business Context

From ROAD-6548:

> The central problem is the **"noisy neighbor"** risk, where a single application's traffic spike on the Authentication API — impacting both M2M (`client_credentials`) and interactive user flows — exhausts a customer's **global** rate limit, causing a catastrophic, tenant-wide outage that blocks all other applications and legitimate logins.

**The approach: Client-ID Based Throttling.** Administrators define per-application rate-limit policies. When an application exceeds its unique limit it is throttled **independently at the edge**, and that blocked traffic **does not count against the tenant's global rate limit** — isolating the misbehaving app and protecting everyone else.

**Strategic driver:** makes it safe for customers to build/scale ecosystems of third-party applications (API economy / enterprise strategy). Auth0 priority theme: *Lead AI Identity*.

Terraform support specifically lets reliability engineers and SaaS product owners manage these isolation policies declaratively as part of their IaC, instead of manual dashboard/API intervention.

---

## 4. When — Timeline & Sequencing

| Milestone | Date / State |
|---|---|
| OAS PR merged & published | 2026-05-21 ✅ |
| Go & Node SDKs generated (SDKREQ-196) | In progress (picked up sprint starting ~week of May 25) |
| CDT work started (auto-flag on SDKREQ-196) | 2026-05-28 ✅ |
| Original product launch target | 2026-05-31 ❌ (blocked) |
| Infra (ingress-nginx → Istio) rollout | ~end of June 2026 |

**Dependency chain (from the Launch Checklist):**
`OpenAPI public schema → Go & Node SDKs → Deploy CLI, Terraform & Auth0 CLI`

Good news for us: **the Go SDK dependency is already satisfied** — see §5.

---

## 5. How — Technical Implementation

### 5.1 Endpoints (from `api2-3.1-internal.json`)

| Method | Path | Operation | Request | Success | Scope |
|---|---|---|---|---|---|
| GET | `/rate-limit-policies` | list | query params (below) | 200 `ListRateLimitPoliciesPaginatedResponseContent` | `read:rate_limit_policies` |
| POST | `/rate-limit-policies` | create | `CreateRateLimitPolicyRequestContent` | **201** `CreateRateLimitPolicyResponseContent` | `create:rate_limit_policies` |
| GET | `/rate-limit-policies/{id}` | get | path `id` | 200 `GetRateLimitPolicyResponseContent` | `read:rate_limit_policies` |
| PATCH | `/rate-limit-policies/{id}` | update | `PatchRateLimitPolicyRequestContent` | 200 `UpdateRateLimitPolicyResponseContent` | `update:rate_limit_policies` |
| DELETE | `/rate-limit-policies/{id}` | delete | path `id` | **204** (no body) | `delete:rate_limit_policies` |

**List query params:** `resource` (enum), `consumer` (enum), `consumer_selector` (string ≤255), `take` (int 1–100, default 50), `from` (cursor ≤1000). Response: `{ rate_limit_policies: [...], next: <cursor> }`.

**Error responses to handle:**
- **409** on POST — *"A rate limit policy with the same `resource`, `consumer`, and `consumer_selector` already exists."* → confirms the uniqueness tuple (see schema note below); surface a clear error.
- **404** on GET/PATCH/DELETE by id — policy doesn't exist → `internalError.HandleAPIError` should drop it from state on read.
- **400** invalid body, **401** invalid/non-global client token, **403** insufficient scope, **429** rate limited.

### 5.2 SDK readiness ✅ (already in this repo)

`go-auth0/v2 v2.12.0` is **already in [go.mod](go.mod)** and ships the generated client matching the spec above. No SDK bump required to start.

- Subclient: `management/client.Management.RateLimitPolicies` (`*ratelimitpolicies.Client`)
- Access in provider code: `meta.(*config.Config).GetAPIV2().RateLimitPolicies`
- Operations:
  - `List(ctx, *ListRateLimitPoliciesRequestParameters, opts...)` → cursor pager
  - `Create(ctx, *CreateRateLimitPolicyRequestContent, opts...)` → `*CreateRateLimitPolicyResponseContent`
  - `Get(ctx, id, opts...)` → `*GetRateLimitPolicyResponseContent`
  - `Update(ctx, id, *PatchRateLimitPolicyRequestContent, opts...)` → `*UpdateRateLimitPolicyResponseContent`
  - `Delete(ctx, id, opts...)` → `error`

### 5.3 Data model (authoritative — from the OpenAPI spec)

`RateLimitPolicy` (= Create/Get/Update response shape; all of `id`,`resource`,`consumer`,`consumer_selector`,`configuration` are **required**):

| Field | Type | Constraints (from spec) | TF mapping |
|---|---|---|---|
| `id` | string | `maxLength: 26`, format `rate-limit-policy-id` | Computed |
| `resource` | enum | **only** `oauth_authentication_api` | Required, `ForceNew` |
| `consumer` | enum | **only** `client` | Required, `ForceNew` |
| `consumer_selector` | string | `maxLength: 255` | Required, `ForceNew` |
| `configuration` | oneOf (3 variants) | required | Required nested block |
| `created_at` | string | format `date-time` | Computed |
| `updated_at` | string | format `date-time` | Computed |

**`consumer_selector` supported values** (≤255 chars):
`client_id:<client_id>` · `client_id:<cimd_uri>` · `cimd_clients` · `third_party_clients` · `default`.

**`configuration` is a `oneOf` discriminated by `action`:**

| Variant | `action` enum | `limit` | `redirect_uri` |
|---|---|---|---|
| 1 | `allow` | — (not allowed) | — |
| 2 | `block` \| `log` | **required**, int `0–10000` | — |
| 3 | `redirect` | **required**, int `0–10000` | **required**, string format `strict-https-uri` |

> **Important nuances the SDK types alone don't show (taken from the spec):**
> - Each variant is `additionalProperties: false` — sending `limit` with `allow`, or `redirect_uri` without `redirect`, is invalid. Expand must emit **only** the fields valid for the chosen `action`.
> - `limit` bounds are **0–10000** → add `validation.IntBetween(0, 10000)`.
> - `redirect_uri` is `strict-https-uri` → validate `https://` scheme client-side.
> - **PATCH only accepts `configuration`** (`PatchRateLimitPolicyRequestContent` has that single required field). `resource`/`consumer`/`consumer_selector` are **not updatable** → they must be `ForceNew: true`. *(Resolves OQ#3.)*
> - The Patch config union (`PatchRateLimitPolicyConfigurationRequestContent`) mirrors the create union, so expand logic can be shared.

> **Implementation note:** `RateLimitPolicyConfiguration` is a polymorphic type in the SDK (`…Zero`=allow / `…One`=block,log / `…Action`=redirect members + a `Visitor`). Model it in Terraform as a single nested `configuration` block with `action` (validated enum), optional `limit`, optional `redirect_uri`, plus a `CustomizeDiff`/validation enforcing the per-action field requirements. Branch on `action` in both `expand.go` and `flatten.go`.

### 5.4 Suggested resource shape (`auth0_rate_limit_policy`)

```hcl
resource "auth0_rate_limit_policy" "noisy_app" {
  resource          = "oauth_authentication_api"  # ForceNew (only valid value)
  consumer          = "client"                    # ForceNew (only valid value)
  consumer_selector = "client_id:abc123"          # ForceNew, <=255 chars

  configuration {
    action       = "redirect"      # allow | block | log | redirect
    limit        = 1000            # required for block/log/redirect; 0-10000; omit for allow
    redirect_uri = "https://..."   # required for redirect only; must be https
  }
}
```

### 5.5 Code layout & template to follow

Mirror an existing **v2-SDK** resource. Closest small template: [internal/auth0/supplementalsignals/resource.go](internal/auth0/supplementalsignals/resource.go) (uses `GetAPIV2()`, `expand.go`/`flatten.go` split). For full CRUD + import, also reference [internal/auth0/client/resource_cimd.go](internal/auth0/client/resource_cimd.go).

Create a new package `internal/auth0/ratelimitpolicy/`:

| File | Responsibility |
|---|---|
| `resource.go` | `NewResource()` — schema + Create/Read/Update/Delete/Import contexts. All descriptions tagged **`(EA only)`** (§6.1). |
| `data_source.go` | `NewDataSource()` — **singular** `auth0_rate_limit_policy`, lookup by `id`, calls `Get`, returns one policy. Reuse the resource schema via `internalSchema.TransformResourceToDataSource(NewResource().Schema)`. See §5.5.1 for why two data sources. |
| `data_source_rate_limit_policies.go` | `NewRateLimitPoliciesDataSource()` — **plural** `auth0_rate_limit_policies`, optional filter args (`resource`/`consumer`/`consumer_selector`), paginates `List`, exposes a computed `rate_limit_policies` `TypeList`. See §5.5.1. |
| `expand.go` | Terraform state → `Create…RequestContent` / `Patch…RequestContent` (handle the config union — emit only fields valid for the chosen `action`). |
| `flatten.go` | API response → Terraform state (single policy + list flatten helpers). |
| `resource_test.go` / `data_source_test.go` / `data_source_rate_limit_policies_test.go` | Acceptance tests (HTTP recordings under `test/data/recordings/`). |

Then register in [internal/provider/provider.go](internal/provider/provider.go):
- `ResourcesMap`: `"auth0_rate_limit_policy": ratelimitpolicy.NewResource()`
- `DataSourcesMap`:
  - `"auth0_rate_limit_policy": ratelimitpolicy.NewDataSource()` (singular, by `id`)
  - `"auth0_rate_limit_policies": ratelimitpolicy.NewRateLimitPoliciesDataSource()` (plural, filtered list)

Use the config mutex (`meta.(*config.Config).GetMutex()`) and `internalError.HandleAPIError` for 404/state-removal, consistent with existing resources. **No EA feature-flag/403 handling** — the tenant flag will be enabled for tests (Decision #2); server enforces EA access.

### 5.5.1 Why two data sources (singular + plural) — project convention

The API exposes **both** a Get-by-id (`GET /rate-limit-policies/{id}`) and a filtered List (`GET /rate-limit-policies`). This project does **not** mechanically pair every Get with a List data source — it does so only when the List endpoint offers **filtering/search beyond a single-record lookup**. The rate-limit-policies List endpoint does exactly that (filters on `resource`/`consumer`/`consumer_selector` + cursor pagination), so the established **two-data-source pattern** applies:

| Established precedent (this repo) | Singular (Get-by-id) | Plural (List + filter) |
|---|---|---|
| Clients | `auth0_client` | `auth0_clients` |
| Custom domains | `auth0_custom_domain` | `auth0_custom_domains` |
| Client grants | `auth0_client_grant` | `auth0_client_grants` |
| **Rate limit policies (this work)** | **`auth0_rate_limit_policy`** | **`auth0_rate_limit_policies`** |

- **Singular** — pattern from [internal/auth0/customdomain/data_source.go](internal/auth0/customdomain/data_source.go) and [internal/auth0/client/data_source.go](internal/auth0/client/data_source.go): keyed by a lookup arg (`id`), calls `Get`, flattens one object built from `TransformResourceToDataSource(NewResource().Schema)`.
- **Plural** — pattern from [internal/auth0/customdomain/data_source_custom_domains.go](internal/auth0/customdomain/data_source_custom_domains.go): takes optional filter args, runs the `Take(100)` + `From(next)` pagination loop, and exposes a computed `TypeList` whose `Elem` reuses the resource schema; `SetId` is a fixed/synthetic value (e.g. `"rate-limit-policies"`).

> **Note on the alternative single-data-source pattern (Pattern B).** Some resources (e.g. the singleton-ish [custom_domain](internal/auth0/customdomain/data_source.go#L34-L63)) fold both behaviors into *one* data source: an optional `id` does a Get, and absence of `id` falls back to List-then-resolve (errors on 0 or >1). We deliberately **do not** use Pattern B here — rate-limit policies are genuinely multi-instance with meaningful filters, so the separate plural `auth0_rate_limit_policies` (returning the full filtered list, not just resolving to one) is the better fit and matches the `clients`/`custom_domains`/`client_grants` precedent above.

Both data sources are tagged **`(EA only)`** in their descriptions (§6.1).

### 5.6 Docs, examples, changelog (required for merge)

- `docs/resources/rate_limit_policy.md` — generated via `make docs` (uses `tfplugindocs`). Do **not** hand-author.
- `examples/resources/auth0_rate_limit_policy/` — `resource.tf` + `import.sh`.
- [CHANGELOG.md](CHANGELOG.md) — add a `FEATURES` entry referencing the PR.
- Mark the resource description **"(EA only)"**, consistent with other EA features in the provider.

---

## 5.7 Sibling reference — Deploy CLI PR (DXCDT-1674)

The Deploy CLI implementation ([auth0/auth0-deploy-cli#1395](https://github.com/auth0/auth0-deploy-cli/pull/1395), by Ankita Tripathi, **OPEN** as of 2026-05-29; +952/-2, 18 files) is the most useful precedent — same endpoint, same SDK contract, a sibling CDT tool. Key decisions we should mirror for **consistency across tooling**:

**SDK & types**
- Bumps `auth0` (node-auth0) to `^5.11.0`; types align with node-auth0 PR [#1348](https://github.com/auth0/node-auth0/pull/1348). (Our Go equivalent is already satisfied by `go-auth0/v2 v2.12.0`.)
- Models `configuration` as a **discriminated union** exactly matching the spec: `{action:'allow'}` | `{action:'block'|'log', limit}` | `{action:'redirect', limit, redirect_uri}`, each `additionalProperties:false`. **Note:** the Deploy CLI JSON schema does *not* encode the `0–10000` limit bound — our provider should add `IntBetween(0,10000)` to be stricter (server enforces it regardless).

**Identity & lifecycle (most important for us)**
- **Config-as-code identity is `consumer_selector`**, not `id`. Directory mode writes one JSON file per policy named after `sanitize(consumer_selector)`; the handler's `identifiers: ['id', 'consumer_selector']`.
- **`id`, `created_at`, `updated_at` are stripped on export** (treated as server-computed) → in Terraform terms these are `Computed`.
- **Update sends `configuration` only** — `stripUpdateFields: ['id','resource','consumer','consumer_selector','created_at','updated_at']`. This independently confirms our **`ForceNew` decision** on `resource`/`consumer`/`consumer_selector` (the Deploy CLI literally cannot update them via PATCH).
- List uses **checkpoint pagination** (`paginate(..., { checkpoint: true })`).

**EA / feature-gating behavior**
- `getType()` swallows **403/404/501** and returns `null` ("feature not enabled for this tenant") rather than erroring — lets the tool run cleanly on tenants without the flag.
- E2E test recordings stub `GET /api/v2/rate-limit-policies?take=50` → **403** `insufficient_scope` ("Rate Limit Policies feature is not enabled for this tenant"). **We can reuse this exact gating pattern** for our acceptance-test strategy (OQ#2): treat 403/404 on read as "not enabled / absent" and skip.
- Delete is guarded by `AUTH0_ALLOW_DELETE` (warns instead of deleting when false) — Terraform handles destroy natively, so no analog needed, but it confirms delete semantics (hard delete by `id`, 204).

**Implications for the Terraform resource**
1. ✅ Reinforces `resource`/`consumer`/`consumer_selector` = `Required` + `ForceNew`; `configuration` = updatable; `id`/timestamps = `Computed`.
2. **Coordinate naming with Ankita** so the Terraform attribute names match the Deploy CLI JSON/YAML keys 1:1 (they already match the API: `resource`, `consumer`, `consumer_selector`, `configuration.{action,limit,redirect_uri}`). No drift expected — keep it that way.
3. ⚠️ **Do NOT copy** the Deploy CLI's 403/404/501-as-"not-enabled" swallow — per Decision #2 the test tenant will have the flag enabled, so the resource should behave normally and use plain `internalError.HandleAPIError` (404 → remove from state). The Deploy CLI swallows it only because it scans all asset types regardless of tenant entitlement.
4. Their PR ships **no docs** (checklist unchecked) — our provider **must** generate `docs/` + examples (a gap we should not copy).

---

## 6. Decisions (resolved)

Resolved by the OpenAPI spec (`api2-3.1-internal.json`), the Deploy CLI sibling PR, and **direct answers from the ticket owner (Kiran Kumar, 2026-05-31):**

**From the spec / sibling PR:**
- ✅ **Multi-instance, not singleton** — policies are keyed by `id` and list-able (unlike `supplemental_signals`).
- ✅ **Uniqueness tuple** — POST returns **409** when `(resource, consumer, consumer_selector)` already exists. Surface as a clear error; no client-side dedup.
- ✅ **Immutability** — PATCH accepts **only** `configuration` → `resource`/`consumer`/`consumer_selector` are `ForceNew: true`.
- ✅ **`redirect_uri`** — format `strict-https-uri`; enforce `https://` client-side.
- ✅ **`limit` bounds** — `0–10000`.

**From the owner:**
1. ✅ **Data source IS in scope — ship TWO.** Following the project's singular+plural convention (`auth0_client`/`auth0_clients`, `auth0_custom_domain`/`auth0_custom_domains`), ship a **singular** `auth0_rate_limit_policy` (Get by `id`) **and** a **plural** `auth0_rate_limit_policies` (filtered List on `resource`/`consumer`/`consumer_selector` with pagination). The List endpoint offers real filtering/search, which is exactly when this repo creates a dedicated plural data source. *(Full rationale + precedents in §5.5.1; file layout in §5.5.)*
2. ✅ **No "feature-not-enabled" handling needed.** Owner will enable the `rate_limit_policies_management` flag on the test tenant. **Do not** add Deploy-CLI-style 403/404-as-disabled logic; use standard `internalError.HandleAPIError` (404→remove from state) only. *(Supersedes the earlier suggestion to mirror the Deploy CLI 403 swallow.)*
3. ⏳ **Add `ValidateFunc`s — but AFTER live API testing.** First exercise the real API (done next, by me) to confirm exact accepted values/bounds, then add validation for the enums (`resource`, `consumer`, `action`), `limit` (`IntBetween(0,10000)`), and `redirect_uri` (https). Sequence: implement → API-test → harden validation.
4. ✅ **Test setup owned by Kiran** — global client / tenant config will be arranged; no action needed from the implementation side beyond standard acceptance-test wiring.
5. ✅ **Ship as EA.** Follow the project's established EA convention — see §6.1.
6. ✅ **Proceed to release** with the EA changes; make whatever EA-specific adjustments are required and continue.

### 6.1 How EA resources are added in this project (convention)

From project history (e.g. PRs #1561 Google Workspace synced groups, #1545 ACR, #1494 password auth methods, #1452 `allow_all_scopes`, #1239 M2M usage limit, #1197 Tenant ACL), the EA convention is **documentation-only — there is NO code-level feature gating**:

- Append **`(EA only)`** to the resource's `Description` and to the `Description` of EA-specific attributes. Example, [internal/auth0/connection/resource_directory_synchronized_groups.go](internal/auth0/connection/resource_directory_synchronized_groups.go):
  > `"…synchronized via directory provisioning for an Auth0 connection. (EA only)"`
- These descriptions flow into the generated `docs/` via `make docs` (tfplugindocs) — no separate doc banner needed.
- **No conditional logic, no flag checks** in the resource; the server enforces EA access (returns 403 if the tenant lacks the flag). The provider just calls the API normally.
- Register in `provider.go` like any other resource/data source.

→ **Apply the same:** mark `auth0_rate_limit_policy` (resource + data source) descriptions with `(EA only)`; no gating code.

---

## 7. Assumptions

- **A1:** `go-auth0/v2 v2.12.0` already vendored is the SDK version we build against — no bump needed. *(Verified in go.mod.)*
- **A2:** No Auth0 **Authentication API SDK** changes are needed (SDKREQ-196 explicitly states this); CDT change is Mgmt-API only.
- **A3:** [api2-3.1-internal.json](api2-3.1-internal.json) is the **source of truth** for API structure/params and matches the merged OAS PR #12200; enum values (`oauth_authentication_api`, `client`) and the `0–10000` limit bound won't change before EA.
- **A4:** Standard acceptance-test infra (HTTP recordings, `make test-acc`) applies; no new test harness needed.
- **A5:** Resource follows the established v2-SDK package pattern (expand/flatten split, mutex, `HandleAPIError`).

---

## 8. Dependencies

| Dependency | Owner | State | Impact on us |
|---|---|---|---|
| OpenAPI public schema (EA flag removed) | Platform Services | PR in review | Defines contract; merged in api2 #12200 ✅ |
| Go SDK generated | Authentication SDKs | **Satisfied** (v2.12.0) ✅ | None — unblocked |
| `rate_limit_policies_management` server flag enabled on test tenant | Platform Services | — | Needed for live acceptance testing |
| ingress-nginx → Istio rollout (ROAD-2371/8534) | Platform Network | ~end June | Gates **product GA**, not our code |
| Node SDK (node-auth0 #1348) / Deploy CLI ([#1395](https://github.com/auth0/auth0-deploy-cli/pull/1395)) / Auth0 CLI (DXCDT-1674) | CDT / SDKs | Deploy CLI PR **open** | Parallel; share the same contract — keep field names 1:1 with Deploy CLI JSON/YAML keys |

---

## 9. Risks & Mitigations

| # | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| R1 | Union-typed `configuration` mismodeled in TF schema → confusing diffs / plan errors. | Med | High | Single nested block with cross-field validation; thorough expand/flatten unit + acceptance tests covering all 4 actions. |
| R2 | Product launch slips (DPR Red, infra blocker) and we ship a resource against an endpoint not yet broadly enabled. | High | Med | Ship as **EA-only** (documented), gate live tests on the feature flag, coordinate release notes with product. |
| R3 | OAS contract changes after EA feedback (new actions/consumers/resources). | Med | Med | Keep enum validation centralized & forgiving; isolate union handling in `expand.go`/`flatten.go`. |
| R4 | Endpoint is `x-internal: true` / `beta` — may not be sanctioned for public Terraform exposure yet. | Med | High | Resolve OQ#5 with API/SDK owners before merging; gate behind EA docs. |
| R5 | Naming drift between Terraform, Deploy CLI, and Auth0 CLI for the same concept. | Low | Med | Align field/attribute names with DXCDT-1674 owner (Ankita Tripathi) early. |
| R6 | Acceptance tests can't run because the flag is off / client not global on the shared test tenant. | Med | Low | Use HTTP recordings; document the flag + global-client requirement in the test file header. |
| R7 | Mismodeling the `oneOf` config (e.g. sending `limit` with `allow`) → 400s, since each variant is `additionalProperties: false`. | Med | High | Strict per-action expand + `CustomizeDiff` validation; acceptance tests for all 4 actions incl. negative cases. |

---

## 10. Definition of Done

- [ ] `auth0_rate_limit_policy` **resource** with full CRUD + import, in `internal/auth0/ratelimitpolicy/`.
- [ ] `auth0_rate_limit_policy` **singular data source** (Get by `id`) — **in scope** (Decision #1, §5.5.1).
- [ ] `auth0_rate_limit_policies` **plural data source** (filtered List + pagination) — **in scope** (Decision #1, §5.5.1).
- [ ] All three registered in `internal/provider/provider.go` (`ResourcesMap` + `DataSourcesMap` × 2).
- [ ] Expand/flatten handle all `configuration` variants (`allow`/`block`/`log`/`redirect`), emitting only action-valid fields.
- [ ] **Live API testing done first** (by Kiran), then `ValidateFunc`s added (enums, `limit` 0–10000, https `redirect_uri`) — Decision #3 sequence.
- [ ] Acceptance tests with recordings, all actions covered; passes `make test`. (Test tenant has `rate_limit_policies_management` enabled — Decision #2; no 403-disabled handling.)
- [ ] `make docs` regenerated; `examples/resources/auth0_rate_limit_policy/` + `examples/data-sources/auth0_rate_limit_policy/` added.
- [ ] CHANGELOG entry + PR linked back to DXCDT-1673.
- [ ] **EA convention applied** — `(EA only)` on resource + data source + attribute descriptions (§6.1); no gating code.
- [ ] `make lint` clean.

---

## 11. Immediate next step

Per Decision #3, the next action is **live API testing** against a tenant with `rate_limit_policies_management` enabled (Kiran to provide setup) — exercise all 5 endpoints and all 4 `configuration` actions to confirm exact accepted values, the `limit` bound, `redirect_uri` rules, and the 409 uniqueness behavior. Findings then drive the `ValidateFunc` definitions and the recorded acceptance-test fixtures.

---

*Compiled from Jira (DXCDT-1673/1672/1674, ROAD-6548, SDKREQ-196), the Rate Limit Policies / Edge Rate Limits EA Confluence checklists, the **`api2-3.1-internal.json` OpenAPI spec (source of truth for API structure & parameters)**, the Deploy CLI sibling PR ([#1395](https://github.com/auth0/auth0-deploy-cli/pull/1395)), direct inspection of `go-auth0/v2 v2.12.0` (vendored), and direct decisions from the ticket owner (2026-05-31).*
