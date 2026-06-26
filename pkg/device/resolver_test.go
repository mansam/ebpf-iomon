package device

import "testing"

func TestParseKubeletVolumePath(t *testing.T) {
	tests := []struct {
		name       string
		mountPoint string
		wantUID    string
		wantPV     string
		wantOK     bool
	}{
		{
			name:       "NFS volume",
			mountPoint: "/var/lib/kubelet/pods/6a76960a-a927-4211-96e6-1f187b126e90/volumes/kubernetes.io~nfs/example",
			wantUID:    "6a76960a-a927-4211-96e6-1f187b126e90",
			wantPV:     "example",
			wantOK:     true,
		},
		{
			name:       "CSI volume with /mount suffix",
			mountPoint: "/var/lib/kubelet/pods/aabbccdd-1234-5678-9abc-def012345678/volumes/kubernetes.io~csi/pvc-deadbeef-0000-1111-2222-333344445555/mount",
			wantUID:    "aabbccdd-1234-5678-9abc-def012345678",
			wantPV:     "pvc-deadbeef-0000-1111-2222-333344445555",
			wantOK:     true,
		},
		{
			name:       "local volume",
			mountPoint: "/var/lib/kubelet/pods/11111111-2222-3333-4444-555566667777/volumes/kubernetes.io~local-volume/local-pv-abc",
			wantUID:    "11111111-2222-3333-4444-555566667777",
			wantPV:     "local-pv-abc",
			wantOK:     true,
		},
		{
			name:       "non-kubelet mount",
			mountPoint: "/mnt/data",
			wantOK:     false,
		},
		{
			name:       "root filesystem",
			mountPoint: "/",
			wantOK:     false,
		},
		{
			name:       "kubelet path without volume",
			mountPoint: "/var/lib/kubelet/pods/6a76960a-a927-4211-96e6-1f187b126e90/containers/app",
			wantOK:     false,
		},
		{
			name:       "invalid UUID in path",
			mountPoint: "/var/lib/kubelet/pods/not-a-uuid/volumes/kubernetes.io~nfs/example",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, pv, ok := parseKubeletVolumePath(tt.mountPoint)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if uid != tt.wantUID {
				t.Errorf("podUID = %q, want %q", uid, tt.wantUID)
			}
			if pv != tt.wantPV {
				t.Errorf("pvName = %q, want %q", pv, tt.wantPV)
			}
		})
	}
}
