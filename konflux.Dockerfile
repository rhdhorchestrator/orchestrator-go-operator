# Build the manager binary
#FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.23 as builder
FROM registry.access.redhat.com/ubi9/go-toolset:1.23.9-1751375493 as builder

ARG TARGETARCH
ENV GOEXPERIMENT=strictfipsruntime

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
#COPY vendor/ vendor/

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/
COPY LICENSE /licenses/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use redhat ubi9 as minimal base image to package the manager binary
# https://registry.access.redhat.com/ubi9/ubi-minimal
FROM registry.access.redhat.com/ubi9-minimal:9.6-1749489516
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

LABEL com.redhat.component="RHDH Orchestrator Operator"
LABEL distribution-scope="public"
LABEL name="rhdh-orchestrator-operator-bundle"
LABEL release="1.6.0"
LABEL version="1.6.0"
LABEL maintainer="Red Hat jubah@redhat.com"
LABEL url="https://github.com/rhdhorchestrator/orchestrator-go-operator"
LABEL vendor="Red Hat, Inc."
LABEL description="RHDH Orchestrator introduces serverless asynchronous workflows to Backstage, \
				  with a focus on facilitating the transition of applications to the cloud, \
				  onboarding developers, and enabling users to create workflows for backstage \
				  actions or external systems."
LABEL io.k8s.description="RHDH Orchestrator introduces serverless asynchronous workflows to Backstage, \
				  with a focus on facilitating the transition of applications to the cloud, \
				  onboarding developers, and enabling users to create workflows for backstage \
				  actions or external systems."
LABEL summary="RHDH Orchestrator introduces serverless asynchronous workflows to Backstage, \
				  with a focus on facilitating the transition of applications to the cloud, \
				  onboarding developers, and enabling users to create workflows for backstage \
				  actions or external systems."
LABEL io.k8s.display-name="RHDH Orchestrator Operator"
LABEL io.openshift.tags="openshift,operator,rhdh,orchestrator"

ENTRYPOINT ["/manager"]
