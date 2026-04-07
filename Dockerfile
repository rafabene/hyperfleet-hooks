ARG BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:latest

FROM registry.access.redhat.com/ubi9/go-toolset:1.25 AS builder

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

USER root
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -X github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/version.Version=${VERSION} \
    -X github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/version.GitCommit=${GIT_COMMIT} \
    -X github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/version.BuildDate=${BUILD_DATE}" \
    -o hyperfleet-hooks \
    ./cmd/hyperfleet-hooks

# Runtime stage
FROM ${BASE_IMAGE}

# Install git-core for commit validation
RUN microdnf install -y git-core \
    && microdnf clean all \
    && rm -rf /var/cache/yum

WORKDIR /workspace

COPY --from=builder /build/hyperfleet-hooks /usr/local/bin/hyperfleet-hooks

USER 65532:65532

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD hyperfleet-hooks version || exit 1

ENTRYPOINT ["/usr/local/bin/hyperfleet-hooks"]
CMD ["--help"]

ARG VERSION=dev
LABEL name="hyperfleet-hooks" \
      vendor="Red Hat" \
      version="${VERSION}" \
      summary="HyperFleet Hooks - Commit Message Validator" \
      description="Validates commit messages against HyperFleet standards"
