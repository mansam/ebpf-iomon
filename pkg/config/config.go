package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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
}

func Parse() *Config {
	c := &Config{}
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
	flag.Parse()

	applyEnvOverrides(c)

	return c
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
