# API Test Plan — Rate Limit Policies (EA)

> **Endpoint group:** `/api/v2/rate-limit-policies` (5 operations) — DXCDT-1673
> **Spec (source of truth):** [api2-3.1-internal.json](api2-3.1-internal.json)
> **SDK:** `go-auth0/v2 v2.12.0` — `Management.RateLimitPolicies`
> **Status:** ⏸️ Staged — awaiting a token carrying `read/create/update/delete:rate_limit_policies` (see [DXCDT-1673-API-QA.md](DXCDT-1673-API-QA.md) for the current scope blocker).
> **Purpose:** Execute end-to-end once unblocked; record actual req/resp under each case; flag every deviation from spec.

---

## 0. How to run this plan

### 0.1 Prereqs
1. A bearer token in `token.txt` (or `tus_token.txt`) with all four `*:rate_limit_policies` scopes, from a **global** client (spec 401: "Client is not global"), on a tenant with `rate_limit_policies_management` enabled.
2. Confirm scopes are live first:
   ```bash
   TOK=$(tr -d '\n\r ' < token.txt)
   BASE="https://<TENANT>/api/v2"     # e.g. https://kiran-dev.us.auth0.com/api/v2
   # Expect the 4 rate_limit scopes to be listed:
   curl -s -H "Authorization: Bearer $TOK" "$BASE/resource-servers" \
     | jq '.[] | select(.identifier|endswith("/api/v2/")) | [.scopes[].value | select(test("rate_limit"))]'
   ```
   If that array is empty → **stop**, the blocker is not resolved.

### 0.2 Conventions used below
- Every case lists: **Request** (curl), **Expected** (status + body per spec), **Capture** (what to paste in the results table), **Why** (what it validates).
- Replace `$BASE`, `$TOK`, and `{id}` placeholders.
- Always capture status code (`-w "\n[HTTP %{http_code}]\n"`) **and** body.
- Record response headers on at least one call per verb (look for `X-RateLimit-*`).
- **Run order matters:** §3 (list-empty) → §4 (create) → §5 (get) → §6 (update) → §7 (delete). Negative/edge cases interleaved.
- Keep created IDs in a scratch list; clean up everything in §8.

### 0.3 Reference: spec contract (condensed)

**`consumer_selector`** (string, ≤255): `client_id:<client_id>` · `client_id:<cimd_uri>` · `cimd_clients` · `third_party_clients` · `default`

**`configuration`** = `oneOf` discriminated by `action`, each `additionalProperties:false`:

| Variant | `action` | `limit` (int 0–10000) | `redirect_uri` (strict-https-uri) |
|---|---|---|---|
| allow | `allow` | ✗ not allowed | ✗ not allowed |
| throttle | `block` \| `log` | ✓ required | ✗ not allowed |
| redirect | `redirect` | ✓ required | ✓ required |

**Enums:** `resource` ∈ {`oauth_authentication_api`}; `consumer` ∈ {`client`}.
**`id`:** ≤26 chars, format `rate-limit-policy-id`. **`created_at`/`updated_at`:** RFC3339.

---

## 1. Endpoint inventory

| # | Op | Method | Path | Scope | Success | Section |
|---|---|---|---|---|---|---|
| E1 | list | GET | `/rate-limit-policies` | read | 200 | §3 |
| E2 | create | POST | `/rate-limit-policies` | create | 201 | §4 |
| E3 | get | GET | `/rate-limit-policies/{id}` | read | 200 | §5 |
| E4 | update | PATCH | `/rate-limit-policies/{id}` | update | 200 | §6 |
| E5 | delete | DELETE | `/rate-limit-policies/{id}` | delete | 204 | §7 |

---

## 2. Pre-flight / auth sanity (run once)

| ID | Scenario | Request | Expected | Why |
|---|---|---|---|---|
| P1 | Token valid at all | `GET $BASE/clients?per_page=1` | 200 | Confirms token works before blaming the feature. |
| P2 | Scopes present | resource-server jq check (§0.1) | 4 rate_limit scopes listed | Confirms blocker resolved. |
| P3 | No-auth call | `GET $BASE/rate-limit-policies` with **no** `Authorization` | 401 | Baseline auth enforcement. |
| P4 | Bad token | `Authorization: Bearer garbage` | 401 invalid token | Spec 401 path. |

