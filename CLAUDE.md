# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # build binary (injects version/commit/date via ldflags)
make install        # install to $GOPATH/bin
make test           # go test ./...
make lint           # golangci-lint run ./...
make clean          # remove binary

go build -o dartcli .          # quick dev build (no version injection)
go test ./internal/render/...  # run tests for a specific package
```

The binary requires a DART API key at runtime. Set `DART_API_KEY` env var or use `~/.dartcli/config.yaml`.

## Architecture

```
main.go → cmd.Execute()
cmd/           CLI layer (cobra commands, global state)
internal/      Core logic (no direct CLI dependencies)
  api/         DART OpenAPI HTTP client
  cache/       Corp code cache (~12k companies, persisted at ~/.dartcli/cache/corpcode.json)
  config/      Viper-based config (~/.dartcli/config.yaml + env vars)
  render/      API response → markdown, DART XML → markdown
  httpclient/  HTTP client with relaxed TLS (DART servers use legacy TLS 1.0+)
pkg/dartcli/   Version metadata (injected by ldflags at build time)
```

### cmd/root.go — shared state

`root.go` owns three globals used across all commands:
- `cfg` — loaded once via `cobra.OnInitialize`
- `renderer` — glamour-based terminal renderer
- `corpStore` — lazy-loaded on first corp lookup (`loadCorpStore()`)

`resolveCorpCode(query)` searches the corp store (exact stock code → exact corp code → exact name → substring), and if multiple matches are found presents an interactive `huh.Select` prompt.

### internal/api — DART OpenAPI client

All endpoints are under `https://opendart.fss.or.kr`. Status `"000"` = success; `"013"` = no data (treated as empty result, not an error). The client injects the API key as a query parameter on every request.

Financial period codes: `11011` annual, `11013` Q1, `11012` half-year, `11014` Q3.

### internal/cache — corp code store

Downloads `GET /api/corpCode.xml` (returns a ZIP containing CORPCODE.xml), parses the XML, and saves as JSON. Auto-refreshes when older than 7 days; falls back to stale data with a warning if the refresh fails. The in-memory `Store` has three indices: by corp code, by lowercase name, by stock code.

### internal/render/document.go — DART XML → markdown

This is the most complex file. DART discloses documents as ZIP archives containing proprietary `dart4` XML. The parser uses a streaming `xml.Decoder` with an element-name stack to produce markdown:

- **`sanitizeXML()`** must run before decoding — DART XML routinely embeds Korean text in bare angle brackets (e.g. `<이사ㆍ감사 보수현황>`, `< TV 시장점유율 >`) that are visually used as emphasis markers but cause Go's xml decoder to fatal-error. The sanitizer replaces any `<` not followed by a valid XML name start character (`[a-zA-Z_:/?!>]`) with `&lt;`.
- **Do not set `dec.AutoClose = xml.HTMLAutoClose`** — DART XML uses explicit `<IMG>...</IMG>` pairs; auto-closing IMG causes a "unexpected end element `</IMG>`" fatal error.
- `LIBRARY` is a transparent container (like `SECTION-*`), not metadata to skip. It wraps entire document sections in complex reports.
- `SECTION-N` depth determines heading level: `SECTION-1` → `##`, `SECTION-2` → `###`, etc. The heading level is derived from the stack at the time a `<TITLE>` element is encountered.
- Single-cell single-row tables are rendered as paragraphs (avoids glamour bullet rendering).
- `DocumentFromZIP` reads up to 3 largest XML/HTML files from the ZIP, sorted by: rcptNo-matching filename first, then by size descending.

### Rendering pipeline

`renderer.Print(md)` uses glamour with TTY detection (`go-isatty`) and auto terminal width (capped at 120). `renderer.PrintWide(md)` sets `WithWordWrap(0)` — used for document view to prevent wide financial table truncation.

Financial amounts are formatted in 억원 (÷100,000,000) when ≥ 1억, otherwise 만원.

## Key DART API quirks

- The API key goes in the query string as `crtfc_key`.
- The corp code (`corp_code`) is an 8-digit zero-padded string, distinct from the 6-digit stock ticker (`stock_code`).
- `GET /api/document.xml?rcpNo=<접수번호>` returns a ZIP (despite the `.xml` endpoint name).
- ZIPs for 사업보고서 can contain multiple XML files (e.g. main + two supplementary audit files). All up to 3 files are rendered in sequence.
