# Security and resilience audit

Scope: source review of upstream v0.3.0 (`ad44dbb`) and local hardening of the HTTP client and EPG codecs. No provider accounts, credentials, or live endpoints were used. This is a focused audit, not a comprehensive security certification.

## Findings and changes

- **High — credentials in errors.** Request failures wrapped `net/url` errors containing the credential-bearing URL. Decoder failures could retain arbitrary provider values. Client-created errors now retain only safe sentinel causes or a generic diagnostic, never the original transport/decoder error. `errors.Is` remains useful for cancellation, deadlines, unexpected EOF and the new `ErrResponseTooLarge`. HTTP error bodies are no longer read, retained or printed. The exported `HTTPError.Body` field remains for source compatibility but is not populated by this client.
- **High — redirect disclosure.** Automatic redirects could send credential-bearing query parameters or Referer headers to another origin. Redirects are now restricted to the initial scheme and authority, with a ten-hop limit; Referer is removed. Caller redirect callbacks are retained, but cannot permit a cross-origin redirect. The supplied HTTP client is copied rather than modified.
- **Medium — unbounded response consumption.** Error bodies were read without a limit; successful JSON, M3U and XML bodies were also unbounded. Error bodies are closed without reading. Success responses default to a 64 MiB decompressed-byte ceiling, configurable through `WithMaxResponseBytes`. Oversized declared lengths are rejected early; streamed/chunked bytes are checked with an explicit, sticky size error rather than a synthetic EOF. All successful parsers finish consuming the bounded stream; JSON additionally rejects trailing values.
- **Medium — host concatenation.** A host containing a query or fragment could misroute or truncate appended parameters. Hosts are now parsed as HTTP(S) URLs and reject userinfo, query strings, fragments and missing authorities. A base directory is supported and trailing slashes are normalized. Invalid hosts fail at request time to preserve the existing constructor signature.
- **Low — fragile EPG text and null times.** Invalid base64 in a single title or description previously rejected the guide. Plain text and unpadded base64 are now accepted, with invalid UTF-8 decoded content preserved as original text. Empty/null Unix timestamps now represent Go's zero time, not a decode failure or Unix epoch. Other malformed numeric values still fail rather than silently inventing times.
- **Low — nil options.** Nil options and a nil `WithHTTPClient` argument no longer panic. A context is checked at each bounded body read in addition to the HTTP request's existing context.

## Compatibility and integration

The module remains `github.com/sherif-fanous/xtreamcodes`; existing imports, constructors and methods are unchanged. Use a Go module `replace` to select this fork. New API: `WithMaxResponseBytes(int64)` and `ErrResponseTooLarge`. Nonpositive size options retain the default.

Intentional behavioral changes: error text no longer includes server bodies or detailed underlying errors; `errors.As` into raw transport/decoder errors is not supported for client-created errors. HTTP status and redacted request URL remain available. Cross-origin redirects, including HTTP-to-HTTPS redirects, are blocked: configure the final HTTPS host directly. Large XMLTV feeds may need a higher explicit byte limit.

## Verification

- Existing `go test ./...`: passed.
- Existing `go test -race ./...`: passed.
- `go vet ./...`: passed.
- `go build ./...`: passed.
- `git diff --check`: passed.

No new tests were added under the task's explicit constraint. Existing tests exercise local HTTP fixtures and serialization round trips; they do not provide dedicated regression coverage of every new security boundary. Redirect/size/privacy guarantees above are based on implementation review, not a live-provider test.

## Remaining limitations

- HTTP is still supported for compatibility and exposes credentials in transit; callers should configure HTTPS. Query credentials are intrinsic to the provider protocol and are visible to the configured provider, proxies and caller-supplied transport hooks.
- This is not an SSRF sandbox. Callers accepting untrusted host input must enforce destination/network policy. A custom transport or proxy remains trusted.
- The byte ceiling limits response input, not total decoded-object memory or CPU. Concurrent calls multiply resource use. Custom HTTP clients retain their configured timeouts; callers should also use context deadlines.
- Plain text that also happens to be valid base64 yielding valid UTF-8 is inherently ambiguous. Numeric and structural provider inconsistencies beyond the documented codecs can still fail decoding.
- Public authentication models contain provider-returned credentials by design. Applications must not log whole authentication models, response objects or client internals. Direct calls to model `json.Unmarshal` are outside the HTTP client's error-sanitization boundary.
- Third-party parser dependencies were not comprehensively audited, and no dependency-vulnerability database scan or fuzzing was performed.
