package ghz

import (
	"encoding/json"
	"sort"
	"time"
)

// Reporter gethers all the results
type Reporter struct {
	options *Options
	results chan *callResult
	done    chan bool

	avgTotal float64

	lats       []float64
	errors     []string
	statuses   []string
	timestamps []time.Time

	errorDist      map[string]int
	statusCodeDist map[string]int
	totalCount     uint64
}

// Report holds the data for the full test
type Report struct {
	Name string `json:"name,omitempty"`

	Options *Options  `json:"options,omitempty"`
	Date    time.Time `json:"date"`

	Count   uint64        `json:"count"`
	Total   time.Duration `json:"total"`
	Average time.Duration `json:"average"`
	Fastest time.Duration `json:"fastest"`
	Slowest time.Duration `json:"slowest"`
	Rps     float64       `json:"rps"`

	ErrorDist      map[string]int `json:"errorDistribution"`
	StatusCodeDist map[string]int `json:"statusCodeDistribution"`

	LatencyDistribution []LatencyDistribution `json:"latencyDistribution"`
	Histogram           []Bucket              `json:"histogram"`
	Details             []ResultDetail        `json:"details"`
}

// MarshalJSON is custom marshal for report to properly format the date
func (r Report) MarshalJSON() ([]byte, error) {
	type Alias Report
	return json.Marshal(&struct {
		Date string `json:"date"`
		*Alias
	}{
		Date:  r.Date.Format(time.RFC3339),
		Alias: (*Alias)(&r),
	})
}

// LatencyDistribution holds latency distribution data
type LatencyDistribution struct {
	Percentage int           `json:"percentage"`
	Latency    time.Duration `json:"latency"`
}

// Bucket holds histogram data
type Bucket struct {
	// The Mark for histogram bucket in seconds
	Mark float64 `json:"mark"`

	// The count in the bucket
	Count int `json:"count"`

	// The frequency of results in the bucket as a decimal percentage
	Frequency float64 `json:"frequency"`
}

// ResultDetail data for each result
type ResultDetail struct {
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error"`
	Status    string        `json:"status"`
}

func newReporter(results chan *callResult, options *Options) *Reporter {

	cap := min(options.N, maxResult)

	return &Reporter{
		options:        options,
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
			errStr := res.err.Error()
			r.errorDist[errStr]++

			if len(r.errors) < maxResult {
				r.errors = append(r.errors, errStr)
				r.statuses = append(r.statuses, res.status)
			}
		} else {
			r.avgTotal += res.duration.Seconds()
			r.statusCodeDist[res.status]++

			if len(r.lats) < maxResult {
				r.lats = append(r.lats, res.duration.Seconds())
				r.errors = append(r.errors, "")
				r.statuses = append(r.statuses, res.status)
				r.timestamps = append(r.timestamps, time.Now())
			}
		}
	}
	r.done <- true
}

// Finalize all the gathered data into a final report
func (r *Reporter) Finalize(total time.Duration) *Report {
	average := r.avgTotal / float64(len(r.lats))
	avgDuration := time.Duration(average * float64(time.Second))
	rps := float64(r.totalCount) / total.Seconds()

	rep := &Report{
		Name:           r.options.Name,
		Options:        r.options,
		Date:           time.Now(),
		Count:          r.totalCount,
		Total:          total,
		Average:        avgDuration,
		Rps:            rps,
		ErrorDist:      r.errorDist,
		StatusCodeDist: r.statusCodeDist}

	if len(r.lats) > 0 {
		lats := make([]float64, len(r.lats))
		copy(lats, r.lats)
		sort.Float64s(lats)

		var fastestNum, slowestNum float64
		fastestNum = lats[0]
		slowestNum = lats[len(lats)-1]

		rep.Fastest = time.Duration(fastestNum * float64(time.Second))
		rep.Slowest = time.Duration(slowestNum * float64(time.Second))
		rep.Histogram = histogram(&lats, slowestNum, fastestNum)
		rep.LatencyDistribution = latencies(&lats)

		rep.Details = make([]ResultDetail, len(r.lats))
		for i, num := range r.lats {
			lat := time.Duration(num * float64(time.Second))
			rep.Details[i] = ResultDetail{
				Latency:   lat,
				Error:     r.errors[i],
				Status:    r.statuses[i],
				Timestamp: r.timestamps[i],
			}
		}
	}

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
