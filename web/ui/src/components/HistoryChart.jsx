import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Line } from 'react-chartjs-2'

import {
  createLineChart
} from '../lib/projectChartData'

export default class HistoryChart extends Component {
  constructor (props) {
    super(props)

    this.state = {
      config: createLineChart(this.props.reports)
    }
  }

  componentDidUpdate (prevProps) {
    if ((prevProps.projectId !== this.props.projectId) ||
      (prevProps.reports.length !== this.props.reports.length)) {
      const config = createLineChart(this.props.reports)
      this.setState({ config })
    }
  }

  render () {
    const { config } = this.state
    if (!config) {
      return (
        <Pane />
      )
    }

    return (
      <Pane>
        <Line data={config.data} options={config.options} />
      </Pane>
    )
  }
}
