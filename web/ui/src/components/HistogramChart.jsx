import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { HorizontalBar } from 'react-chartjs-2'

import {
  createHistogramChart
} from '../lib/histogramData'

export default class HistoryChart extends Component {
  constructor (props) {
    super(props)

    this.config = null

    this.state = {
      config: this.config
    }
  }

  componentDidMount () {
    this.config = createHistogramChart(this.props.report)
    this.setState({
      config: this.config
    })
  }

  componentDidUpdate (prevProps) {
    if (!this.config ||
      (prevProps.report.id !== this.props.report.id)) {
      this.config = createHistogramChart(this.props.report)
    }
  }

  render () {
    const config = this.state.config || this.config

    if (!config) {
      return (
        <Pane />
      )
    }

    return (
      <Pane>
        <HorizontalBar data={config.data} options={config.options} />
      </Pane>
    )
  }
}
