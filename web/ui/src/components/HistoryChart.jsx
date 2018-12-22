import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Line } from 'react-chartjs-2'

import {
  createLineChart
} from './projectChartData'

export default class HistoryChart extends Component {
  constructor (props) {
    super(props)

    this.config = null
  }

  componentDidMount () {
    console.log('HistoryChart: componentDidMount')
    this.config = createLineChart(this.props.reports)
  }

  componentDidUpdate (prevProps) {
    console.log('HistoryChart: componentDidUpdate')
    console.log(prevProps)
    console.log(this.props)
    if (!this.config ||
      (prevProps.projectId !== this.props.projectId)) {
      this.config = createLineChart(this.props.reports)
    }
  }

  render () {
    console.log(this.config)
    if (!this.config) {
      return (
        <Pane />
      )
    }

    return (
      <Pane>
        <Line data={this.config.data} options={this.config.options} />
      </Pane>
    )
  }
}
