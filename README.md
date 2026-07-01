# ebpf-iomon

eBPF-based I/O latency monitoring agent for OpenShift 4.20+ clusters. Runs as a DaemonSet, attaches to kernel tracepoints and kprobes to measure block and NFS I/O latency, and exposes Prometheus histogram metrics. A Perses dashboard is included.

## Requirements

- OpenShift 4.20+ (RHCOS kernel with BTF support)
- Cluster Observability Operator (for Perses dashboards)
- `BPF`, `PERFMON`, `SYS_RESOURCE` capabilities (handled by the included SCC)

## Deploy

```sh
make image push IMAGE_REPO=quay.io/yourorg/ebpf-iomon IMAGE_TAG=latest
```

Edit `deploy/daemonset.yaml` to set your image, then:

```sh
oc apply -k deploy/
```

This creates the `ebpf-iomon` namespace and deploys the DaemonSet, ServiceMonitor, SCC, and Perses dashboard.

To remove:

```sh
oc delete -k deploy/
```

## Subsystems

| Subsystem | Flag | Default | What it traces |
|---|---|---|---|
| Block | `--enable-block` | `true` | Block I/O latency via `block_io_start`/`block_io_done` tracepoints |
| NFS | `--enable-nfs` | `true` | NFS read/write latency via `nfs_initiate_read`/`nfs_readpage_done` tracepoints |
| NFS Kprobe | `--enable-nfs-kprobe` | `false` | NFS VFS calls (read/write/open/getattr) via kprobes |

## Metrics

All histograms use configurable bucket boundaries (see [Configuration](#configuration)).

| Metric | Labels | Description |
|---|---|---|
| `ebpf_block_io_latency_seconds` | `node`, `persistentvolume`, `pod_uid`, `operation` | Block I/O latency for pod volumes |
| `ebpf_system_block_io_latency_seconds` | `node`, `device`, `operation` | Block I/O latency for system/unresolvable devices |
| `ebpf_nfs_io_latency_seconds` | `node`, `persistentvolume`, `pod_uid`, `operation` | NFS I/O latency (tracepoint) |
| `ebpf_nfs_vfs_latency_seconds` | `node`, `persistentvolume`, `pod_uid`, `operation` | NFS VFS call latency (kprobe) |
| `ebpf_iomon_subsystem_active` | `subsystem` | Whether each subsystem loaded successfully (1/0) |

## Configuration

All flags can be overridden by environment variables with the `OCP_EBPF_IOMON_` prefix.

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--metrics-port` | `OCP_EBPF_IOMON_METRICS_PORT` | `9090` | Prometheus metrics port |
| `--scan-interval` | `OCP_EBPF_IOMON_SCAN_INTERVAL` | `30` | Device-to-pod resolution scan interval (seconds) |
| `--enable-block` | `OCP_EBPF_IOMON_ENABLE_BLOCK` | `true` | Enable block I/O tracing |
| `--enable-nfs` | `OCP_EBPF_IOMON_ENABLE_NFS` | `true` | Enable NFS tracepoint tracing |
| `--enable-nfs-kprobe` | `OCP_EBPF_IOMON_ENABLE_NFS_KPROBE` | `false` | Enable NFS kprobe tracing |
| `--boundaries` | `OCP_EBPF_IOMON_BOUNDARIES` | `10000000,100000000,1000000000` | Histogram bucket boundaries in nanoseconds (comma-separated) |
| `--log-level` | `OCP_EBPF_IOMON_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `--node-name` | `NODE_NAME` | (required) | Node name, typically set via downward API |
| `--block-map-size` | | `10240` | Max entries for block start timestamp map |
| `--nfs-map-size` | | `10240` | Max entries for NFS start timestamp map |
| `--nfs-kprobe-map-size` | | `10240` | Max entries for NFS kprobe start timestamp map |

The default boundaries correspond to 10ms, 100ms, and 1s. Prometheus adds +Inf automatically. The boundary format matches [kubevirt-storage-latency-exporter](https://github.com/mhenriks/kubevirt-storage-latency-exporter) for consistency.

## Build

```sh
make build          # requires go generate (needs clang, bpftool)
make test           # run unit tests
make image          # build container image with podman
make lint           # golangci-lint
```

## Dashboard

The Perses dashboard (`deploy/dashboard.yaml`) is deployed automatically via kustomize. It requires the Cluster Observability Operator and a `PrometheusDatasource` named `cluster-monitoring`.

Panels:
- Subsystem status (block, NFS, NFS kprobe)
- Block I/O latency quantiles (pod volumes)
- Block I/O latency quantiles (system devices)
- NFS I/O latency quantiles (tracepoint)
- NFS VFS latency quantiles (kprobe)

Variables filter by node, persistent volume, pod UID, and operation.
