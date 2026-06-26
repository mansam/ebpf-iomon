package device

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"
)

var kubeletVolumeRe = regexp.MustCompile(
	`.*/pods/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})/volumes/[^/]+/([^/]+)(?:/mount)?$`,
)

type DeviceInfo struct {
	PVName    string
	PodUID    string
	NodeName  string
	MountPath string
	DevPath   string
	IsNFS     bool
}

type Resolver struct {
	mu       sync.RWMutex
	devices  map[uint32]DeviceInfo
	nodeName string
	interval time.Duration
	procPath string
	log      *slog.Logger
}

func NewResolver(nodeName, procPath string, interval time.Duration, log *slog.Logger) *Resolver {
	return &Resolver{
		devices:  make(map[uint32]DeviceInfo),
		nodeName: nodeName,
		interval: interval,
		procPath: procPath,
		log:      log,
	}
}

func (r *Resolver) Lookup(dev uint32) (DeviceInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.devices[dev]
	return info, ok
}

func (r *Resolver) Run(ctx context.Context) {
	r.scan()
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.scan()
		}
	}
}

func (r *Resolver) scan() {
	mountInfoPath := fmt.Sprintf("%s/1/mountinfo", r.procPath)
	mounts, err := ParseMountInfo(mountInfoPath)
	if err != nil {
		r.log.Error("parsing mountinfo", "path", mountInfoPath, "error", err)
		return
	}

	devices := make(map[uint32]DeviceInfo)

	for _, m := range mounts {
		podUID, pvName, ok := parseKubeletVolumePath(m.MountPoint)
		if !ok {
			continue
		}

		dev := MkDev(m.Major, m.Minor)
		info := DeviceInfo{
			PVName:    pvName,
			PodUID:    podUID,
			NodeName:  r.nodeName,
			MountPath: m.MountPoint,
			DevPath:   m.Source,
			IsNFS:     m.FSType == "nfs" || m.FSType == "nfs4",
		}

		devices[dev] = info
		r.log.Debug("resolved kubelet volume",
			"dev", dev,
			"pv", pvName,
			"podUID", podUID,
			"source", m.Source,
			"fsType", m.FSType,
		)
	}

	r.mu.Lock()
	r.devices = devices
	r.mu.Unlock()

	r.log.Debug("device scan complete", "resolved", len(devices))
}

func parseKubeletVolumePath(mountPoint string) (podUID, pvName string, ok bool) {
	matches := kubeletVolumeRe.FindStringSubmatch(mountPoint)
	if matches == nil {
		return "", "", false
	}
	return matches[1], matches[2], true
}

func MkDev(major, minor uint32) uint32 {
	return (major << 20) | minor
}

func DevToString(dev uint32) string {
	major := dev >> 20
	minor := dev & ((1 << 20) - 1)
	return fmt.Sprintf("%d:%d", major, minor)
}
