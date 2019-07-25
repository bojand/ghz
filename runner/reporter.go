package runner

import (
	"encoding/json"
	"sort"
	"time"
)

// Reporter gathers all the results
type Reporter struct {
	config *RunConfig

	results chan *callResult
	done    chan bool

	totalLatenciesSec float64

	details []ResultDetail

	errorDist      map[string]int
	statusCodeDist map[string]int
	totalCount     uint64
}

// Options represents the request options
type Options struct {
	Host          string        `json:"host,omitempty"`
	Proto         string        `json:"proto,omitempty"`
	Protoset      string        `json:"protoset,omitempty"`
	ImportPaths   []string      `json:"import-paths,omitempty"`
	Call          string        `json:"call,omitempty"`
	CACert        string        `json:"cacert,omitempty"`
	Cert          string        `json:"cert,omitempty"`
	Key           string        `json:"key,omitempty"`
	SkipTLS       bool          `json:"skipTLS,omitempty"`
	CName         string        `json:"cname,omitempty"`
	Authority     string        `json:"authority,omitempty"`
	Insecure      bool          `json:"insecure"`
	Total         uint          `json:"total,omitempty"`
	Concurrency   uint          `json:"concurrency,omitempty"`
	QPS           uint          `json:"qps,omitempty"`
	Connections   uint          `json:"connections,omitempty"`
	Duration      time.Duration `json:"duration,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty"`
	DialTimeout   time.Duration `json:"dial-timeout,omitempty"`
	KeepaliveTime time.Duration `json:"keepalive,omitempty"`

	Data     interface{}        `json:"data,omitempty"`
	Binary   bool               `json:"binary"`
	Metadata *map[string]string `json:"metadata,omitempty"`

	CPUs int    `json:"CPUs"`
	Name string `json:"name,omitempty"`
}

// Report holds the data for the full test
type Report struct {
	Name      string     `json:"name,omitempty"`
	EndReason StopReason `json:"endReason,omitempty"`

	Options Options   `json:"options,omitempty"`
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

	Tags map[string]string `json:"tags,omitempty"`
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

func newReporter(results chan *callResult, c *RunConfig) *Reporter {

	cap := min(c.n, maxResult)

	return &Reporter{
		config:  c,
		results: results,
		done:    make(chan bool, 1),
		details: make([]ResultDetail, 0, cap),

		statusCodeDist: make(map[string]int),
		errorDist:      make(map[string]int),
	}
}

// Run runs the reporter
func (r *Reporter) Run() {
	for res := range r.results {
		errStr := ""

		r.totalCount++
		r.totalLatenciesSec += res.duration.Seconds()
		r.statusCodeDist[res.status]++

		if res.err != nil {
			errStr = res.err.Error()
			r.errorDist[errStr]++
		}

		if len(r.details) < maxResult {
			r.details = append(r.details, ResultDetail{
				Latency:   res.duration,
				Timestamp: res.timestamp,
				Status:    res.status,
				Error:     errStr,
			})
		}
	}
	r.done <- true
}

// Finalize all the gathered data into a final report
func (r *Reporter) Finalize(stopReason StopReason, total time.Duration) *Report {
	rep := &Report{
		Name:           r.config.name,
		EndReason:      stopReason,
		Date:           time.Now(),
		Count:          r.totalCount,
		Total:          total,
		ErrorDist:      r.errorDist,
		StatusCodeDist: r.statusCodeDist}

	rep.Options = Options{
		Host:          r.config.host,
		Proto:         r.config.proto,
		Protoset:      r.config.protoset,
		ImportPaths:   r.config.importPaths,
		Call:          r.config.call,
		CACert:        r.config.cacert,
		Cert:          r.config.cert,
		Key:           r.config.key,
		CName:         r.config.cname,
		SkipTLS:       r.config.skipVerify,
		Insecure:      r.config.insecure,
		Authority:     r.config.authority,
		Total:         uint(r.config.n),
		Concurrency:   uint(r.config.c),
		QPS:           uint(r.config.qps),
		Connections:   uint(r.config.nConns),
		Duration:      r.config.z,
		Timeout:       r.config.timeout,
		DialTimeout:   r.config.dialTimeout,
		KeepaliveTime: r.config.keepaliveTime,
		Binary:        r.config.binary,
		CPUs:          r.config.cpus,
		Name:          r.config.name,
	}

	_ = json.Unmarshal(r.config.data, &rep.Options.Data)

	_ = json.Unmarshal(r.config.metadata, &rep.Options.Metadata)

	_ = json.Unmarshal(r.config.tags, &rep.Tags)

	if len(r.details) > 0 {
		average := r.totalLatenciesSec / float64(r.totalCount)
		rep.Average = time.Duration(average * float64(time.Second))

		rep.Rps = float64(r.totalCount) / total.Seconds()

		okLats := make([]float64, 0)
		for _, d := range r.details {
			if d.Error == "" {
				okLats = append(okLats, d.Latency.Seconds())
			}
		}
		sort.Float64s(okLats)
		if len(okLats) > 0 {
			var fastestNum, slowestNum float64
			fastestNum = okLats[0]
			slowestNum = okLats[len(okLats)-1]

			rep.Fastest = time.Duration(fastestNum * float64(time.Second))
			rep.Slowest = time.Duration(slowestNum * float64(time.Second))
			rep.Histogram = histogram(okLats, slowestNum, fastestNum)
			rep.LatencyDistribution = latencies(okLats)
		}

		rep.Details = r.details
	}

	return rep
}

func latencies(latencies []float64) []LatencyDistribution {
	pctls := []int{10, 25, 50, 75, 90, 95, 99}
	data := make([]float64, len(pctls))
	j := 0
	for i := 0; i < len(latencies) && j < len(pctls); i++ {
		current := i * 100 / len(latencies)
		if current >= pctls[j] {
			data[j] = latencies[i]
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

func histogram(latencies []float64, slowest, fastest float64) []Bucket {
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
	for i := 0; i < len(latencies); {
		if latencies[i] <= buckets[bi] {
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
			Frequency: float64(counts[i]) / float64(len(latencies)),
		}
	}
	return res
}
