package printer

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/bojand/ghz/runner"
	promtypes "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md

func (rp *ReportPrinter) printPrometheus() error {
	encoder := expfmt.NewEncoder(rp.Out, expfmt.NewFormat(expfmt.TypeTextPlain))

	labels, err := rp.getCommonPrometheusLabels()
	if err != nil {
		return err
	}

	if err := rp.printPrometheusMetricGauge(
		encoder, labels,
		"ghz_run_count",
		&promtypes.Gauge{Value: ptrFloat64(float64(rp.Report.Count))}); err != nil {
		return err
	}

	if err := rp.printPrometheusMetricGauge(
		encoder, labels,
		"ghz_run_total", &promtypes.Gauge{Value: ptrFloat64(float64(rp.Report.Total))}); err != nil {
		return err
	}

	if err := rp.printPrometheusMetricGauge(
		encoder, labels,
		"ghz_run_average", &promtypes.Gauge{Value: ptrFloat64(float64(rp.Report.Average.Nanoseconds()))}); err != nil {
		return err
	}

	if err := rp.printPrometheusMetricGauge(
		encoder, labels,
		"ghz_run_fastest", &promtypes.Gauge{Value: ptrFloat64(float64(rp.Report.Fastest.Nanoseconds()))}); err != nil {
		return err
	}

	if err := rp.printPrometheusMetricGauge(
		encoder, labels,
		"ghz_run_slowest", &promtypes.Gauge{Value: ptrFloat64(float64(rp.Report.Slowest.Nanoseconds()))}); err != nil {
		return err
	}

	if err := rp.printPrometheusMetricGauge(
		encoder, labels,
		"ghz_run_rps", &promtypes.Gauge{Value: ptrFloat64(rp.Report.Rps)}); err != nil {
		return err
	}

	// histogram
	latencyName := "ghz_run_histogram"
	metricType := promtypes.MetricType_HISTOGRAM
	mf := promtypes.MetricFamily{
		Name: &latencyName,
		Type: &metricType,
	}

	metrics := make([]*promtypes.Metric, 0, 1)

	metrics = append(metrics, &promtypes.Metric{
		Label: labels,
		Histogram: &promtypes.Histogram{
			SampleCount: &rp.Report.Count,
			SampleSum:   ptrFloat64(float64(rp.Report.Total.Nanoseconds())),
			Bucket:      make([]*promtypes.Bucket, 0, len(rp.Report.Histogram)),
		},
	})

	mf.Metric = append(mf.Metric, metrics...)

	for _, v := range rp.Report.Histogram {
		metrics[0].Histogram.Bucket = append(metrics[0].Histogram.Bucket,
			&promtypes.Bucket{
				CumulativeCount: ptrUint64(uint64(v.Count)),
				UpperBound:      ptrFloat64(v.Mark),
			})
	}

	if err := encoder.Encode(&mf); err != nil {
		return err
	}

	// latency distribution
	latencyName = "ghz_run_latency"
	metricType = promtypes.MetricType_SUMMARY
	mf = promtypes.MetricFamily{
		Name: &latencyName,
		Type: &metricType,
	}

	metrics = make([]*promtypes.Metric, 0, 1)

	metrics = append(metrics, &promtypes.Metric{
		Label: labels,
		Summary: &promtypes.Summary{
			SampleCount: &rp.Report.Count,
			SampleSum:   ptrFloat64(float64(rp.Report.Total.Nanoseconds())),
			Quantile:    make([]*promtypes.Quantile, 0, len(rp.Report.LatencyDistribution)),
		},
	})

	mf.Metric = append(mf.Metric, metrics...)

	for _, v := range rp.Report.LatencyDistribution {
		metrics[0].Summary.Quantile = append(metrics[0].Summary.Quantile,
			&promtypes.Quantile{
				Quantile: ptrFloat64(float64(v.Percentage) / 100.0),
				Value:    ptrFloat64(float64(v.Latency.Nanoseconds())),
			})
	}

	if err := encoder.Encode(&mf); err != nil {
		return err
	}

	// errors
	errCount := 0
	for _, v := range rp.Report.ErrorDist {
		errCount += v
	}

	if err := rp.printPrometheusMetricGauge(
		encoder, labels,
		"ghz_run_errors", &promtypes.Gauge{Value: ptrFloat64(float64(errCount))}); err != nil {
		return err
	}

	return nil
}

func (rp *ReportPrinter) printPrometheusMetricGauge(
	encoder expfmt.Encoder, labels []*promtypes.LabelPair,
	name string, value *promtypes.Gauge) error {
	metricType := promtypes.MetricType_GAUGE
	mf := promtypes.MetricFamily{
		Name: &name,
		Type: &metricType,
	}

	metrics := make([]*promtypes.Metric, 0, 1)

	metrics = append(metrics, &promtypes.Metric{
		Label: labels,
		Gauge: value,
	})

	mf.Metric = append(mf.Metric, metrics...)

	return encoder.Encode(&mf)
}

