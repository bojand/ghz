import _ from 'lodash'
import Chart from 'chart.js'

import { colors } from './colors'

export function createHistogramChart (report) {
  const categories = _.map(report.histogram, h => {
    return Number.parseFloat(h.mark * 1000).toFixed(2)
  })
  const series = _.map(report.histogram, 'count')
  const totalCount = report.count
  const color = Chart.helpers.color
  const barChartData = {
    labels: categories,
    datasets: [
      {
        label: 'Count',
        backgroundColor: color(colors.skyBlue)
          .alpha(0.5)
          .rgbString(),
        borderColor: colors.skyBlue,
        borderWidth: 1,
        data: series
      }
    ]
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
    tooltips: {
      callbacks: {
        title: function (tooltipItem, data) {
          const value = Number.parseInt(tooltipItem[0].xLabel)
          const percent = value / totalCount * 100
          return value + ' ' + '(' + Number.parseFloat(percent).toFixed(1) + ' %)'
        }
      }
    },
    scales: {
      xAxes: [
        {
          display: true,
          scaleLabel: {
            display: true,
            labelString: 'Count'
          }
        }
      ],
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
    type: 'horizontalBar',
    data: barChartData,
    options: barOptions
  }

  return barConfig
}
