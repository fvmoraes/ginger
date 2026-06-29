# Existing API safety fixture

This is deliberately not a Ginger-generated project. It uses `net/http`, its
own `internal/httpapi`, `internal/core`, and `internal/store` layout, plus an
existing test that generation must preserve.

Copy this directory to a temporary location before running the mutation flow:

```bash
cp -R examples/existing-api /tmp/ginger-existing-api
go build -o /tmp/ginger ./cmd/ginger
cd /tmp/ginger-existing-api
/tmp/ginger init
/tmp/ginger inspect
/tmp/ginger add swagger --plan
/tmp/ginger generate tests --scan --plan
/tmp/ginger doctor
```

Swagger planning creates new integration/docs files and a reviewable patch for
the existing `net/http` router; it does not rewrite `router.go`.
