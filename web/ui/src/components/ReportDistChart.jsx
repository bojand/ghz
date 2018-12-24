import React, { Component } from 'react'
import { Pane } from 'evergreen-ui'
import { Doughnut } from 'react-chartjs-2'

import {
  createDoughnutChart
} from '../lib/doughnutData'

export default class ReportDist extends Component {
  constructor (props) {
    super(props)

    this.config = null

    this.state = {
      config: this.config
    }
  }

  componentDidMount () {
    this.config = createDoughnutChart(
      this.props.label,
      this.props.report[this.props.dataMapKey]
    )

    this.setState({
      config: this.config
    })
  }

  componentDidUpdate (prevProps) {
    if (!this.config ||
      (prevProps.report.id !== this.props.report.id)) {
      this.config = createDoughnutChart(
        this.props.label,
        this.props.report[this.props.dataMapKey]
      )
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
      <Pane alignItems='left'>
        <Doughnut data={config.data} options={config.options} />
      </Pane>
    )
  }
}
