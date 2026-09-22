package stats

import (
	"math"
	"sort"
	"time"
)

type Distribution struct {
	Count int     `json:"count"`
	MinMS float64 `json:"min_ms"`
	AvgMS float64 `json:"avg_ms"`
	P50MS float64 `json:"p50_ms"`
	P95MS float64 `json:"p95_ms"`
	P99MS float64 `json:"p99_ms"`
	MaxMS float64 `json:"max_ms"`
}

func FromDurations(values []time.Duration) Distribution {
	if len(values) == 0 {
		return Distribution{}
	}

	cp := append([]time.Duration(nil), values...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })

	var sum time.Duration
	for _, v := range cp {
		sum += v
	}

	return Distribution{
		Count: len(cp),
		MinMS: toMS(cp[0]),
		AvgMS: toMS(time.Duration(int64(sum) / int64(len(cp)))),
		P50MS: toMS(percentile(cp, 0.50)),
		P95MS: toMS(percentile(cp, 0.95)),
		P99MS: toMS(percentile(cp, 0.99)),
		MaxMS: toMS(cp[len(cp)-1]),
	}
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	rank := p * float64(len(sorted)-1)
	low := int(math.Floor(rank))
	high := int(math.Ceil(rank))
	if low == high {
		return sorted[low]
	}
	weight := rank - float64(low)
	a := float64(sorted[low])
	b := float64(sorted[high])
	return time.Duration(a + (b-a)*weight)
}

func toMS(v time.Duration) float64 {
	return float64(v) / float64(time.Millisecond)
}
