import _ from 'lodash'
import { colors } from './colors.js'

function createChartData (reports) {
  let data = reports
  const avgs = data.map(d => d.average / 1000000)
  const fasts = data.map(d => d.fastest / 1000000)
  const slows = data.map(d => d.slowest / 1000000)
  const rps = data.map(d => d.rps)
  const nine5 = _(data)
    .map(r => {
      const elem = _.find(r.latencyDistribution, ['percentage', 95])
      if (elem) {
        return elem.latency / 1000000
      }
    })
    .compact()
    .valueOf()
  const dates = data.map(d => d.date)
  return {
    averate: avgs,
    fastest: fasts,
    slowest: slows,
    nine5: nine5,
    rps,
    dates
  }
}

function createLineChart (projects) {
  if (!projects) {
    return
  }

  const chartData = this.createChartData()
  const dates = chartData.dates
  const avgData = []
  const fastData = []
  const slowData = []
  const n5Data = []
  const rpsData = []

  dates.forEach((v, i) => {
    const d = new Date(v)
    avgData[i] = {
      x: d,
      y: this.formatFloat(chartData.averate[i])
    }
    fastData[i] = {
      x: d,
      y: this.formatFloat(chartData.fastest[i])
    }
    slowData[i] = {
      x: d,
      y: this.formatFloat(chartData.slowest[i])
    }
    n5Data[i] = {
      x: d,
      y: this.formatFloat(chartData.nine5[i])
    }
    rpsData[i] = {
      x: d,
      y: this.formatFloat(chartData.rps[i])
    }
  })

  const datasets = [
    {
      label: 'Average',
      backgroundColor: colors.blue,
      borderColor: colors.blue,
      fill: false,
      data: avgData,
      yAxisID: 'y-axis-lat'
    },
    {
      label: 'Fastest',
      backgroundColor: colors.green,
      borderColor: colors.green,
      fill: false,
      data: fastData,
      yAxisID: 'y-axis-lat'
    },
    {
      label: 'Slowest',
      backgroundColor: colors.red,
      borderColor: colors.red,
      fill: false,
      data: slowData,
      yAxisID: 'y-axis-lat'
    },
    {
      label: '95th',
      backgroundColor: colors.orange,
      borderColor: colors.orange,
      fill: false,
      data: n5Data,
      yAxisID: 'y-axis-lat'
    },
    {
      label: 'RPS',
      backgroundColor: colors.grey,
      borderColor: colors.grey,
      fill: false,
      data: rpsData,
      yAxisID: 'y-axis-rps'
    }
  ]

  var config = {
    type: 'line',
    data: {
      labels: dates,
      datasets: datasets
    },
    options: {
      responsive: true,
      title: {
        display: true,
        text: 'Change Over Time'
      },
      tooltips: {
        mode: 'index',
        intersect: false
      },
      hover: {
        mode: 'nearest',
        intersect: true
      },
      scales: {
        xAxes: [
          {
            display: true,
            scaleLabel: {
              display: true,
              labelString: 'Date'
            },
            type: 'time'
          }
        ],
        yAxes: [
          {
            display: true,
            position: 'left',
            id: 'y-axis-lat',
            scaleLabel: {
              display: true,
              labelString: 'Latency (ms)'
            }
          },
          {
            type: 'linear',
            display: true,
            scaleLabel: {
              display: true,
              labelString: 'RPS'
            },
            position: 'right',
            id: 'y-axis-rps',
            // grid line settings
            gridLines: {
              drawOnChartArea: false // only want the grid lines for one axis to show up
            }
          }
        ]
      }
    }
  }

  return config
}

module.exports = {
  createChartData,
  createLineChart
}