func (rp *ReportPrinter) getCommonPrometheusLabels() ([]*promtypes.LabelPair, error) {

	options := map[string]string{
		"name":       rp.Report.Name,
		"end_reason": string(rp.Report.EndReason),
	}

	type Alias runner.Options
	j, err := json.Marshal(&struct {
		ImportPaths      string `json:"import-paths"`
		SkipTLS          string `json:"skipTLS,omitempty"`
		Insecure         string `json:"insecure"`
		Async            string `json:"async,omitempty"`
		Binary           string `json:"binary"`
		CountErrors      string `json:"count-errors,omitempty"`
		RPS              string `json:"rps,omitempty"`
		LoadStart        string `json:"load-start"`
		LoadEnd          string `json:"load-end"`
		LoadStep         string `json:"load-step"`
		Concurrency      string `json:"concurrency,omitempty"`
		CStart           string `json:"concurrency-start"`
		CEnd             string `json:"concurrency-end"`
		CStep            string `json:"concurrency-step"`
		Total            string `json:"total,omitempty"`
		Connections      string `json:"connections,omitempty"`
		CPUs             string `json:"CPUs"`
		SkipFirst        string `json:"skipFirst,omitempty"`
		Data             string `json:"data,omitempty"`
		Metadata         string `json:"metadata,omitempty"`
		LoadStepDuration string `json:"load-step-duration"`
		LoadMaxDuration  string `json:"load-max-duration"`
		CStepDuration    string `json:"concurrency-step-duration"`
		CMaxDuration     string `json:"concurrency-max-duration"`
		Duration         string `json:"duration,omitempty"`
		Timeout          string `json:"timeout,omitempty"`
		DialTimeout      string `json:"dial-timeout,omitempty"`
		KeepaliveTime    string `json:"keepalive,omitempty"`
		*Alias
	}{
		ImportPaths:      strings.Join(rp.Report.Options.ImportPaths, ","),
		SkipTLS:          *ptrBoolToStr(rp.Report.Options.SkipTLS),
		Insecure:         *ptrBoolToStr(rp.Report.Options.Insecure),
		Async:            *ptrBoolToStr(rp.Report.Options.Async),
		Binary:           *ptrBoolToStr(rp.Report.Options.Binary),
		CountErrors:      *ptrBoolToStr(rp.Report.Options.CountErrors),
		RPS:              *ptrString(strconv.Itoa(rp.Report.Options.RPS)),
		LoadStart:        *ptrString(strconv.Itoa(rp.Report.Options.LoadStart)),
		LoadEnd:          *ptrString(strconv.Itoa(rp.Report.Options.LoadEnd)),
		LoadStep:         *ptrString(strconv.Itoa(rp.Report.Options.LoadStep)),
		Concurrency:      *ptrString(strconv.Itoa(rp.Report.Options.Concurrency)),
		CStart:           *ptrString(strconv.Itoa(rp.Report.Options.CStart)),
		CEnd:             *ptrString(strconv.Itoa(rp.Report.Options.CEnd)),
		CStep:            *ptrString(strconv.Itoa(rp.Report.Options.CStep)),
		Total:            *ptrString(strconv.Itoa(rp.Report.Options.Total)),
		Connections:      *ptrString(strconv.Itoa(rp.Report.Options.Connections)),
		CPUs:             *ptrString(strconv.Itoa(rp.Report.Options.CPUs)),
		SkipFirst:        *ptrString(strconv.Itoa(rp.Report.Options.SkipFirst)),
		Data:             "",
		Metadata:         "",
		LoadStepDuration: *ptrString(strconv.Itoa(int(rp.Report.Options.LoadStepDuration.Nanoseconds()))),
		LoadMaxDuration:  *ptrString(strconv.Itoa(int(rp.Report.Options.LoadMaxDuration.Nanoseconds()))),
		CStepDuration:    *ptrString(strconv.Itoa(int(rp.Report.Options.CStepDuration.Nanoseconds()))),
		CMaxDuration:     *ptrString(strconv.Itoa(int(rp.Report.Options.CMaxDuration.Nanoseconds()))),
		Duration:         *ptrString(strconv.Itoa(int(rp.Report.Options.Duration.Nanoseconds()))),
		Timeout:          *ptrString(strconv.Itoa(int(rp.Report.Options.Timeout.Nanoseconds()))),
		DialTimeout:      *ptrString(strconv.Itoa(int(rp.Report.Options.DialTimeout.Nanoseconds()))),
		KeepaliveTime:    *ptrString(strconv.Itoa(int(rp.Report.Options.KeepaliveTime.Nanoseconds()))),
		Alias:            (*Alias)(&rp.Report.Options),
	})
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &options)
	if err != nil {
		return nil, err
	}

	if rp.Report.Options.CSchedule == "const" {
		delete(options, "concurrency-schedule")
		delete(options, "concurrency-start")
		delete(options, "concurrency-end")
		delete(options, "concurrency-step")
		delete(options, "concurrency-step-duration")
		delete(options, "concurrency-max-duration")
	}

	if rp.Report.Options.LoadSchedule == "const" {
		delete(options, "load-schedule")
		delete(options, "load-start")
		delete(options, "load-end")
		delete(options, "load-step")
		delete(options, "load-step-duration")
		delete(options, "load-max-duration")
	}

	labels := make([]*promtypes.LabelPair, 0, len(rp.Report.Tags)+5)

	for k, v := range options {
		k, v := k, v

		k = strings.Replace(k, "-", "_", -1)

		labels = append(labels, &promtypes.LabelPair{
			Name:  &k,
			Value: &v,
		})
	}

	return labels, nil
}

func ptrUint64(v uint64) *uint64 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrString(v string) *string {
	return &v
}

func ptrBoolToStr(v bool) *string {
	switch v {
	case true:
		return ptrString("true")
	default:
		return ptrString("false")
	}
}