---

## 3. E1 — List `GET /rate-limit-policies`

### 3.1 Happy / structure
| ID | Scenario | Request | Expected | Capture |
|---|---|---|---|---|
| L1 | Empty list (before any create) | `curl -s -w '\n[%{http_code}]' -H "Authorization: Bearer $TOK" "$BASE/rate-limit-policies"` | 200; body `{"rate_limit_policies":[...]}`; **verify** `next` absent when no more pages | full body |
| L2 | Default `take` | `...?take=50` | 200; same as default | note if identical to L1 |
| L3 | Shape audit | (use any populated response) | Each item has exactly `id,resource,consumer,consumer_selector,configuration,created_at,updated_at`; **no extra fields** | list any extra/missing keys |

### 3.2 Pagination
| ID | Scenario | Request | Expected | Why |
|---|---|---|---|---|
| L4 | `take=1` with ≥2 policies | `...?take=1` | 200; 1 item; `next` present (cursor) | min page size |
| L5 | Follow cursor | `...?take=1&from=<next>` | 200; next item; eventually `next` absent | cursor round-trip |
| L6 | `take=100` (max) | `...?take=100` | 200 | upper bound accepted |
| L7 | `take=0` | `...?take=0` | **likely 400** (spec min 1) — *verify* | boundary |
| L8 | `take=101` (over max) | `...?take=101` | **likely 400** (spec max 100) — *verify* | boundary |
| L9 | `take=abc` | `...?take=abc` | 400 | type validation |
| L10 | `from=<garbage>` | `...?from=not-a-cursor` | *verify* — 400 or 200-empty? | cursor robustness |
| L11 | `from` length 1001 | over `maxLength:1000` | *verify* 400 | boundary |

### 3.3 Filters
| ID | Scenario | Request | Expected | Why |
|---|---|---|---|---|
| L12 | Filter by resource | `...?resource=oauth_authentication_api` | 200; only matching | valid filter |
| L13 | Filter by consumer | `...?consumer=client` | 200 | valid filter |
| L14 | Filter by selector | `...?consumer_selector=default` | 200; only that selector | valid filter |
| L15 | Invalid resource | `...?resource=bogus` | **verify** — 400 (enum) or ignored? | enum enforcement on query |
| L16 | Invalid consumer | `...?consumer=bogus` | **verify** 400 vs ignored | enum enforcement |
| L17 | `consumer_selector` 256 chars | over max 255 | *verify* 400 | boundary |
| L18 | Combined filters | `...?resource=...&consumer=...&consumer_selector=...` | 200; AND semantics | combination |

---

## 4. E2 — Create `POST /rate-limit-policies`

### 4.1 Happy paths — one per config variant (these create the fixtures used later)
| ID | Body | Expected |
|---|---|---|
| C1 `allow` | `{"resource":"oauth_authentication_api","consumer":"client","consumer_selector":"default","configuration":{"action":"allow"}}` | **201**; echoes input + `id`,`created_at`,`updated_at`; capture `id` → `$ID_ALLOW` |
| C2 `block` | `{"resource":"oauth_authentication_api","consumer":"client","consumer_selector":"client_id:AAAA1111","configuration":{"action":"block","limit":100}}` | 201; → `$ID_BLOCK` |
| C3 `log` | `{"resource":"oauth_authentication_api","consumer":"client","consumer_selector":"cimd_clients","configuration":{"action":"log","limit":250}}` | 201; → `$ID_LOG` |
| C4 `redirect` | `{"resource":"oauth_authentication_api","consumer":"client","consumer_selector":"third_party_clients","configuration":{"action":"redirect","limit":50,"redirect_uri":"https://example.com/blocked"}}` | 201; → `$ID_REDIRECT` |

Reference curl for C1:
```bash
curl -s -w "\n[%{http_code}]\n" -X POST -H "Authorization: Bearer $TOK" -H "Content-Type: application/json" \
  -d '{"resource":"oauth_authentication_api","consumer":"client","consumer_selector":"default","configuration":{"action":"allow"}}' \
  "$BASE/rate-limit-policies"
```

**For each happy create, capture & verify:**
- [ ] `id` format: length ≤26, matches `rate-limit-policy-id` pattern (record actual, e.g. `rlp_...`)
- [ ] `created_at` and `updated_at` present, RFC3339, equal on create
- [ ] Response config echoes exactly what was sent (no coercion)

