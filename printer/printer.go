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
	s = append(s, fmt.Sprintf("c=%v", options.Concurrency))
	s = append(s, fmt.Sprintf("qps=%v", options.QPS))
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

var (
	defaultTmpl = `
Summary:
{{ if .Name }}  Name:		{{ .Name }}
{{ end }}  Count:	{{ .Count }}
  Total:	{{ formatNanoUnit .Total }}
  Slowest:	{{ formatNanoUnit .Slowest }}
  Fastest:	{{ formatNanoUnit .Fastest }}
  Average:	{{ formatNanoUnit .Average }}
  Requests/sec:	{{ formatSeconds .Rps }}

Response time histogram:
{{ histogram .Histogram }}
Latency distribution:{{ range .LatencyDistribution }}
  {{ .Percentage }} % in {{ formatNanoUnit .Latency }} {{ end }}

{{ if gt (len .StatusCodeDist) 0 }}Status code distribution:
{{ formatStatusCode .StatusCodeDist }}{{ end }}
{{ if gt (len .ErrorDist) 0 }}Error distribution:
{{ formatErrorDist .ErrorDist }}{{ end }}
`

	csvTmpl = `
duration (ms),status,error{{ range $i, $v := .Details }}
{{ formatMilli .Latency.Seconds }},{{ .Status }},{{ .Error }}{{ end }}
`

	htmlTmpl = `
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>ghz{{ if .Name }} - {{ .Name }}{{end}}</title>
    <script src="https://d3js.org/d3.v5.min.js"></script>
		<script src="https://cdn.jsdelivr.net/npm/papaparse@4.5.0/papaparse.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/britecharts@2/dist/bundled/britecharts.min.js"></script>

    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/britecharts/dist/css/britecharts.min.css" type="text/css" /></head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.7.1/css/bulma.min.css" />

  </head>

	<body>

		<section class="section">

		<div class="container">
			{{ if .Name }}
			<h1 class="title">{{ .Name }}</h1>
			{{ end }}
			{{ if .Date }}
				<h2 class="subtitle">{{ formatDate .Date }}</h2>
			{{ end }}
		</div>
		<br />

		<div class="container">
      <nav class="breadcrumb has-bullet-separator" aria-label="breadcrumbs">
        <ul>
          <li>
            <a href="#summary">
              <span class="icon is-small">
                <i class="fas fa-clipboard-list" aria-hidden="true"></i>
              </span>
              <span>Summary</span>
            </a>
          </li>
          <li>
            <a href="#histogram">
              <span class="icon is-small">
                <i class="fas fa-chart-bar" aria-hidden="true"></i>
              </span>
              <span>Histogram</span>
            </a>
          </li>
          <li>
            <a href="#latency">
              <span class="icon is-small">
                <i class="far fa-clock" aria-hidden="true"></i>
              </span>
              <span>Latency Distribution</span>
            </a>
          </li>
          <li>
            <a href="#status">
              <span class="icon is-small">
                <i class="far fa-check-square" aria-hidden="true"></i>
              </span>
              <span>Status Distribution</span>
            </a>
					</li>
					{{ if gt (len .ErrorDist) 0 }}
          <li>
            <a href="#errors">
              <span class="icon is-small">
                <i class="fas fa-exclamation-circle" aria-hidden="true"></i>
              </span>
              <span>Errors</span>
            </a>
					</li>
					{{ end }}
          <li>
            <a href="#data">
              <span class="icon is-small">
                <i class="far fa-file-alt" aria-hidden="true"></i>
              </span>
              <span>Data</span>
            </a>
          </li>
        </ul>
      </nav>
      <hr />
		</div>

		{{ if gt (len .Tags) 0 }}

			<div class="container">
				<div class="field is-grouped">

				{{ range $tag, $val := .Tags }}

					<div class="control">
						<div class="tags has-addons">
							<span class="tag is-dark">{{ $tag }}</span>
							<span class="tag is-primary">{{ $val }}</span>
						</div>
					</div>

				{{ end }}

				</div>
			</div>
			<br />
		{{ end }}

	  <div class="container">
			<div class="columns">
				<div class="column is-narrow">
					<div class="content">
						<a name="summary">
							<h3>Summary</h3>
						</a>
						<table class="table">
							<tbody>
								<tr>
									<th>Count</th>
									<td>{{ .Count }}</td>
								</tr>
								<tr>
									<th>Total</th>
									<td>{{ formatNanoUnit .Total }}</td>
								</tr>
								<tr>
									<th>Slowest</th>
								<td>{{ formatNanoUnit .Slowest }}</td>
								</tr>
								<tr>
									<th>Fastest</th>
									<td>{{ formatNanoUnit .Fastest }}</td>
								</tr>
								<tr>
									<th>Average</th>
									<td>{{ formatNanoUnit .Average }}</td>
								</tr>
								<tr>
									<th>Requests / sec</th>
									<td>{{ formatSeconds .Rps }}</td>
								</tr>
							</tbody>
						</table>
					</div>
				</div>
				<div class="column">
					<div class="content">
						<span class="title is-5">
							<strong>Options</strong>
						</span>
						<article class="message">
  						<div class="message-body">
								<pre style="background-color: transparent;">{{ jsonify .Options true }}</pre>
							</div>
						</article>
					</div>
				</div>
			</div>
	  </div>

	  <br />
		<div class="container">
			<div class="content">
				<a name="histogram">
					<h3>Histogram</h3>
				</a>
				<p>
					<div class="js-bar-container"></div>
				</p>
			</div>
	  </div>

	  <br />
		<div class="container">
			<div class="content">
				<a name="latency">
					<h3>Latency distribution</h3>
				</a>
				<table class="table is-fullwidth">
					<thead>
						<tr>
							{{ range .LatencyDistribution }}
								<th>{{ .Percentage }} %</th>
							{{ end }}
						</tr>
					</thead>
					<tbody>
						<tr>
							{{ range .LatencyDistribution }}
								<td>{{ formatNanoUnit .Latency }}</td>
							{{ end }}
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		<br />
		<div class="container">
			<div class="columns">
				<div class="column is-narrow">
					<div class="content">
						<a name="status">
							<h3>Status distribution</h3>
						</a>
						<table class="table is-hoverable">
							<thead>
								<tr>
									<th>Status</th>
									<th>Count</th>
									<th>% of Total</th>
								</tr>
							</thead>
							<tbody>
							  {{ range $code, $num := .StatusCodeDist }}
									<tr>
									  <td>{{ $code }}</td>
										<td>{{ $num }}</td>
										<td>{{ formatPercent $num $.Count }} %</td>
									</tr>
									{{ end }}
								</tbody>
							</table>
						</div>
					</div>
				</div>
			</div>

			{{ if gt (len .ErrorDist) 0 }}

				<br />
				<div class="container">
					<div class="columns">
						<div class="column is-narrow">
							<div class="content">
								<a name="errors">
									<h3>Errors</h3>
								</a>
								<table class="table is-hoverable">
									<thead>
										<tr>
											<th>Error</th>
											<th>Count</th>
											<th>% of Total</th>
										</tr>
									</thead>
									<tbody>
										{{ range $err, $num := .ErrorDist }}
											<tr>
												<td>{{ $err }}</td>
												<td>{{ $num }}</td>
												<td>{{ formatPercent $num $.Count }} %</td>
											</tr>
											{{ end }}
										</tbody>
									</table>
								</div>
							</div>
						</div>
					</div>

			{{ end }}

			<br />
      <div class="container">
        <div class="columns">
          <div class="column is-narrow">
            <div class="content">
              <a name="data">
                <h3>Data</h3>
              </a>

              <a class="button" id="dlJSON">JSON</a>
              <a class="button" id="dlCSV">CSV</a>
            </div>
          </div>
        </div>
			</div>

			<div class="container">
        <hr />
        <div class="content has-text-centered">
          <p>
            Generated by <strong>ghz</strong>
          </p>
          <a href="https://github.com/bojand/ghz"><i class="icon is-medium fab fa-github"></i></a>
        </div>
      </div>

		</section>

  </body>

  <script>

	const count = {{ .Count }};

	const rawData = {{ jsonify .Details false }};

	const data = [
		{{ range .Histogram }}
			{ name: {{ formatMark .Mark }}, value: {{ .Count }} },
		{{ end }}
	];

	function createHorizontalBarChart() {
		let barChart = britecharts.bar(),
			tooltip = britecharts.miniTooltip(),
			barContainer = d3.select('.js-bar-container'),
			containerWidth = barContainer.node() ? barContainer.node().getBoundingClientRect().width : false,
			tooltipContainer,
			dataset;

		tooltip.numberFormat('')
		tooltip.valueFormatter(function(v) {
			var percent = v / count * 100;
			return v + ' ' + '(' + Number.parseFloat(percent).toFixed(1) + ' %)';
		})

		if (containerWidth) {
			dataset = data;
			barChart
				.isHorizontal(true)
				.isAnimated(true)
				.margin({
					left: 100,
					right: 20,
					top: 20,
					bottom: 20
				})
				.colorSchema(britecharts.colors.colorSchemas.teal)
				.width(containerWidth)
				.yAxisPaddingBetweenChart(20)
				.height(400)
				// .hasPercentage(true)
				.enableLabels(true)
				.labelsNumberFormat('')
				.percentageAxisToMaxRatio(1.3)
				.on('customMouseOver', tooltip.show)
				.on('customMouseMove', tooltip.update)
				.on('customMouseOut', tooltip.hide);

			barChart.orderingFunction(function(a, b) {
				var nA = a.name.replace(/ms/gi, '');
				var nB = b.name.replace(/ms/gi, '');

				var vA = Number.parseFloat(nA);
				var vB = Number.parseFloat(nB);

				return vB - vA;
			})

			barContainer.datum(dataset).call(barChart);

			tooltipContainer = d3.select('.js-bar-container .bar-chart .metadata-group');
			tooltipContainer.datum([]).call(tooltip);
		}
	}

	function setJSONDownloadLink () {
		var filename = "data.json";
		var btn = document.getElementById('dlJSON');
		var jsonData = JSON.stringify(rawData)
		var blob = new Blob([jsonData], { type: 'text/json;charset=utf-8;' });
		var url = URL.createObjectURL(blob);
		btn.setAttribute("href", url);
		btn.setAttribute("download", filename);
	}

	function setCSVDownloadLink () {
		var filename = "data.csv";
		var btn = document.getElementById('dlCSV');
		var csv = Papa.unparse(rawData)
		var blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
		var url = URL.createObjectURL(blob);
		btn.setAttribute("href", url);
		btn.setAttribute("download", filename);
	}

	createHorizontalBarChart();

	setJSONDownloadLink();

	setCSVDownloadLink();

	</script>

	<script defer src="https://use.fontawesome.com/releases/v5.1.0/js/all.js"></script>

</html>
`
)
