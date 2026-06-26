package metrics

import "math"

const MaxSlots = 26

var log2BucketsSec [MaxSlots]float64

func init() {
	for i := 0; i < MaxSlots; i++ {
		log2BucketsSec[i] = float64(uint64(1)<<uint(i)) / 1_000_000.0
	}
}

func SlotsToConstHistogram(slots [MaxSlots]uint64) (count uint64, sum float64, buckets map[float64]uint64) {
	buckets = make(map[float64]uint64, MaxSlots)
	var cumulative uint64

	for i := 0; i < MaxSlots; i++ {
		cumulative += slots[i]
		count += slots[i]

		var midpointUS float64
		if i == 0 {
			midpointUS = 0.5
		} else {
			midpointUS = float64(uint64(1)<<uint(i-1)) * math.Sqrt2
		}
		sum += float64(slots[i]) * (midpointUS / 1_000_000.0)

		buckets[log2BucketsSec[i]] = cumulative
	}
	return count, sum, buckets
}
