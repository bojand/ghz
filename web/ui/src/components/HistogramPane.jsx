import React, { Component } from 'react'
import { Pane, Heading } from 'evergreen-ui'

import HistogramChart from './HistogramChart'

export default class HistogramPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      reportId: props.reportId || 0
    }
  }

  componentDidMount () {
    this.props.histogramStore.fetchHistogram(this.props.reportId)
  }

  componentDidUpdate (prevProps) {
    if (!this.props.histogramStore.state.isFetching &&
      (this.props.reportId !== prevProps.reportId)) {
      this.props.histogramStore.fetchHistogram(this.props.reportId)
    }
  }

  render () {
    const { state: { histogram } } = this.props.histogramStore

    if (!histogram || !histogram.length) {
      return (<Pane />)
    }

    const report = {
      id: this.props.reportId,
      histogram,
      count: this.props.count
    }

    return (
      <Pane>
        <Heading size={600}>Histogram</Heading>
        <Pane marginTop={20}>
          <HistogramChart report={report} />
        </Pane>
      </Pane>
    )
  }
}
