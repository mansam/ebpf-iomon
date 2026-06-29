package metrics

import (
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
	if len(buckets) != len(promBuckets) {
		t.Errorf("len(buckets) = %d, want %d", len(buckets), len(promBuckets))
	}
	for _, v := range buckets {
		if v != 0 {
			t.Errorf("expected all bucket counts to be 0, got %d", v)
		}
	}
}

func TestSlotsToConstHistogramFastOps(t *testing.T) {
	var slots [MaxSlots]uint64
	slots[5] = 100 // slot 5 = [32µs, 64µs) — well below 100ms

	count, sum, buckets := SlotsToConstHistogram(slots)

	if count != 100 {
		t.Errorf("count = %d, want 100", count)
	}
	if sum <= 0 {
		t.Error("sum should be positive")
	}

	// All 100 ops are sub-100ms, so every bucket should contain 100
	for _, b := range promBuckets {
		if buckets[b] != 100 {
			t.Errorf("bucket le=%.2f = %d, want 100", b, buckets[b])
		}
	}
}

func TestSlotsToConstHistogramSlowOps(t *testing.T) {
	var slots [MaxSlots]uint64
	// Slot 21: [2^21 µs, 2^22 µs) = [2.1s, 4.2s)
	slots[21] = 50

	count, _, buckets := SlotsToConstHistogram(slots)

	if count != 50 {
		t.Errorf("count = %d, want 50", count)
	}

	// slotCeiling[21] = 2^21/1e6 ≈ 2.097s, which is ≤ 2.5s
	// So buckets ≤ 2.5s should include this slot
	for _, b := range promBuckets {
		if b < 2.5 {
			if buckets[b] != 0 {
				t.Errorf("bucket le=%.2f = %d, want 0 (slot 21 is above this)", b, buckets[b])
			}
		} else {
			if buckets[b] != 50 {
				t.Errorf("bucket le=%.2f = %d, want 50", b, buckets[b])
			}
		}
	}
}

func TestSlotsToConstHistogramCumulative(t *testing.T) {
	var slots [MaxSlots]uint64
	slots[5] = 10  // ~32µs — below 100ms
	slots[19] = 20 // ~524ms — slotCeiling[19] ≈ 0.524s, ≤ 1s
	slots[23] = 30 // ~8.4s — slotCeiling[23] ≈ 8.39s, ≤ 10s

	count, _, buckets := SlotsToConstHistogram(slots)

	if count != 60 {
		t.Errorf("count = %d, want 60", count)
	}

	expects := map[float64]uint64{
		0.1:  10, // slot 5 only
		0.25: 10,
		0.5:  10,
		1:    30, // slots 5 + 19
		2.5:  30,
		5:    30,
		10:   60, // slots 5 + 19 + 23
		30:   60,
		60:   60,
	}

	for b, want := range expects {
		if buckets[b] != want {
			t.Errorf("bucket le=%.2f = %d, want %d", b, buckets[b], want)
		}
	}
}

func TestSlotsToConstHistogramBeyond60s(t *testing.T) {
	var slots [MaxSlots]uint64
	// Slot 25: [2^25 µs, ...) = [33.5s, ...) — beyond 30s bucket
	slots[25] = 5
	slots[10] = 10 // ~1ms, well below 100ms

	count, _, buckets := SlotsToConstHistogram(slots)

	if count != 15 {
		t.Errorf("count = %d, want 15", count)
	}

	// Slot 25 ceiling is ~33.5s which is ≤ 60s, so le=60 should include it
	if buckets[30] != 10 {
		t.Errorf("bucket le=30 = %d, want 10 (should not include slot 25)", buckets[30])
	}
	if buckets[60] != 15 {
		t.Errorf("bucket le=60 = %d, want 15 (should include slot 25)", buckets[60])
	}
}

func TestPromBucketBoundaries(t *testing.T) {
	expected := []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60}
	if len(promBuckets) != len(expected) {
		t.Fatalf("len(promBuckets) = %d, want %d", len(promBuckets), len(expected))
	}
	for i, b := range promBuckets {
		if b != expected[i] {
			t.Errorf("promBuckets[%d] = %f, want %f", i, b, expected[i])
		}
	}
}
