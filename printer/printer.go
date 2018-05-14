package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/alecthomas/template"
	"github.com/bojand/grpcannon"
)

const (
	barChar = "∎"
)

// ReportPrinter is used for printing the report
type ReportPrinter struct {
	Out    io.Writer
	Report *grpcannon.Report
}

// Print the report using the given format
// If format is "csv" detailed listing is printer in csv format.
// Otherwise the summary of results is printed.
func (rp *ReportPrinter) Print(format string) {
	switch format {
	case "", "csv":
		outputTmpl := defaultTmpl
		if format == "csv" {
			outputTmpl = csvTmpl
		}
		buf := &bytes.Buffer{}
		templ := template.Must(template.New("tmpl").Funcs(tmplFuncMap).Parse(outputTmpl))
		if err := templ.Execute(buf, *rp.Report); err != nil {
			log.Println("error:", err.Error())
			return
		}

		rp.printf(buf.String())

		rp.printf("\n")
	case "json", "pretty":
		rep, err := json.Marshal(*rp.Report)
		if err != nil {
			log.Println("error:", err.Error())
			return
		}

		if format == "pretty" {
			var out bytes.Buffer
			err = json.Indent(&out, rep, "", "  ")
			if err != nil {
				log.Println("error:", err.Error())
				return
			}
			rep = out.Bytes()
		}

		rp.printf(string(rep))
	}
}

func (rp *ReportPrinter) printf(s string, v ...interface{}) {
	fmt.Fprintf(rp.Out, s, v...)
}

var tmplFuncMap = template.FuncMap{
	"formatMilli":   formatMilli,
	"formatSeconds": formatSeconds,
	"histogram":     histogram,
	"jsonify":       jsonify,
}

func jsonify(v interface{}) string {
	d, _ := json.Marshal(v)
	return string(d)
}

func formatMilli(duration float64) string {
	return fmt.Sprintf("%4.2f", duration*1000)
}

func formatSeconds(duration float64) string {
	return fmt.Sprintf("%4.2f", duration)
}

func histogram(buckets []grpcannon.Bucket) string {
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

var (
	defaultTmpl = `
Summary:
  Count:	{{ .Count }}
  Total:	{{ formatMilli .Total.Seconds }} ms
  Slowest:	{{ formatMilli .Slowest.Seconds }} ms
  Fastest:	{{ formatMilli .Fastest.Seconds }} ms
  Average:	{{ formatMilli .Average.Seconds }} ms
  Requests/sec:	{{ formatSeconds .Rps }}

Response time histogram:
{{ histogram .Histogram }}
Latency distribution:{{ range .LatencyDistribution }}
  {{ .Percentage }}%% in {{ formatMilli .Latency.Seconds }} ms{{ end }}
Status code distribution:{{ range $code, $num := .StatusCodeDist }}
  [{{ $code }}]	{{ $num }} responses{{ end }}
{{ if gt (len .ErrorDist) 0 }}Error distribution:{{ range $err, $num := .ErrorDist }}
  [{{ $num }}]	{{ $err }}{{ end }}{{ end }}
`

	csvTmpl = `
duration (ms),status,error{{ range $i, $v := .Details }}
{{ formatMilli .Latency.Seconds }},{{ .Status }},{{ .Error }}{{ end }}
`
)
