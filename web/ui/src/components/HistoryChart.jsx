import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Line } from 'react-chartjs-2'

import {
  createLineChart
} from '../lib/projectChartData'

export default class HistoryChart extends Component {
  constructor (props) {
    super(props)

    this.config = null
  }

  componentDidMount () {
    this.config = createLineChart(this.props.reports)
  }

  componentDidUpdate (prevProps) {
    if (!this.config ||
      (prevProps.projectId !== this.props.projectId)) {
      this.config = createLineChart(this.props.reports)
    }
  }

  render () {
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
