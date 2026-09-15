## Security Impact & Threat Model
<!-- Describe the vulnerability, threat vector (SSRF, DoS, auth bypass, injection, secrets exposure), or hardening requirement. -->

## Hardening / Mitigation Implemented
<!-- Detail the defensive measures applied (e.g. SafeDialContext, MaxBytesReader, bounds clamping, security headers, token strength checks). -->

## Security Invariants Audited
- [ ] **SSRF Defense**: Outgoing network requests use `netutil.SafeHTTPClient` or `netutil.SafeDialContext`.
- [ ] **DoS & Memory Protection**: Request bodies are capped with `http.MaxBytesReader` and pagination bounds clamped.
- [ ] **Secrets & Auth Protection**: No secrets, tokens, or credential files committed or leaked.
- [ ] **Security Headers**: Standard headers (COOP, CORP, HSTS, X-Content-Type-Options) verified.

## Mandatory Zero-Failure Pre-Commit Verification (AGENTS.md)
*All commands below must be executed locally and pass with Exit Code 0 prior to creating the PR:*

### Backend Suite (if backend code changed)
- [ ] `gofmt -w . && gofmt -l .` (Output must be completely empty)
- [ ] `go vet ./...` (Static analysis clean)
- [ ] `go test -count=1 -v ./...` (All tests pass with cache disabled)

### Frontend Suite (if frontend code changed)
- [ ] `npm run format && npm run lint` (Prettier code style & ESLint clean)
- [ ] `npm run check` (SvelteKit TypeScript validation clean)
- [ ] `npm run build` (Production build succeeds with static adapter)

## Local Verification Evidence
<!-- Provide the actual terminal execution output summary showing 0 errors across test/lint suites. -->
```text
<!-- Paste local test/build verification output here -->
```
