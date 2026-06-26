FROM quay.io/centos/centos:stream9 AS builder

RUN dnf install -y --enablerepo=crb \
    clang \
    llvm \
    libbpf-devel \
    elfutils-libelf-devel \
    zlib-devel \
    make \
    gcc \
    golang \
    bpftool \
    && dnf clean all

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate ./pkg/ebpf/...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /ebpf-iomon ./cmd/ebpf-iomon/

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

COPY --from=builder /ebpf-iomon /usr/local/bin/ebpf-iomon

LABEL name="ebpf-iomon" \
      summary="eBPF-based I/O monitoring agent for OpenShift" \
      description="Monitors block and NFS I/O latency using eBPF tracepoints" \
      io.k8s.display-name="OCP eBPF I/O Monitor" \
      io.openshift.tags="ebpf,monitoring,io"

USER 0
ENTRYPOINT ["/usr/local/bin/ebpf-iomon"]
