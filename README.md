# aeo-crawler

A breadth-first crawler for the [AEO Protocol v0.1](https://github.com/mizcausevic-dev/aeo-protocol-spec).

Give it one seed origin. It fetches that origin's `/.well-known/aeo.json`, then follows every `authority.primary_sources` URI as a candidate origin to fetch next — up to a configurable depth and total fetch budget. Output is one JSON Lines record per origin attempted, suitable for piping into `jq`, a graph database, or any analytics pipeline.

Built on top of [aeo-sdk-go](https://github.com/mizcausevic-dev/aeo-sdk-go).

## Install

```bash
go install github.com/mizcausevic-dev/aeo-crawler/cmd/aeo-crawler@latest
```

## Usage

```bash
aeo-crawler --seed https://mizcausevic-dev.github.io
```

Output (one JSON object per line):

```json
{"origin":"https://mizcausevic-dev.github.io","depth":0,"success":true,"entity_name":"Miz Causevic","entity_type":"Person","claims_count":6,"audit_mode":"none","fetched_at":"2026-05-12T04:00:00Z"}
{"origin":"https://github.com","depth":1,"success":false,"error":"HTTP 404","fetched_at":"2026-05-12T04:00:01Z"}
{"origin":"https://www.linkedin.com","depth":1,"success":false,"error":"HTTP 404","fetched_at":"2026-05-12T04:00:01Z"}
{"origin":"https://mizcausevic.com","depth":1,"success":false,"error":"HTTP 404","fetched_at":"2026-05-12T04:00:01Z"}
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--seed` | required | Seed origin URL. |
| `--depth` | `2` | Maximum graph distance from the seed. `0` = only fetch the seed. |
| `--max-fetches` | `100` | Global cap on total fetches. |
| `--concurrency` | `4` | Maximum in-flight HTTP requests. |
| `--timeout` | `10` | Per-request timeout in seconds. |

## Useful pipelines

**Count successful AEO declarations:**
```bash
aeo-crawler --seed https://mizcausevic-dev.github.io | jq -c 'select(.success==true)' | wc -l
```

**List unique entity names:**
```bash
aeo-crawler --seed https://mizcausevic-dev.github.io | jq -r 'select(.success==true) | .entity_name' | sort -u
```

**Find origins that declare an `audit_mode` of `signature`:**
```bash
aeo-crawler --seed https://example.com --depth 3 | jq -c 'select(.audit_mode=="signature")'
```

## How discovery works

For each fetched declaration, `authority.primary_sources` is treated as the source of next-hop candidate origins. Each URI is normalized to its scheme + host (path stripped). Already-visited origins are not re-fetched. The crawler does not currently chase `citation_preferences.canonical_links` or `claims[].evidence` — those are roadmap for v0.2.

## Conformance

Operates against AEO Protocol v0.1 declarations at **conformance Level 1 (Declare)**. Signature verification (L2) and audit-report submission (L3) are not invoked; signed documents are recorded as `audit_mode: "signature"` but not verified.

## Dependencies

- [github.com/mizcausevic-dev/aeo-sdk-go](https://github.com/mizcausevic-dev/aeo-sdk-go) — Go SDK for parsing and fetching AEO declarations
- Go standard library (`net/http`, `encoding/json`, `context`, `sync`)

## Development

```bash
go vet ./...
go test -v ./...
go build ./cmd/aeo-crawler
```

Tests use `httptest` to serve fixture AEO documents — no network is required.

## Specification

Full spec at [github.com/mizcausevic-dev/aeo-protocol-spec](https://github.com/mizcausevic-dev/aeo-protocol-spec).

## License

AGPL-3.0.

## Kinetic Gain Protocol Suite

| Spec | Implementation |
|---|---|
| [AEO Protocol](https://github.com/mizcausevic-dev/aeo-protocol-spec) | [aeo-sdk-python](https://github.com/mizcausevic-dev/aeo-sdk-python) · [aeo-sdk-typescript](https://github.com/mizcausevic-dev/aeo-sdk-typescript) · [aeo-sdk-rust](https://github.com/mizcausevic-dev/aeo-sdk-rust) · [aeo-sdk-go](https://github.com/mizcausevic-dev/aeo-sdk-go) · [aeo-cli](https://github.com/mizcausevic-dev/aeo-cli) · **aeo-crawler** (this) |
| [Prompt Provenance](https://github.com/mizcausevic-dev/prompt-provenance-spec) | — |
| [Agent Cards](https://github.com/mizcausevic-dev/agent-cards-spec) | — |
| [AI Evidence Format](https://github.com/mizcausevic-dev/ai-evidence-format-spec) | — |
| [MCP Tool Cards](https://github.com/mizcausevic-dev/mcp-tool-card-spec) | — |

---

**Connect:** [LinkedIn](https://www.linkedin.com/in/mirzacausevic/) · [Kinetic Gain](https://kineticgain.com) · [Medium](https://medium.com/@mizcausevic/) · [Skills](https://mizcausevic.com/skills/)
