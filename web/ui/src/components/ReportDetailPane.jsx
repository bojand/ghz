import React, { Component } from 'react'
import { Pane, Heading, Icon, Pre, Strong, Table } from 'evergreen-ui'
import _ from 'lodash'

import {
  getIconForStatus,
  getColorForStatus,
  formatNano,
  pretty
} from '../lib/common'

import HistogramChart from './HistogramChart'

export default class ProjectDetailPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      reportId: props.reportId || 0
    }
  }

  componentDidMount () {
    this.props.reportStore.fetchReport(this.props.reportId)
  }

  componentDidUpdate (prevProps) {
    if (!this.props.reportStore.state.isFetching &&
      (this.props.reportId !== prevProps.reportId)) {
      this.props.reportStore.fetchReport(this.props.reportId)
    }
  }

  render () {
    const { state: { currentReport } } = this.props.reportStore

    if (!currentReport || !currentReport.id) {
      return (<Pane />)
    }

    const maxWidthLabel = 100

    return (
      <Pane>
        <Pane display='flex' marginTop={6} marginBottom={10}>
          <Icon
            marginRight={16}
            icon={getIconForStatus(currentReport.status)}
            color={getColorForStatus(currentReport.status)} />
          <Heading size={500}>
            {currentReport.name || `REPORT: ${currentReport.id}`}
          </Heading>
        </Pane>

        <Pane display='flex'>
          <Pane flex={1} paddingY={16} minWidth={260} maxWidth={260}>
            <Heading>
              Summary
            </Heading>
            <Pane marginTop={16}>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}>
                  <Strong>Count</Strong>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {currentReport.count}
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Total</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(currentReport.total)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Average</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(currentReport.average)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Slowest</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(currentReport.slowest)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Fastest</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(currentReport.fastest)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>RPS</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {currentReport.rps}
                </Table.TextCell>
              </Table.Row>
            </Pane>
          </Pane>
          <Pane flex={2} paddingY={16} marginLeft={20}>
            <Heading>
              Options
            </Heading>
            <Pane background='tint2' marginTop={16}>
              <Pre fontFamily='monospace' paddingX={16} paddingY={16}>
                {pretty(currentReport.options)}
              </Pre>
            </Pane>
          </Pane>
        </Pane>

        <Pane display='flex' alignItems='left' marginTop={24} marginBottom={24}>
          <Pane flex={4}>
            <Pane>
              <Heading size={500}>Histogram</Heading>
            </Pane>
            <Pane paddingY={20}>
              <HistogramChart report={currentReport} />
            </Pane>
          </Pane>
          <Pane flex={1} marginLeft={30} marginRight={16}>
            <Heading>
              Latency Distribution
            </Heading>
            <Pane paddingY={10}>
              {currentReport.latencyDistribution.map(p => (
                <Table.Row>
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
        </Pane>

        <Pane display='flex'>
          <Pane flex={1} paddingY={16} paddingX={16} minWidth={250} maxWidth={250}>
            <Heading>
              Status Code Distribution
            </Heading>
            {_.map(currentReport.statusCodeDistribution, (v, k) => (
              <Table.Row>
                <Table.TextCell maxWidth={60}>
                  <Strong>{k.toString()}</Strong>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {v.toString() + ' (' + Number.parseFloat((v / currentReport.count) * 100).toFixed(2) + ' %)'}
                </Table.TextCell>
              </Table.Row>
            ))}
          </Pane>
          <Pane flex={3} paddingY={16} />
        </Pane>
      </Pane>
    )
  }
}
