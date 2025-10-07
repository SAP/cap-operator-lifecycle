# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS builder
ARG TARGETOS TARGETARCH

WORKDIR /workspace
# Copy the go module manifests
COPY go.mod go.sum .
# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go sources
COPY main.go main.go
COPY api/ api/
COPY internal/ internal/
COPY pkg/ pkg/

# Build
ENV CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH}
RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY chart/ chart/
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
