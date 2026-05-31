![goswagger](https://github.com/R0X4R/goswagger/blob/main/.github/image.png?raw=true)

`goswagger` is a minimal [**SwaggerHub**](https://swagger.io/) **OSINT scanner** written in Go.
It searches SwaggerHub APIs for a query, fetches **discovered target URLs**, and matches each **fetched body against regex patterns** from `regex.yaml`.

## Background

This project is a Go rewrite created to improve my Go coding skills, and it is totally inspired by the original Python [**SwaggerSpy**](https://github.com/UndeadSec/SwaggerSpy) project.

### Overview

The goal is to make SwaggerHub reconnaissance fast and repeatable for security researchers, developers, and IT teams by combining:

- API discovery from SwaggerHub search results
- configurable regex-based inspection
- minimal, scan-friendly output

### Swagger And OpenAPI

Swagger (OpenAPI tooling) is a standard ecosystem for describing REST APIs in JSON or YAML, generating documentation, and improving API integration workflows.

### SwaggerHub Context

SwaggerHub is a collaborative API platform built around Swagger/OpenAPI where teams design, version, and publish API definitions.

### Why OSINT Matters

Public API responses can accidentally include sensitive values. OSINT scanning helps reduce this risk by:

1. Catching developer oversights early
2. Supporting secure-by-default development habits
3. Reducing chances of credential and token leaks
4. Helping risk and exposure triage
5. Supporting compliance and privacy reviews
6. Creating practical feedback loops for engineering teams

### Detection Workflow

`goswagger` queries SwaggerHub, collects discovered target URLs, downloads each response, and applies regex patterns to detect potential secrets and credentials.

## Install

```bash
go install github.com/R0X4R/goswagger@latest
```

After the first run, `goswagger` automatically seeds the default regex file at:

```text
~/.config/goswagger/regex.yaml
```

If the file already exists, it is left unchanged so your custom regex edits are preserved.

**Install from source**

```bash
git clone https://github.com/R0X4R/goswagger.git && cd goswagger && go install .
```

## Usage

```bash
goswagger [flags]
```

Example:

```bash
goswagger -q example.com -t 25
```

Flags:

| Flag              | Description                                                               |
| ----------------- | ------------------------------------------------------------------------- |
| `-q, -query`      | Search query required by SwaggerHub                                       |
| `-r, -regex-file` | Path to the regex YAML file, defaults to `~/.config/goswagger/regex.yaml` |
| `-t, -threads`    | Worker count for fetching and matching, defaults to `25`                  |
| `-m, -max-pages`  | Maximum SwaggerHub pages to fetch, `0` means all pages                    |
| `-b, -base-url`   | Override the SwaggerHub search URL format                                 |
| `-o, -output`     | Append matches to a text file while still printing to stdout              |

## Output

Matches are printed in a compact bracketed format:

```text
[HIGH] [GITHUB PERSONAL ACCESS TOKEN] https://example.com/swagger.json [ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx]
```

## Regex File Format

`regex.yaml` uses a structured pattern format with metadata support.

Each pattern entry contains:

* `regex` → the detection regex
* `description` → human-readable description
* optional `confidence` → severity or confidence level

Example:

```yaml
patterns:

  github_personal_access_token:
    regex: 'ghp_[a-zA-Z0-9]{36}'
    description: GitHub Personal Access Token
    confidence: high

  stripe_live_secret_key:
    regex: 'sk_live_[0-9a-zA-Z]{24}'
    description: Stripe Live Secret Key
    confidence: high

  generic_api_key:
    regex: '(?i)(api[_\-]?key|apikey)[\s]*[:=][\s]*[a-zA-Z0-9_\-]{16,64}'
    description: Generic API Key
    confidence: medium
```

## Pattern Fields

| Field         | Required | Description                                    |
| ------------- | -------- | ---------------------------------------------- |
| `regex`       | Yes      | Go-compatible regular expression               |
| `description` | Yes      | Human-readable detection label                 |
| `confidence`  | No       | Detection confidence (`low`, `medium`, `high`) |

## Adding New Regex Patterns

1. Open `regex.yaml`
2. Add a new entry under `patterns:`
3. Provide a regex and description
4. Run a scan to validate the pattern

**Example:**

```yaml
patterns:

  my_custom_token:
    regex: 'token_[A-Za-z0-9]{24}'
    description: Custom Internal Token
    confidence: medium
```

**Output structure:**

```text
[CONFIDENCE] [PATTERN NAME] URL [MATCH]
```

## Notes About Regex Compatibility

* Regexes use Go's `regexp` engine
* Invalid regex entries are skipped automatically
* YAML strings should usually use single quotes `'...'`
* Escape backslashes properly when needed

**Example:**

```yaml
regex: 'sk_live_[0-9a-zA-Z]{24}'
```

**Not:**

```yaml
regex: "sk_live_[0-9a-zA-Z]{24}"
```

unless escaping is required.

## Recommended Confidence Levels

| Confidence | Meaning                                       |
| ---------- | --------------------------------------------- |
| `high`     | Very likely to be a real credential or secret |
| `medium`   | Possible secret, may generate false positives |
| `low`      | Informational or noisy detections             |


## Examples

Run a scan and print to the terminal:

```bash
goswagger -q admin -t 5
```

Run a scan and save output to a file:

```bash
goswagger -q admin -t 5 -o results/output.txt
```

Limit the crawl to a small number of SwaggerHub pages:

```bash
goswagger -q swagger -m 2
```

## Local Test Server

Run the bundled development server for local scanner testing while developing:

```bash
go run ./testserver/cmd
```

Environment variables:

| Variable                 | Description                                         |
| ------------------------ | --------------------------------------------------- |
| `TESTSERVER_PORT`        | Port to listen on, defaults to `8081`               |
| `TESTSERVER_MODE`        | `basic` or `multi`, defaults to `multi`             |
| `TESTSERVER_BASE_PATH`   | Search endpoint path, defaults to `/apiproxy/specs` |
| `TESTSERVER_TOTAL_COUNT` | Override the SwaggerHub `totalCount` value          |
| `TESTSERVER_DELAY_MS`    | Add a response delay for slow-network testing       |

The dev server exposes:

| Endpoint                                          | Purpose                                                   |
| ------------------------------------------------- | --------------------------------------------------------- |
| `/healthz`                                        | Health check endpoint                                     |
| `/apiproxy/specs`                                 | Search endpoint used by the scanner                       |
| `/specs/1`, `/specs/2`, `/specs/3`, `/specs/edge` | Sample bodies with different token and edge-case patterns |

## Development Workflow

1. Start the test server in one terminal:

    ```bash
    go run ./testserver/cmd
    ```

2. Run goswagger against the local server in another terminal:

    ```bash
    goswagger -q testquery -m 2 -b "http://127.0.0.1:8081/apiproxy/specs?sort=BEST_MATCH&order=DESC&query=%s&page=%d&limit=100"
    ```

3. Edit `regex.yaml`, `pkg/*.go`, or the test server, then rerun the same command to verify behavior.

**Useful dev modes:**

| Mode                                   | Effect                                       |
| -------------------------------------- | -------------------------------------------- |
| `TESTSERVER_MODE=basic`                | Return a single page and a single sample URL |
| `TESTSERVER_MODE=multi`                | Exercise multiple pages and patterns         |
| `TESTSERVER_DELAY_MS=500`              | Simulate a slower service                    |
| `TESTSERVER_BASE_PATH=/apiproxy/specs` | Override the search endpoint path            |

## Notes

- Output is intentionally minimal.
- Invalid regex entries are skipped during compilation.
- The tool keeps scanning even if one pattern or one URL fails.

## Credits

- Original inspiration: [SwaggerSpy by UndeadSec](https://github.com/UndeadSec/SwaggerSpy)
