package config

import (
	"os"
	"testing"
)

func TestApplyEnvOverrides(t *testing.T) {
	c := &Config{
		MetricsPort:  9090,
		ScanInterval: 30,
		EnableBlock:  true,
		EnableNFS:    true,
		LogLevel:     "info",
	}

	t.Setenv("OCP_EBPF_IOMON_METRICS_PORT", "8080")
	t.Setenv("OCP_EBPF_IOMON_SCAN_INTERVAL", "15")
	t.Setenv("OCP_EBPF_IOMON_ENABLE_NFS", "false")
	t.Setenv("OCP_EBPF_IOMON_LOG_LEVEL", "debug")

	applyEnvOverrides(c)

	if c.MetricsPort != 8080 {
		t.Errorf("MetricsPort = %d, want 8080", c.MetricsPort)
	}
	if c.ScanInterval != 15 {
		t.Errorf("ScanInterval = %d, want 15", c.ScanInterval)
	}
	if c.EnableNFS {
		t.Error("EnableNFS should be false")
	}
	if c.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", c.LogLevel, "debug")
	}
}

func TestNodeNameFromEnv(t *testing.T) {
	c := &Config{}
	t.Setenv("NODE_NAME", "worker-1")
	applyEnvOverrides(c)

	if c.NodeName != "worker-1" {
		t.Errorf("NodeName = %q, want %q", c.NodeName, "worker-1")
	}
}

func TestNodeNameFlagTakesPrecedence(t *testing.T) {
	c := &Config{NodeName: "from-flag"}
	os.Setenv("NODE_NAME", "from-env")
	defer os.Unsetenv("NODE_NAME")
	applyEnvOverrides(c)

	if c.NodeName != "from-flag" {
		t.Errorf("NodeName = %q, want %q (flag should take precedence)", c.NodeName, "from-flag")
	}
}

func TestValidate(t *testing.T) {
	valid := &Config{
		NodeName:         "worker-1",
		MetricsPort:      9090,
		ScanInterval:     30,
		BlockMapSize:     10240,
		NFSMapSize:       10240,
		NFSKprobeMapSize: 10240,
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid config: unexpected error: %v", err)
	}

	tests := []struct {
		name   string
		modify func(*Config)
	}{
		{"empty node name", func(c *Config) { c.NodeName = "" }},
		{"zero scan interval", func(c *Config) { c.ScanInterval = 0 }},
		{"negative scan interval", func(c *Config) { c.ScanInterval = -1 }},
		{"zero port", func(c *Config) { c.MetricsPort = 0 }},
		{"port too high", func(c *Config) { c.MetricsPort = 70000 }},
		{"negative block map size", func(c *Config) { c.BlockMapSize = -1 }},
		{"negative nfs map size", func(c *Config) { c.NFSMapSize = -1 }},
		{"negative nfs kprobe map size", func(c *Config) { c.NFSKprobeMapSize = -1 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := *valid
			tt.modify(&c)
			if err := c.Validate(); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}
