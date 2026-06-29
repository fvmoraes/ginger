# Optional telemetry module

This directory is a separate Go module so the Ginger core can remain compatible
with Go 1.22. Applications opt into it explicitly:

```bash
go get github.com/fvmoraes/ginger/pkg/telemetry
```

The module currently requires Go 1.25 because its OpenTelemetry dependencies do.
