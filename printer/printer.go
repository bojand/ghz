package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/template"
	"github.com/bojand/ghz/runner"
)

const (
	barChar = "∎"
)

// ReportPrinter is used for printing the report
type ReportPrinter struct {
	Out    io.Writer
	Report *runner.Report
}

// Print the report using the given format
// If format is "csv" detailed listing is printer in csv format.
// Otherwise the summary of results is printed.
//
// Supported Format:
//
// 		summary
// 		csv
// 		json
// 		pretty
// 		html
// 		influx-summary
// 		influx-details
func (rp *ReportPrinter) Print(format string) error {
	if format == "" {
		format = "summary"
	}

	switch format {
	case "summary", "csv":
		outputTmpl := defaultTmpl
		if format == "csv" {
			outputTmpl = csvTmpl
		}
		buf := &bytes.Buffer{}
		templ := template.Must(template.New("tmpl").Funcs(tmplFuncMap).Parse(outputTmpl))
		if err := templ.Execute(buf, *rp.Report); err != nil {
			return err
		}

		return rp.print(buf.String())
	case "json", "pretty":
		rep, err := json.Marshal(*rp.Report)
		if err != nil {
			return err
		}

		if format == "pretty" {
			var out bytes.Buffer
			err = json.Indent(&out, rep, "", "  ")
			if err != nil {
				return err
			}
			rep = out.Bytes()
		}
		return rp.print(string(rep))
	case "html":
		buf := &bytes.Buffer{}
		templ := template.Must(template.New("tmpl").Funcs(tmplFuncMap).Parse(htmlTmpl))
		if err := templ.Execute(buf, *rp.Report); err != nil {
			return err
		}
		return rp.print(buf.String())
	case "influx-summary":
		return rp.print(rp.getInfluxLine())
	case "influx-details":
		return rp.printInfluxDetails()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (rp *ReportPrinter) getInfluxLine() string {
	measurement := "ghz_run"
	tags := rp.getInfluxTags(true)
	fields := rp.getInfluxFields()
	timestamp := rp.Report.Date.UnixNano()
	if timestamp < 0 {
		timestamp = 0
	}

	return fmt.Sprintf("%v,%v %v %v", measurement, tags, fields, timestamp)
}

func (rp *ReportPrinter) printInfluxDetails() error {
	measurement := "ghz_detail"
	commonTags := rp.getInfluxTags(false)

	for _, v := range rp.Report.Details {
		values := make([]string, 3)
		values[0] = fmt.Sprintf("latency=%v", v.Latency.Nanoseconds())
		values[1] = fmt.Sprintf(`error="%v"`, cleanInfluxString(v.Error))
		values[2] = fmt.Sprintf(`status="%v"`, v.Status)

		tags := commonTags

		if v.Error != "" {
			tags = tags + ",hasError=true"
		} else {
			tags = tags + ",hasError=false"
		}

		timestamp := v.Timestamp.UnixNano()

		fields := strings.Join(values, ",")

		if _, err := fmt.Fprintf(rp.Out, "%v,%v %v %v\n", measurement, tags, fields, timestamp); err != nil {
			return err
		}
	}
	return nil
}

func (rp *ReportPrinter) getInfluxTags(addErrors bool) string {
	s := make([]string, 0, 10)

	if rp.Report.Name != "" {
		s = append(s, fmt.Sprintf(`name="%v"`, cleanInfluxString(strings.TrimSpace(rp.Report.Name))))
	}

	options := rp.Report.Options

	if options.Proto != "" {
		s = append(s, fmt.Sprintf(`proto="%v"`, options.Proto))
	} else if options.Protoset != "" {
		s = append(s, fmt.Sprintf(`Protoset="%v"`, options.Protoset))
	}

	s = append(s, fmt.Sprintf(`call="%v"`, options.Call))
	s = append(s, fmt.Sprintf(`host="%v"`, options.Host))
	s = append(s, fmt.Sprintf("n=%v", options.Total))

	if options.CSchedule == "const" {
		s = append(s, fmt.Sprintf("c=%v", options.Concurrency))
	} else {
		s = append(s, fmt.Sprintf("concurrency-schedule=%v", options.CSchedule))
		s = append(s, fmt.Sprintf("concurrency-start=%v", options.CStart))
		s = append(s, fmt.Sprintf("concurrency-end=%v", options.CEnd))
		s = append(s, fmt.Sprintf("concurrency-step=%v", options.CStep))
		s = append(s, fmt.Sprintf("concurrency-step-duration=%v", options.CStepDuration))
		s = append(s, fmt.Sprintf("concurrency-max-duration=%v", options.CMaxDuration))
	}

	if options.LoadSchedule == "const" {
		s = append(s, fmt.Sprintf("rps=%v", options.RPS))
	} else {
		s = append(s, fmt.Sprintf("load-schedule=%v", options.LoadSchedule))
		s = append(s, fmt.Sprintf("load-start=%v", options.LoadStart))
		s = append(s, fmt.Sprintf("load-end=%v", options.LoadEnd))
		s = append(s, fmt.Sprintf("load-step=%v", options.LoadStep))
		s = append(s, fmt.Sprintf("load-step-duration=%v", options.LoadStepDuration))
		s = append(s, fmt.Sprintf("load-max-duration=%v", options.LoadMaxDuration))
	}

	s = append(s, fmt.Sprintf("z=%v", options.Duration.Nanoseconds()))
	s = append(s, fmt.Sprintf("timeout=%v", options.Timeout.Seconds()))
	s = append(s, fmt.Sprintf("dial_timeout=%v", options.DialTimeout.Seconds()))
	s = append(s, fmt.Sprintf("keepalive=%v", options.KeepaliveTime.Seconds()))

	dataStr := `""`
	dataBytes, err := json.Marshal(options.Data)
	if err == nil && len(dataBytes) > 0 {
		dataBytes, err = json.Marshal(string(dataBytes))
		if err == nil {
			dataStr = string(dataBytes)
		}
	}

	dataStr = cleanInfluxString(dataStr)

	s = append(s, fmt.Sprintf("data=%s", dataStr))

	mdStr := `""`
	if options.Metadata != nil {
		mdBytes, err := json.Marshal(options.Metadata)
		if err == nil {
			mdBytes, err = json.Marshal(string(mdBytes))
			if err == nil {
				mdStr = string(mdBytes)
			}
		}

		mdStr = cleanInfluxString(mdStr)
	}

	s = append(s, fmt.Sprintf("metadata=%s", mdStr))

	callTagsStr := `""`
	if len(rp.Report.Tags) > 0 {
		callTagsBytes, err := json.Marshal(rp.Report.Tags)
		if err == nil {
			callTagsBytes, err = json.Marshal(string(callTagsBytes))
			if err == nil {
				callTagsStr = string(callTagsBytes)
			}
		}

		callTagsStr = cleanInfluxString(callTagsStr)
	}

	s = append(s, fmt.Sprintf("tags=%s", callTagsStr))

	if addErrors {
		errCount := 0
		if len(rp.Report.ErrorDist) > 0 {
			for _, v := range rp.Report.ErrorDist {
				errCount += v
			}
		}

		s = append(s, fmt.Sprintf("errors=%v", errCount))

		hasErrors := false
		if errCount > 0 {
			hasErrors = true
		}

		s = append(s, fmt.Sprintf("has_errors=%v", hasErrors))
	}

	return strings.Join(s, ",")
}

func (rp *ReportPrinter) getInfluxFields() string {
	s := make([]string, 0, 5)

	s = append(s, fmt.Sprintf("count=%v", rp.Report.Count))
	s = append(s, fmt.Sprintf("total=%v", rp.Report.Total.Nanoseconds()))
	s = append(s, fmt.Sprintf("average=%v", rp.Report.Average.Nanoseconds()))
	s = append(s, fmt.Sprintf("fastest=%v", rp.Report.Fastest.Nanoseconds()))
	s = append(s, fmt.Sprintf("slowest=%v", rp.Report.Slowest.Nanoseconds()))
	s = append(s, fmt.Sprintf("rps=%4.2f", rp.Report.Rps))

	if len(rp.Report.LatencyDistribution) > 0 {
		for _, v := range rp.Report.LatencyDistribution {
			if v.Percentage == 50 {
				s = append(s, fmt.Sprintf("median=%v", v.Latency.Nanoseconds()))
			}

			if v.Percentage == 95 {
				s = append(s, fmt.Sprintf("p95=%v", v.Latency.Nanoseconds()))
			}
		}
	}

	errCount := 0
	if len(rp.Report.ErrorDist) > 0 {
		for _, v := range rp.Report.ErrorDist {
			errCount += v
		}
	}

	s = append(s, fmt.Sprintf("errors=%v", errCount))

	return strings.Join(s, ",")
}

func (rp *ReportPrinter) print(s string) error {
	_, err := fmt.Fprint(rp.Out, s)
	return err
}

var tmplFuncMap = template.FuncMap{
	"formatMilli":      formatMilli,
	"formatSeconds":    formatSeconds,
	"histogram":        histogram,
	"jsonify":          jsonify,
	"formatMark":       formatMarkMs,
	"formatPercent":    formatPercent,
	"formatStatusCode": formatStatusCode,
	"formatErrorDist":  formatErrorDist,
	"formatDate":       formatDate,
	"formatNanoUnit":   formatNanoUnit,
}

func jsonify(v interface{}, pretty bool) string {
	d, _ := json.Marshal(v)
	if !pretty {
		return string(d)
	}

	var out bytes.Buffer
	err := json.Indent(&out, d, "", "  ")
	if err != nil {
		return string(d)
	}

	return out.String()
}

func formatNanoUnit(d time.Duration) string {
	v := d.Nanoseconds()
	if v < 10000 {
		return fmt.Sprintf("%+v ns", v)
	}

	valMs := float64(v) / 1000000.0
	if valMs < 1000 {
		return fmt.Sprintf("%4.2f ms", valMs)
	}

	return fmt.Sprintf("%4.2f s", float64(valMs)/1000.0)
}

func formatMilli(duration float64) string {
	return fmt.Sprintf("%4.2f", duration*1000)
}

func formatDate(d time.Time) string {
	return d.Format("Mon Jan _2 2006 @ 15:04:05")
}

func formatSeconds(duration float64) string {
	return fmt.Sprintf("%4.2f", duration)
}

func formatPercent(num int, total uint64) string {
	p := float64(num) / float64(total)
	return fmt.Sprintf("%.2f", p*100)
}

func histogram(buckets []runner.Bucket) string {
	max := 0
	for _, b := range buckets {
		if v := b.Count; v > max {
			max = v
		}
	}
	res := new(bytes.Buffer)
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = (buckets[i].Count*40 + max/2) / max
		}
		res.WriteString(fmt.Sprintf("  %4.3f [%v]\t|%v\n", buckets[i].Mark*1000, buckets[i].Count, strings.Repeat(barChar, barLen)))
	}
	return res.String()
}

