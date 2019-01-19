import _ from 'lodash'
// import Chart from 'chart.js'

import { colors, randomColor } from './colors'

const staticColors = [
  colors.red,
  colors.teal,
  colors.orange,
  colors.blue,
  colors.purple,
  colors.yellow
]

export function createDoughnutChart (label, dataMap) {
  const dataArray = []
  const backgroundColor = []
  const labels = []

  let counter = 0

  _.forEach(dataMap, (v, k) => {
    dataArray.push(v)

    if (k.indexOf('code =') > 0) {
      const start = k.indexOf('code =') + 'code ='.length
      let l = k.substring(start)
      if (l.indexOf('desc') > 0) {
        l = l.substring(0, l.indexOf('desc'))
      }
      labels.push(l)
    } else {
      labels.push(k)
    }

    if (k.toString().toLowerCase() === 'ok') {
      backgroundColor.push(colors.green)
    } else if (counter < staticColors.length) {
      backgroundColor.push(staticColors[counter])
    } else {
      backgroundColor.push(randomColor())
    }

    counter++
  })

  const data = {
    datasets: [{
      data: dataArray,
      backgroundColor,
      label
    }],
    labels
  }

  const options = {
    legend: {
      position: 'bottom'
    }
  }

  const config = {
    data,
    options
  }

  return config
}
