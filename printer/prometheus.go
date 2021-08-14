package printer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bojand/ghz/runner"
	promtypes "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md

func (rp *ReportPrinter) printPrometheusLine() error {
	encoder := expfmt.NewEncoder(rp.Out, expfmt.FmtText)

	name := "ghz_run_count"
	metricType := promtypes.MetricType_COUNTER
	mf := promtypes.MetricFamily{
		Name: &name,
		Type: &metricType,
	}

	metrics := make([]*promtypes.Metric, 0, 1)

	labels, err := rp.getCommonPrometheusLabels()
	if err != nil {
		return err
	}

	metrics = append(metrics, &promtypes.Metric{
		Label:   labels,
		Counter: &promtypes.Counter{Value: ptrFloat64(float64(rp.Report.Count))},
	})

	mf.Metric = append(mf.Metric, metrics...)

	err = encoder.Encode(&mf)
	if err != nil {
		return err
	}

	return nil
}

func (rp *ReportPrinter) getCommonPrometheusLabels() ([]*promtypes.LabelPair, error) {
	labels := make([]*promtypes.LabelPair, 0, len(rp.Report.Tags)+5)

	labels = append(labels,
		&promtypes.LabelPair{
			Name:  ptrString("name"),
			Value: &rp.Report.Name,
		},
		&promtypes.LabelPair{
			Name:  ptrString("end_reason"),
			Value: ptrString(string(rp.Report.EndReason)),
		},
	)

	options := map[string]string{}
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
		LoadStepDuration: *ptrString(strconv.Itoa(int(rp.Report.Options.LoadStepDuration))),
		LoadMaxDuration:  *ptrString(strconv.Itoa(int(rp.Report.Options.LoadMaxDuration))),
		CStepDuration:    *ptrString(strconv.Itoa(int(rp.Report.Options.CStepDuration))),
		CMaxDuration:     *ptrString(strconv.Itoa(int(rp.Report.Options.CMaxDuration))),
		Duration:         *ptrString(strconv.Itoa(int(rp.Report.Options.Duration))),
		Timeout:          *ptrString(strconv.Itoa(int(rp.Report.Options.Timeout))),
		DialTimeout:      *ptrString(strconv.Itoa(int(rp.Report.Options.DialTimeout))),
		KeepaliveTime:    *ptrString(strconv.Itoa(int(rp.Report.Options.KeepaliveTime))),
		Alias:            (*Alias)(&rp.Report.Options),
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(string(j))

	err = json.Unmarshal(j, &options)
	if err != nil {
		return nil, err
	}

	for k, v := range options {
		k, v := k, v

		labels = append(labels, &promtypes.LabelPair{
			Name:  &k,
			Value: &v,
		})
	}

	return labels, nil
}

func (rp *ReportPrinter) printPrometheusDetails() error {
	return nil
}

func ptrInt32(v int32) *int32 {
	return &v
}

func ptrInt64(v int64) *int64 {
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
