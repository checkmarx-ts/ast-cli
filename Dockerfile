FROM cgr.dev/chainguard/bash:latest@sha256:ef75f390f366847bc158243206eda32a4ff616125251f5d6d7cd73b1d3e2362e

USER nonroot

COPY cx /app/bin/cx

ENTRYPOINT ["/app/bin/cx"]