### 4.2 `consumer_selector` variants (valid)
| ID | `consumer_selector` | Expected |
|---|---|---|
| C5 | `client_id:<real client_id from this tenant>` | 201 — *does it validate the client exists?* (RFD/IPS-5789 says "Validate client exists") |
| C6 | `client_id:<nonexistent id>` | **verify** — 400/404 (client-exists validation) vs 201 |
| C7 | `client_id:<cimd_uri>` (e.g. `client_id:https://cimd.example/app`) | 201 — CIMD form |
| C8 | `default` | covered by C1 |

### 4.3 `limit` boundary & type matrix (use `block` to isolate `limit`)
Base body: `{"resource":"oauth_authentication_api","consumer":"client","consumer_selector":"client_id:LIMITTEST<n>","configuration":{"action":"block","limit":<X>}}`
| ID | `limit` value | Expected |
|---|---|---|
| C9 | `0` (min) | 201 (spec `minimum:0`) — *verify 0 is actually accepted* |
| C10 | `10000` (max) | 201 |
| C11 | `-1` | 400 |
| C12 | `10001` (over) | 400 |
| C13 | `1.5` (float) | 400 (type integer) |
| C14 | `"100"` (string) | **verify** 400 vs coerced |
| C15 | `null` | 400 (required for block) |
| C16 | missing `limit` | 400 (required for block/log/redirect) |

### 4.4 `configuration` union violations
| ID | Body fragment | Expected | Why |
|---|---|---|---|
| C17 | `allow` + `limit`: `"configuration":{"action":"allow","limit":100}` | **400** | `additionalProperties:false` on allow variant |
| C18 | `block` + `redirect_uri`: `{"action":"block","limit":10,"redirect_uri":"https://x.com"}` | **400** | redirect_uri not allowed on block |
| C19 | `redirect` missing `redirect_uri`: `{"action":"redirect","limit":10}` | 400 | required |
| C20 | `redirect` missing `limit`: `{"action":"redirect","redirect_uri":"https://x.com"}` | 400 | required |
| C21 | unknown action: `{"action":"deny","limit":10}` | 400 | enum |
| C22 | action missing: `{"limit":10}` | 400 | required discriminator |
| C23 | `configuration` empty `{}` | 400 | no variant matches |
| C24 | `configuration` null | 400 | required |
| C25 | `configuration` missing entirely | 400 | required |
| C26 | extra prop in config: `{"action":"allow","foo":1}` | 400 | additionalProperties:false |

### 4.5 `redirect_uri` format (use `redirect` action)
| ID | `redirect_uri` | Expected |
|---|---|---|
| C27 | `http://example.com` (non-https) | **400** (`strict-https-uri`) |
| C28 | `https://example.com/path?q=1` | 201 |
| C29 | `ftp://x` / `notaurl` | 400 |
| C30 | `""` (empty) | 400 |
| C31 | `null` | 400 |
| C32 | very long URI (>2048) | *verify* behavior |
| C33 | `https://` only (no host) | *verify* 400 |

### 4.6 Top-level field validation
| ID | Scenario | Expected |
|---|---|---|
| C34 | missing `resource` | 400 |
| C35 | missing `consumer` | 400 |
| C36 | missing `consumer_selector` | 400 |
| C37 | `resource="bogus"` | 400 (enum) |
| C38 | `consumer="bogus"` | 400 (enum) |
| C39 | `consumer_selector=""` | **verify** 400 |
| C40 | `consumer_selector` 256 chars | 400 (>255) |
| C41 | `consumer_selector` exactly 255 | 201 |
| C42 | extra top-level field `{"foo":"bar",...}` | 400 (additionalProperties:false) |
| C43 | empty body `{}` | 400 |
| C44 | malformed JSON `{not json` | 400 |
| C45 | wrong content-type (form-encoded body) | *verify* — spec lists `application/x-www-form-urlencoded` as accepted; test it |
| C46 | `resource`/`consumer` correct casing only? try `OAUTH_AUTHENTICATION_API` | *verify* case-sensitivity |

