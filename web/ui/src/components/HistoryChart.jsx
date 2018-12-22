import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import Chart from 'chart.js'

import {
  createLineChart
} from './ProjectChartData'

export default class HistoryChart extends Component {
  componentDidMount () {
    console.log('HistoryChart: componentDidMount')
    const node = this.node
    const config = createLineChart(this.props.reports)
    const myChart = new Chart(node, config)
  }

  componentDidUpdate () {
    console.log('HistoryChart: componentDidUpdate')
    const node = this.node
    const config = createLineChart(this.props.reports)
    const myChart = new Chart(node, config)
  }

  render () {
    return (
      <Pane>
        <canvas
          style={{ width: 800, height: 300 }}
          ref={node => (this.node = node)}
        />
      </Pane>
    )
  }
}
