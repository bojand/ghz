import _ from 'lodash'
// import Chart from 'chart.js'

import { colors, randomColor } from './colors'

const staticColors = [
  colors.red,
  colors.blue,
  colors.orage,
  colors.yellow,
  colors.purple
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

  // const data = {
  //   datasets: [{
  //     data: [10, 20, 30],
  //     backgroundColor: [
  //       colors.red,
  //       colors.orange,
  //       randomColor()
  //     ],
  //     label: 'Status Codes'
  //   }],
  //   labels: [
  //     'Red',
  //     'Yellow',
  //     'Blue'
  //   ]
  // }

  const options = {
    legend: {
      position: 'bottom'
    }
  }

  const config = {
    data,
    options
  }

  console.log(data.datasets[0])

  return config
}
