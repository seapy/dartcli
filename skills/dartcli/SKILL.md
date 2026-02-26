---
name: dartcli
description: Use the dartcli tool to look up Korean corporate disclosure information from DART (금융감독원 전자공시시스템). Trigger this skill whenever the user asks about Korean company filings, financial statements, corporate disclosures, DART data, Korean stock/company lookup, 공시, 재무제표, 기업개황, or any Korean financial regulatory filing. Also trigger when the user mentions dartcli, DART OpenAPI, or wants to search for Korean companies by name or stock code. Even if the user just says something like "삼성전자 재무정보 알려줘" or "Look up Samsung Electronics financials on DART", use this skill.
---

# dartcli — Korean DART Corporate Disclosure CLI

`dartcli` is a Go-based CLI tool that queries the Korean Financial Supervisory Service's DART electronic disclosure system. It outputs structured markdown — ideal for LLM consumption.

**Repository**: `github.com/seapy/dartcli`

## Prerequisites

1. **DART OpenAPI Key**: Required for all commands except `search`. Users can get a free key at https://opendart.fss.or.kr/uss/umt/EgovMberInsertView.do (instant for personal accounts).
2. **Network access**: The CLI calls the DART OpenAPI. If network is disabled, inform the user.

## Installation

Install via the official install script:

```bash
curl -fsSL https://raw.githubusercontent.com/seapy/dartcli/main/install.sh | sh
```

This auto-detects OS/architecture and installs to `/usr/local/bin`. Verify with:

```bash
dartcli version
```

If `curl` or network is unavailable, tell the user to install manually from https://github.com/seapy/dartcli/releases/latest.

## API Key Configuration

The DART API key must be set before using most commands. Check if it's available:

```bash
# Option 1: Environment variable (recommended)
export DART_API_KEY=<key>

# Option 2: Config file
mkdir -p ~/.dartcli
echo "api_key: <key>" > ~/.dartcli/config.yaml

# Option 3: Per-command flag
dartcli company 삼성전자 --api-key <key>
```

**If no API key is set**, ask the user to provide one or set it up. Do NOT proceed without it (except for `search`).

## Commands Reference

### `search` — Company Search (no API key needed)
Find the exact company name or Corp Code. Use this first when the company name is ambiguous.

```bash
dartcli search 삼성        # Search by name
dartcli search 005930      # Search by stock code
```

### `company` — Company Overview
Basic info: CEO, founding date, address, homepage.

```bash
dartcli company 삼성전자
dartcli company 005930     # Also accepts stock code
```

### `list` — Disclosure List
Recent filings. The **접수번호** (receipt number) from results is used with `view`.

```bash
dartcli list 삼성전자                              # Recent filings (default 20)
dartcli list 삼성전자 --type A --limit 5           # Periodic reports only, max 5
dartcli list 삼성전자 --days 90                    # Last 90 days
dartcli list 삼성전자 --start 20240101 --end 20241231
```

**`--type` codes**: A=정기공시, B=주요사항보고, C=발행공시, D=지분공시, E=기타공시, F=외부감사

### `finance` — Financial Statements
Balance sheet, income statement, cash flow in 억원 (100M KRW) with YoY change.

```bash
dartcli finance 삼성전자                           # Last year, annual, consolidated
dartcli finance 삼성전자 --year 2024
dartcli finance 삼성전자 --year 2024 --period half # Semi-annual
dartcli finance 삼성전자 --year 2024 --type ofs    # Separate (non-consolidated)
```

**`--period`**: `annual` (default) | `q1` | `half` | `q3`
**`--type`**: `cfs` (consolidated, default) | `ofs` (separate)

### `view` — Full Disclosure Document
Read the full filing in markdown, open in browser, or download as ZIP.

```bash
dartcli view 20251114002447                        # Render as markdown in terminal
dartcli view 20251114002447 --browser              # Open in browser
dartcli view 20251114002447 --download             # Save as ZIP
dartcli view 20251114002447 --download -o ./file.zip
```

⚠️ Full annual reports (사업보고서) can be 15,000+ lines. Use grep to filter:
```bash
dartcli view 20250311001085 --no-color | grep -A 30 "사업의 개요"
```

### `cache` — Cache Management
Corp Code cache (~12,000 companies), auto-refreshes every 7 days.

```bash
dartcli cache status     # Check cache info
dartcli cache refresh    # Force refresh
dartcli cache clear      # Clear cache
```

## Global Options

| Option | Description |
|---|---|
| `--api-key <key>` | Override API key |
| `--no-color` | Disable color output |
| `--style <style>` | `auto` (default) / `dark` / `light` / `notty` |
| `--config <path>` | Config file path (default: `~/.dartcli/config.yaml`) |

## Usage Patterns for Claude

### Always use `--no-color` or pipe through `cat`
Color escape codes clutter LLM context. Always append `--no-color` or pipe:
```bash
dartcli finance 삼성전자 --no-color
dartcli view 20251114002447 | cat
```

### Ambiguous company names
When multiple companies share a name, `dartcli` shows an interactive prompt that won't work in non-interactive mode. Always `search` first to get the exact name:
```bash
dartcli search 카카오 --no-color
# Then use the exact name from results
dartcli company "카카오(주)" --no-color
```

### Typical workflow
```bash
# 1. Search for company
dartcli search 카카오 --no-color

# 2. Get company overview
dartcli company 카카오 --no-color

# 3. List recent periodic filings
dartcli list 카카오 --type A --limit 10 --no-color

# 4. Read a specific filing (use 접수번호 from list)
dartcli view <접수번호> --no-color | head -200

# 5. Get financial summary
dartcli finance 카카오 --year 2024 --no-color
```

### Large output handling
For `view` commands on large reports, limit output with `head` or `grep`:
```bash
dartcli view <접수번호> --no-color | head -500
dartcli view <접수번호> --no-color | grep -A 50 "재무제표"
```
