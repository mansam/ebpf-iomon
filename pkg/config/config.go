package config

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	MetricsPort      int
	ScanInterval     int
	EnableBlock      bool
	EnableNFS        bool
	EnableNFSKprobe  bool
	LogLevel         string
	BlockMapSize     int
	NFSMapSize       int
	NFSKprobeMapSize int
	NodeName         string
	ProcPath         string
	Boundaries       []float64
}

func Parse() *Config {
	c := &Config{}
	var boundariesStr string
	flag.IntVar(&c.MetricsPort, "metrics-port", 9090, "Port for Prometheus metrics endpoint")
	flag.IntVar(&c.ScanInterval, "scan-interval", 30, "Device scan interval in seconds")
	flag.BoolVar(&c.EnableBlock, "enable-block", true, "Enable block I/O tracing")
	flag.BoolVar(&c.EnableNFS, "enable-nfs", true, "Enable NFS I/O tracing (tracepoint-based)")
	flag.BoolVar(&c.EnableNFSKprobe, "enable-nfs-kprobe", false, "Enable NFS VFS latency tracing (kprobe-based, covers open/getattr)")
	flag.StringVar(&c.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.IntVar(&c.BlockMapSize, "block-map-size", 10240, "Max entries for block start timestamp map")
	flag.IntVar(&c.NFSMapSize, "nfs-map-size", 10240, "Max entries for NFS start timestamp map")
	flag.IntVar(&c.NFSKprobeMapSize, "nfs-kprobe-map-size", 10240, "Max entries for NFS kprobe start timestamp map")
	flag.StringVar(&c.NodeName, "node-name", "", "Node name (auto-detected from env NODE_NAME)")
	flag.StringVar(&c.ProcPath, "proc-path", "/proc", "Path to host proc filesystem")
	flag.StringVar(&boundariesStr, "boundaries", "10000000,100000000,1000000000", "Histogram bucket boundaries in nanoseconds (comma-separated)")
	flag.Parse()

	applyEnvOverrides(c)

	if v := os.Getenv("OCP_EBPF_IOMON_BOUNDARIES"); v != "" {
		boundariesStr = v
	}
	c.Boundaries = parseBoundaries(boundariesStr)

	return c
}

func parseBoundaries(s string) []float64 {
	parts := strings.Split(s, ",")
	buckets := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		ns, err := strconv.ParseFloat(p, 64)
		if err != nil {
			continue
		}
		buckets = append(buckets, ns/1e9)
	}
	sort.Float64s(buckets)
	return buckets
}

func (c *Config) Validate() error {
	if c.NodeName == "" {
		return fmt.Errorf("node name is required: set --node-name or NODE_NAME env var")
	}
	if c.ScanInterval <= 0 {
		return fmt.Errorf("scan-interval must be positive, got %d", c.ScanInterval)
	}
	if c.MetricsPort <= 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("metrics-port must be 1-65535, got %d", c.MetricsPort)
	}
	if c.BlockMapSize <= 0 {
		return fmt.Errorf("block-map-size must be positive, got %d", c.BlockMapSize)
	}
	if c.NFSMapSize <= 0 {
		return fmt.Errorf("nfs-map-size must be positive, got %d", c.NFSMapSize)
	}
	if c.NFSKprobeMapSize <= 0 {
		return fmt.Errorf("nfs-kprobe-map-size must be positive, got %d", c.NFSKprobeMapSize)
	}
	if len(c.Boundaries) == 0 {
		return fmt.Errorf("boundaries must contain at least one value")
	}
	for i, b := range c.Boundaries {
		if b <= 0 {
			return fmt.Errorf("boundary values must be positive, got %g", b)
		}
		if i > 0 && b <= c.Boundaries[i-1] {
			return fmt.Errorf("boundary values must be in ascending order")
		}
	}
	return nil
}

func applyEnvOverrides(c *Config) {
	if v := os.Getenv("OCP_EBPF_IOMON_METRICS_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.MetricsPort = n
		}
	}
	if v := os.Getenv("OCP_EBPF_IOMON_SCAN_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ScanInterval = n
		}
	}
	if v := os.Getenv("OCP_EBPF_IOMON_ENABLE_BLOCK"); v != "" {
		c.EnableBlock = v == "true" || v == "1"
	}
	if v := os.Getenv("OCP_EBPF_IOMON_ENABLE_NFS"); v != "" {
		c.EnableNFS = v == "true" || v == "1"
	}
	if v := os.Getenv("OCP_EBPF_IOMON_ENABLE_NFS_KPROBE"); v != "" {
		c.EnableNFSKprobe = v == "true" || v == "1"
	}
	if v := os.Getenv("OCP_EBPF_IOMON_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("NODE_NAME"); v != "" && c.NodeName == "" {
		c.NodeName = v
	}
}
