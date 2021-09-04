package printer

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (rp *ReportPrinter) printInfluxLine() error {
	measurement := "ghz_run"
	tags := rp.getInfluxTags(true)
	fields := rp.getInfluxFields()
	timestamp := rp.Report.Date.UnixNano()
	if timestamp < 0 {
		timestamp = 0
	}

	if _, err := fmt.Fprintf(rp.Out, "%v,%v %v %v", measurement, tags, fields, timestamp); err != nil {
		return err
	}

	return nil
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

func cleanInfluxString(input string) string {
	input = strings.Replace(input, " ", "\\ ", -1)
	input = strings.Replace(input, ",", "\\,", -1)
	input = strings.Replace(input, "=", "\\=", -1)
	return input
}
