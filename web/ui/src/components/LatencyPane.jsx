import React, { Component } from 'react'
import { Pane, Heading, Table, Strong } from 'evergreen-ui'

import { formatNano } from '../lib/common'

export default class LatencyPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      reportId: props.reportId || 0
    }
  }

  componentDidMount () {
    this.props.latencyStore.fetchLatencyDistribution(this.props.reportId)
  }

  componentDidUpdate (prevProps) {
    if (!this.props.latencyStore.state.isFetching &&
      (this.props.reportId !== prevProps.reportId)) {
      this.props.latencyStore.fetchLatencyDistribution(this.props.reportId)
    }
  }

  render () {
    const { state: { latencyDistribution } } = this.props.latencyStore

    if (!latencyDistribution || !latencyDistribution.length) {
      return (<Pane />)
    }
    let latKey = 0

    return (
      <Pane>
        <Heading>
          Latency Distribution
        </Heading>
        <Pane paddingY={10}>
          {latencyDistribution.map(p => (
            <Table.Row key={'lat-' + latKey++}>
              <Table.TextCell maxWidth={60}>
                <Strong>{p.percentage} %</Strong>
              </Table.TextCell>
              <Table.TextCell isNumber>
                {formatNano(p.latency)} ms
              </Table.TextCell>
            </Table.Row>
          ))}
        </Pane>
      </Pane>
    )
  }
}