### 4.7 Uniqueness (409)
| ID | Scenario | Expected | Why |
|---|---|---|---|
| C47 | Re-create same `(resource,consumer,consumer_selector)` as C1 (`default`) | **409** "already exists" | spec 409 |
| C48 | Same selector but different `configuration` | **409** (uniqueness is on the tuple, not config) — *verify* | confirms tuple semantics |
| C49 | Same selector after deleting the original (see §7) | 201 | delete frees the tuple |

---

## 5. E3 — Get `GET /rate-limit-policies/{id}`

| ID | Scenario | Request | Expected |
|---|---|---|---|
| G1 | Valid `$ID_BLOCK` | `GET $BASE/rate-limit-policies/$ID_BLOCK` | 200; body identical to what create returned (incl config) |
| G2 | Each variant | GET `$ID_ALLOW`,`$ID_LOG`,`$ID_REDIRECT` | 200; correct config per variant |
| G3 | Nonexistent id | `GET .../rlp_doesnotexist0000000000` | **404** "does not exist" |
| G4 | Malformed id (>26 chars) | `GET .../<40-char string>` | **verify** 404 vs 400 |
| G5 | Empty id (`/rate-limit-policies/`) | trailing slash | likely hits list or 404 — *verify* |
| G6 | id with special chars | `GET .../rlp_%20` | *verify* |
| G7 | Shape audit | any | exactly the documented fields, no extras |

---

## 6. E4 — Update `PATCH /rate-limit-policies/{id}`

> Spec: PATCH body = `{"configuration": <union>}` only. `configuration` is **required** in `PatchRateLimitPolicyRequestContent`.

### 6.1 Valid config transitions (on $ID_BLOCK or fresh fixtures)
| ID | Body | Expected | Verify |
|---|---|---|---|
| U1 | `{"configuration":{"action":"allow"}}` | 200 | config now allow; `updated_at` changed; `created_at` unchanged |
| U2 | `{"configuration":{"action":"block","limit":500}}` | 200 | limit updated |
| U3 | `{"configuration":{"action":"log","limit":1}}` | 200 | block→log transition allowed |
| U4 | `{"configuration":{"action":"redirect","limit":10,"redirect_uri":"https://new.example.com"}}` | 200 | →redirect, uri set |
| U5 | redirect → allow (drops limit & uri) | `{"configuration":{"action":"allow"}}` | 200; verify limit/uri removed from response |

### 6.2 Config validation on PATCH (same union rules as create)
| ID | Body | Expected |
|---|---|---|
| U6 | `{"configuration":{"action":"block"}}` (no limit) | 400 |
| U7 | `{"configuration":{"action":"redirect","limit":10}}` (no uri) | 400 |
| U8 | `{"configuration":{"action":"redirect","limit":10,"redirect_uri":"http://x"}}` (non-https) | 400 |
| U9 | `{"configuration":{"action":"allow","limit":5}}` | 400 (additionalProperties) |
| U10 | `{"configuration":{"action":"deny"}}` | 400 (enum) |
| U11 | `limit` = -1 / 10001 / 1.5 / "5" | 400 each |
| U12 | `{"configuration":{}}` | 400 |
| U13 | `{"configuration":null}` | 400 |
| U14 | `{}` (no configuration) | 400 (required) |

### 6.3 Immutability of identity fields (**key for ForceNew decision**)
| ID | Body | Expected | Why |
|---|---|---|---|
| U15 | `{"resource":"oauth_authentication_api","configuration":{"action":"allow"}}` | **verify** — 400 (additionalProperties) OR 200-but-ignored | Is `resource` rejected or silently ignored? |
| U16 | `{"consumer_selector":"client_id:CHANGED","configuration":{"action":"allow"}}` | **verify** — confirm selector does NOT change | Confirms ForceNew necessity |
| U17 | `{"consumer":"client","configuration":{"action":"allow"}}` | **verify** | same |
| U18 | `{"id":"rlp_other","configuration":{"action":"allow"}}` | **verify** ignored/400 | id immutable |

> Record exactly whether these are **rejected (400)** or **accepted-and-ignored**. Either way `resource/consumer/consumer_selector` = `ForceNew` in TF, but the provider error UX differs.

### 6.4 PATCH error paths
| ID | Scenario | Expected |
|---|---|---|
| U19 | PATCH nonexistent id | 404 |
| U20 | malformed JSON | 400 |
| U21 | extra top-level field alongside configuration | **verify** 400 |
| U22 | form-encoded body | *verify* (spec lists it as accepted) |

