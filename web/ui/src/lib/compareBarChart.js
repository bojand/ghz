import _ from 'lodash'
import Chart from 'chart.js'

import { formatDiv } from '../lib/common'

export function createComparisonChart (report1, report2, color1, color2) {
  const color = Chart.helpers.color

  const report1Name = report1.name
    ? `${report1.name} [ID: ${report1.id}]`
    : `Report: ${report1.id}`

  const report2Name = report2.name
    ? `${report2.name} [ID: ${report2.id}]`
    : `Report: ${report2.id}`

  const report1Latencies = _.keyBy(report1.latencyDistribution, 'percentage')
  const report2Latencies = _.keyBy(report2.latencyDistribution, 'percentage')

  let unit = 'ns'
  let testValue = report1.average
  let divr = 1

  if (testValue > 1000000) {
    unit = 'ms'
    divr = 1000000
    testValue = testValue / divr
  }

  if (testValue > 1000000000) {
    unit = 's'
    divr = 1000000000
  }

  const chartData = {
    labels: ['Fastest', 'Average', 'Slowest', '10 %', '25 %', '50 %', '75 %', '90 %', '95 %', '99 %'],
    datasets: [{
      label: report1Name,
      backgroundColor: color(color1)
        .alpha(0.5)
        .rgbString(),
      borderColor: color1,
      borderWidth: 1,
      data: [
        formatDiv(report1.fastest, divr),
        formatDiv(report1.average, divr),
        formatDiv(report1.slowest, divr),
        formatDiv(report1Latencies['10'].latency, divr),
        formatDiv(report1Latencies['25'].latency, divr),
        formatDiv(report1Latencies['50'].latency, divr),
        formatDiv(report1Latencies['75'].latency, divr),
        formatDiv(report1Latencies['90'].latency, divr),
        formatDiv(report1Latencies['95'].latency, divr),
        formatDiv(report1Latencies['99'].latency, divr)
      ]
    }, {
      label: report2Name,
      backgroundColor: color(color2)
        .alpha(0.5)
        .rgbString(),
      borderColor: color2,
      borderWidth: 1,
      data: [
        formatDiv(report2.fastest, divr),
        formatDiv(report2.average, divr),
        formatDiv(report2.slowest, divr),
        formatDiv(report2Latencies['10'].latency, divr),
        formatDiv(report2Latencies['25'].latency, divr),
        formatDiv(report2Latencies['50'].latency, divr),
        formatDiv(report2Latencies['75'].latency, divr),
        formatDiv(report2Latencies['90'].latency, divr),
        formatDiv(report2Latencies['95'].latency, divr),
        formatDiv(report2Latencies['99'].latency, divr)
      ]
    }]
  }

  const labelStr = `Latency (${unit})`

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
            labelString: labelStr
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