func formatMarkMs(m float64) string {
	m = m * 1000.0

	if m < 1 {
		return fmt.Sprintf("'%4.4f ms'", m)
	}

	return fmt.Sprintf("'%4.2f ms'", m)
}

func formatStatusCode(statusCodeDist map[string]int) string {
	padding := 3
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, padding, ' ', 0)
	for status, count := range statusCodeDist {
		// bytes.Buffer can be assumed to not fail on write
		_, _ = fmt.Fprintf(w, "  [%+s]\t%+v responses\t\n", status, count)
	}
	// bytes.Buffer can be assumed to not fail on write
	_ = w.Flush()
	return buf.String()
}

func formatErrorDist(errDist map[string]int) string {
	padding := 3
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, padding, ' ', 0)
	for status, count := range errDist {
		// bytes.Buffer can be assumed to not fail on write
		_, _ = fmt.Fprintf(w, "  [%+v]\t%+s\t\n", count, status)
	}
	// bytes.Buffer can be assumed to not fail on write
	_ = w.Flush()
	return buf.String()
}

func cleanInfluxString(input string) string {
	input = strings.Replace(input, " ", "\\ ", -1)
	input = strings.Replace(input, ",", "\\,", -1)
	input = strings.Replace(input, "=", "\\=", -1)
	return input
}
