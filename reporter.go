package grpcannon

import (
	"sort"
	"time"
)

// We report for max 1M results.
// const maxResult = 1000000

// Reporter gethers all the results
type Reporter struct {
	results chan *callResult
	done    chan bool

	avgTotal       float64
	lats           []float64
	errorDist      map[string]int
	statusCodeDist map[string]int
	totalCount     uint64
}

// Report holds the data for the full test
type Report struct {
	Count   uint64
	Total   time.Duration
	Average time.Duration
	Fastest time.Duration
	Slowest time.Duration
	Rps     float64

	ErrorDist      map[string]int
	StatusCodeDist map[string]int

	LatencyDistribution []LatencyDistribution
	Histogram           []Bucket
}

func newReporter(results chan *callResult, n int) *Reporter {
	cap := min(n, maxResult)
	return &Reporter{
		results:        results,
		done:           make(chan bool, 1),
		statusCodeDist: make(map[string]int),
		errorDist:      make(map[string]int),
		lats:           make([]float64, 0, cap),
	}
}

// Run runs the reporter
func (r *Reporter) Run() {
	for res := range r.results {
		r.totalCount++
		if res.err != nil {
			r.errorDist[res.err.Error()]++
		} else {
			r.avgTotal += res.duration.Seconds()
			r.statusCodeDist[res.status]++
			if len(r.lats) < maxResult {
				r.lats = append(r.lats, res.duration.Seconds())
			}
		}
	}
	r.done <- true
}

// Finalize all the gathered data into a final report
func (r *Reporter) Finalize(total time.Duration) *Report {
	average := r.avgTotal / float64(r.totalCount)
	avgDuration := time.Duration(average * float64(time.Second))
	rps := float64(r.totalCount) / total.Seconds()

	lats := make([]float64, len(r.lats))
	copy(lats, r.lats)

	sort.Float64s(lats)

	fastestNum := lats[0]
	slowestNum := lats[len(lats)-1]
	fastest := time.Duration(fastestNum * float64(time.Second))
	slowest := time.Duration(slowestNum * float64(time.Second))

	rep := &Report{
		Count:          r.totalCount,
		Total:          total,
		Average:        avgDuration,
		Fastest:        fastest,
		Slowest:        slowest,
		Rps:            rps,
		ErrorDist:      r.errorDist,
		StatusCodeDist: r.statusCodeDist}

	rep.Histogram = histogram(&lats, slowestNum, fastestNum)
	rep.LatencyDistribution = latencies(&lats)

	return rep
}

func latencies(latencies *[]float64) []LatencyDistribution {
	lats := *latencies
	pctls := []int{10, 25, 50, 75, 90, 95, 99}
	data := make([]float64, len(pctls))
	j := 0
	for i := 0; i < len(lats) && j < len(pctls); i++ {
		current := i * 100 / len(lats)
		if current >= pctls[j] {
			data[j] = lats[i]
			j++
		}
	}
	res := make([]LatencyDistribution, len(pctls))
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			lat := time.Duration(data[i] * float64(time.Second))
			res[i] = LatencyDistribution{Percentage: pctls[i], Latency: lat}
		}
	}
	return res
}

func histogram(latencies *[]float64, slowest, fastest float64) []Bucket {
	lats := *latencies
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	bs := (slowest - fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = fastest + bs*float64(i)
	}
	buckets[bc] = slowest
	var bi int
	var max int
	for i := 0; i < len(lats); {
		if lats[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
			bi++
		}
	}
	res := make([]Bucket, len(buckets))
	for i := 0; i < len(buckets); i++ {
		res[i] = Bucket{
			Mark:      buckets[i],
			Count:     counts[i],
			Frequency: float64(counts[i]) / float64(len(lats)),
		}
	}
	return res
}

// LatencyDistribution holds latency distribution data
type LatencyDistribution struct {
	Percentage int
	Latency    time.Duration
}

// Bucket holds histogram data
type Bucket struct {
	Mark      float64
	Count     int
	Frequency float64
}
