package printer

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
		  <li>
            <a href="#options">
              <span class="icon is-small">
                <i class="fas fa-cog" aria-hidden="true"></i>
              </span>
              <span>Options</span>
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

			<br />

			<div class="container">
				<div class="content">
					<a name="options">
						<h3>Options</h3>
					</a>
					<article class="message">
						<div class="message-body">
							<pre style="background-color: transparent;">{{ jsonify .Options true }}</pre>
						</div>
					</article>
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
