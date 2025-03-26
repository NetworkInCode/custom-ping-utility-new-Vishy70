package helpers

import (
	"fmt"
	"math"
)

// Statistics for ping results
type PingStats struct {
	transmitted int     // requests sent
	received    int     // replies received
	errors      int     // errors like Destination Host Unreachable
	min         float64 // min time RTT
	max         float64 // max time RTT
	sum1        float64 // cumulative sum RTT
	sum2        float64 // squared sum RTT
	mean        float64 // mean RTT
	stddev      float64 // std deviation RTT
}

// iterativeStats incrementally calculate the
// min, max, avg, S1, S2 RTT using the following formulas:
//
// S1 = sum(Ti)
//
// S2 = sum(Ti ^ 2)
func (stats *PingStats) iterativeStats(time float64) {
	// min is initialized as the first RTT
	if stats.received == 1 {
		stats.min = time
	}

	if time < stats.min {
		stats.min = time
	}
	if time > stats.max {
		stats.max = time
	}

	stats.sum1 += time
	stats.sum2 += math.Pow(time, 2)
}

// finalStats calculate the mean and stddev using the following formulas:
//
// mean = S1 / received
//
// stddev = sqrt(S2 / n - (S1 / n) ^ 2)
func (stats *PingStats) finalStats() {
	stats.mean = stats.sum1 / float64(stats.received)
	stats.stddev = math.Sqrt(stats.sum2/float64(stats.received) - math.Pow(stats.mean, 2))

}

// PrintStatistics is used to summarize all calculated RTT statistics
func PrintStatistics(stats *PingStats) {
	dropPercentage := 0.0
	if stats.transmitted > 0 {
		dropPercentage = float64(stats.transmitted-stats.received) / float64(stats.transmitted) * 100.0
	}

	fmt.Printf("\n--- %s ping statistics ---\n", "target")
	fmt.Printf("%d packets transmitted, %d received, %d errors, %.1f%% packet loss\n",
		stats.transmitted, stats.received, stats.errors, dropPercentage)

	if stats.received > 0 {
		stats.finalStats()
		fmt.Printf("round-trip min/avg/max/stddev = %.3f/%.3f/%.3f/%.3f ms\n",
			stats.min, stats.mean, stats.max, stats.stddev)
	}
}
