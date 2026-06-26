package metrics

import (
	"math"
	"testing"
)

func TestSlotsToConstHistogramEmpty(t *testing.T) {
	var slots [MaxSlots]uint64
	count, sum, buckets := SlotsToConstHistogram(slots)

	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
	if sum != 0 {
		t.Errorf("sum = %f, want 0", sum)
	}
	if len(buckets) != MaxSlots {
		t.Errorf("len(buckets) = %d, want %d", len(buckets), MaxSlots)
	}
	for _, v := range buckets {
		if v != 0 {
			t.Errorf("expected all bucket counts to be 0, got %d", v)
		}
	}
}

func TestSlotsToConstHistogramSingleBucket(t *testing.T) {
	var slots [MaxSlots]uint64
	slots[10] = 100 // bucket 10 = [512us, 1024us)

	count, sum, buckets := SlotsToConstHistogram(slots)

	if count != 100 {
		t.Errorf("count = %d, want 100", count)
	}
	if sum <= 0 {
		t.Error("sum should be positive")
	}

	// Buckets 0-9 should be 0 (cumulative), buckets 10-25 should be 100
	for i := 0; i < MaxSlots; i++ {
		boundary := log2BucketsSec[i]
		expected := uint64(0)
		if i >= 10 {
			expected = 100
		}
		if buckets[boundary] != expected {
			t.Errorf("bucket[%d] (%.6f s) = %d, want %d", i, boundary, buckets[boundary], expected)
		}
	}
}

func TestSlotsToConstHistogramCumulative(t *testing.T) {
	var slots [MaxSlots]uint64
	slots[0] = 10
	slots[5] = 20
	slots[10] = 30

	count, _, buckets := SlotsToConstHistogram(slots)

	if count != 60 {
		t.Errorf("count = %d, want 60", count)
	}

	// bucket 0: 10, bucket 5: 30, bucket 10: 60
	if buckets[log2BucketsSec[0]] != 10 {
		t.Errorf("cumulative at bucket 0 = %d, want 10", buckets[log2BucketsSec[0]])
	}
	if buckets[log2BucketsSec[5]] != 30 {
		t.Errorf("cumulative at bucket 5 = %d, want 30", buckets[log2BucketsSec[5]])
	}
	if buckets[log2BucketsSec[10]] != 60 {
		t.Errorf("cumulative at bucket 10 = %d, want 60", buckets[log2BucketsSec[10]])
	}
}

func TestLog2BucketBoundaries(t *testing.T) {
	if log2BucketsSec[0] != 1e-6 {
		t.Errorf("bucket 0 boundary = %e, want 1e-6", log2BucketsSec[0])
	}
	if log2BucketsSec[10] != 1024e-6 {
		t.Errorf("bucket 10 boundary = %e, want 1024e-6", log2BucketsSec[10])
	}
	if log2BucketsSec[20] != 1048576e-6 {
		t.Errorf("bucket 20 boundary = %e, want ~1.048576 s", log2BucketsSec[20])
	}
	if math.Abs(log2BucketsSec[25]-33.554432) > 1e-6 {
		t.Errorf("bucket 25 boundary = %f, want ~33.554432", log2BucketsSec[25])
	}
}