---

## 7. E5 — Delete `DELETE /rate-limit-policies/{id}`

| ID | Scenario | Request | Expected |
|---|---|---|---|
| D1 | Delete `$ID_REDIRECT` | `DELETE $BASE/rate-limit-policies/$ID_REDIRECT` | **204**, empty body |
| D2 | GET after delete | `GET .../$ID_REDIRECT` | 404 (confirm gone) |
| D3 | Delete already-deleted | `DELETE .../$ID_REDIRECT` again | 404 |
| D4 | Delete nonexistent | `DELETE .../rlp_neverexisted00000000` | 404 |
| D5 | Re-create freed tuple | POST C4's body again (third_party_clients) | 201 (uniqueness freed) → confirms C49 |
| D6 | Delete malformed id | `DELETE .../<40 chars>` | *verify* 404 vs 400 |

---

## 8. Cleanup

| ID | Action |
|---|---|
| X1 | List all, delete every `rlp_*` created during this run (`$ID_ALLOW`, `$ID_BLOCK`, `$ID_LOG`, any LIMITTEST/redirect fixtures). |
| X2 | Final `GET /rate-limit-policies` → confirm tenant back to pre-test state. |

---

## 9. Cross-cutting checks (note throughout)

- [ ] **Response schema strictness:** every 200/201 body matches spec exactly — no undocumented fields, no missing documented fields.
- [ ] **Error body shape:** all errors are `{statusCode,error,message,errorCode}`; record the `errorCode` values seen (e.g. `insufficient_scope`, `invalid_body`, …).
- [ ] **Status-code fidelity:** 201 (not 200) on create; 204 (not 200) on delete.
- [ ] **`created_at` vs `updated_at`:** equal on create; `updated_at` advances on PATCH; `created_at` immutable.
- [ ] **Rate-limit headers:** capture `X-RateLimit-Limit/Remaining/Reset` on a sample call; if 429 hit, record body.
- [ ] **Idempotency/ordering:** does list reflect creates immediately (read-after-write consistency)?
- [ ] **`Content-Type` variants:** JSON vs form-encoded both accepted per spec — verified?
- [ ] **Client-exists validation:** does `client_id:<selector>` validate against real clients (RFD claim)? Critical for TF `ValidateFunc` design.

---

## 10. Results capture template

> Duplicate this block per test ID as you execute. Keep raw req/resp — verbosity is fine.

```
### <ID> — <scenario>
Request:
  <method> <path>
  <body if any>
Response:
  HTTP <code>
  <body>
  <relevant headers>
Expected: <from plan>
Result: ✅ matches spec  |  ⚠️ deviation  |  ❌ fail
Notes / deviation detail:
```

### 10.1 Deviation log (fill as found)

| ID | Endpoint | Expected (spec) | Actual | Severity | Impact on TF provider |
|---|---|---|---|---|---|
| | | | | | |

---

## 11. Feed-forward into the Terraform provider

Findings here directly resolve these implementation choices (track answers as testing completes):

| Question the test answers | Cases | Provider impact |
|---|---|---|
| Are `resource`/`consumer`/`consumer_selector` rejected or ignored on update? | U15–U18 | `ForceNew` (confirmed) + whether to add a plan-time guard |
| Exact `limit` accepted range & type coercion | C9–C16, U11 | `validation.IntBetween(0,10000)`; int type |
| Is `limit=0` semantically valid? | C9 | whether to allow 0 in schema |
| `redirect_uri` scheme/format enforcement | C27–C33, U8 | client-side https `ValidateFunc` |
| Does `client_id:<id>` require an existing client? | C5–C7 | whether provider should validate / document |
| 409 tuple semantics | C47–C49 | error messaging; possible `ConflictsWith`/docs note |
| Real `id` format | C1–C4 | doc examples, import format |
| Enum case-sensitivity & forward values | C37–C38, C46, L15–L16 | `ValidateFunc` strictness |
| Form-encoded acceptance | C45, U22 | n/a (SDK uses JSON) — informational |

---

*Plan derived strictly from `api2-3.1-internal.json`. Execute top-to-bottom once a scoped token is available; paste actuals into §10 and deviations into §10.1, then update the provider validation per §11.*
