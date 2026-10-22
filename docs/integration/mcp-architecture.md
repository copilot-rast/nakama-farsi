# MCP implementation and operation

`integrations/mcp` is a separate Go module using the official `github.com/modelcontextprotocol/go-sdk` v1.7.0, which supports the 2026-07-28 MCP specification. It keeps Nakama's core server and configuration untouched. Its 32 tools call Nakama's authenticated account endpoint and selected native Jazebe RPCs with the configured session. The same adapters serve stdio and loopback-only HTTP. Remote OAuth is not implemented; do not expose the HTTP listener beyond loopback.

## Start

From `integrations/mcp`, after downloading the Go module dependencies:

```powershell
$env:NAKAMA_URL = 'http://127.0.0.1:7350'
$env:NAKAMA_SESSION_TOKEN = '<your Nakama user session>'
go mod tidy
go run . stdio
```

For local HTTP clients:

```powershell
$env:NAKAMA_URL = 'http://127.0.0.1:7350'
go run . http
```

The HTTP endpoint is `http://127.0.0.1:7650/mcp`. `MCP_LISTEN_ADDR` may select another loopback address. Each HTTP request needs `Authorization: Bearer <Nakama user session>`. The server validates that session against Nakama before processing the MCP request. The HTTP transport is stateless, limits requests to 64 KiB, applies a ten second request deadline, rejects cross-origin requests, and binds only to loopback. The SDK handles MCP JSON-RPC, tool discovery, resource and prompt discovery, and version negotiation. Stdio takes its Nakama session from `NAKAMA_SESSION_TOKEN` and writes protocol data only to stdout.

## Catalog

| Primitive | Name / URI | Behavior | Authorization |
|---|---|---|---|
| Tool | `nakama_get_account` | Reads the current caller's Nakama account, user profile, and wallet | Valid Nakama user session |
| Tool | `jazebe_create_brand` | Calls native brand creation | Valid Nakama user session and configured operator allowlist |
| Tool | `jazebe_list_brands` | Lists caller-owned brands | Valid Nakama user session |
| Tool | `jazebe_create_planet` | Calls native planet creation; native RPC checks brand ownership | Valid Nakama user session |
| Tool | `jazebe_list_planets` | Lists active planets | Valid Nakama user session |
| Tool | `jazebe_list_brand_planets` | Lists caller-owned brand drafts and active planets | Valid Nakama user session and brand ownership |
| Tool | `jazebe_get_brand_planet` | Reads a caller-owned planet with settings | Valid Nakama user session and brand ownership |
| Tool | `jazebe_update_brand_planet` | Updates caller-owned planet identity | Valid Nakama user session and brand ownership |
| Tool | `jazebe_check_planet_publish` | Reports unmet planet publish requirements | Valid Nakama user session and brand ownership |
| Tool | `jazebe_request_planet_publish` | Requests review when requirements pass | Valid Nakama user session and brand ownership |
| Tool | `jazebe_withdraw_planet_publish` | Withdraws review request | Valid Nakama user session and brand ownership |
| Tool | `jazebe_create_media_mission_draft` | Creates a caller-owned media mission draft | Valid Nakama user session and brand ownership |
| Tool | `jazebe_list_brand_team` | Lists brand administrators | Valid Nakama session and active main brand role |
| Tool | `jazebe_join_planet` | Joins an active planet once | Valid Nakama user session |
| Tool | `jazebe_list_my_planets` | Lists caller's active planet memberships | Valid Nakama user session |
| Resource | `nakama://me/account` | Same account as JSON | Valid Nakama user session |
| Prompt | `inspect_my_account` | Asks the client to summarize the account resource | Valid Nakama user session |

The account tool has an empty input schema and typed structured output. The Jazebe tools have typed input schemas and forward only to fixed RPC names; the business rules stay in Nakama's native module. Account access is user scoped: a caller cannot select another account ID. Backend errors are returned without echoing session tokens. There are no administrator tools.

A typical local MCP client command configuration, after building the binary with `go build -o nakama-mcp.exe .`, is:

```json
{
  "mcpServers": {
    "nakama-farsi": {
      "command": "C:\\path\\to\\nakama-mcp.exe",
      "args": ["stdio"],
      "env": {
        "NAKAMA_URL": "http://127.0.0.1:7350",
        "NAKAMA_SESSION_TOKEN": "<Nakama user session>"
      }
    }
  }
}
```

`nakama_get_account` takes `{}` and returns structured content shaped like `{"account":{"user":{...},"wallet":"{...}"}}`; the exact account fields come from the running Nakama server. Session expiry produces an error. Do not put a real session token in a committed client configuration file.

## Security and limits

Nakama user sessions are accepted directly here. This is appropriate for a local trusted client or an internal reverse proxy that already manages client authentication. This module does **not** implement MCP OAuth protected resource metadata, an OAuth authorization server, scopes, or administrator role mapping. Therefore the HTTP listener is deliberately restricted to loopback and must not be published as a remote production MCP endpoint. A production remote integration requires a compliant OAuth flow, token audience checks, per capability scopes, audit logs, and deployment tests before opening access. The HTTP handler limits concurrency to 32 requests; it has no distributed rate limiter because remote use is disabled.

## Adding a capability

Add a typed handler to `newServer` that calls an existing authenticated Nakama or Jazebe service through a narrow adapter. Keep the user identity in the request context or a verified token; never accept arbitrary target user IDs without an authorization check. Add a matching SDK client test for discovery and invocation, plus tests for invalid input and access denial. Update this catalog with the tool's exact scope and side effects.
