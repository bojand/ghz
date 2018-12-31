import _ from 'lodash'
import Chart from 'chart.js'

import { formatNano } from '../lib/common'

export function createComparisonChart (report1, report2, color1, color2) {
  const color = Chart.helpers.color

  const report1Name = report1.name
    ? `${report1.name} [ID:${report1.id}]`
    : `Report: ${report1.id}`

  const report2Name = report2.name
    ? `${report2.name} [ID:${report2.id}]`
    : `Report: ${report2.id}`

  const report1Latencies = _.keyBy(report1.latencyDistribution, 'percentage')
  const report2Latencies = _.keyBy(report2.latencyDistribution, 'percentage')

  const chartData = {
    labels: ['Fastest', 'Average', 'Slowest', '10 %', '25 %', '50 %', '75 %', '95 %', '99 %'],
    datasets: [{
      label: report1Name,
      backgroundColor: color(color1)
        .alpha(0.5)
        .rgbString(),
      borderColor: color1,
      borderWidth: 1,
      data: [
        formatNano(report1.fastest),
        formatNano(report1.average),
        formatNano(report1.slowest),
        formatNano(report1Latencies['10'].latency),
        formatNano(report1Latencies['25'].latency),
        formatNano(report1Latencies['50'].latency),
        formatNano(report1Latencies['75'].latency),
        formatNano(report1Latencies['95'].latency),
        formatNano(report1Latencies['99'].latency)
      ]
    }, {
      label: report2Name,
      backgroundColor: color(color2)
        .alpha(0.5)
        .rgbString(),
      borderColor: color2,
      borderWidth: 1,
      data: [
        formatNano(report2.fastest),
        formatNano(report2.average),
        formatNano(report2.slowest),
        formatNano(report2Latencies['10'].latency),
        formatNano(report2Latencies['25'].latency),
        formatNano(report2Latencies['50'].latency),
        formatNano(report2Latencies['75'].latency),
        formatNano(report2Latencies['95'].latency),
        formatNano(report2Latencies['99'].latency)
      ]
    }]
  }

  const barOptions = {
    elements: {
      rectangle: {
        borderWidth: 2
      }
    },
    responsive: true,
    legend: {
      display: false
    },
    scales: {
      yAxes: [
        {
          display: true,
          scaleLabel: {
            display: true,
            labelString: 'Latency (ms)'
          }
        }
      ]
    }
  }

  const barConfig = {
    data: chartData,
    options: barOptions
  }

  return barConfig
}
