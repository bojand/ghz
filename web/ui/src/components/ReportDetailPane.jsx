import React, { Component } from 'react'
import { Pane, Heading, Icon, Pre, Strong, Table } from 'evergreen-ui'

import {
  Order,
  getIconForOrder,
  getIconForMetricStatus,
  getIconForStatus,
  getColorForStatus,
  formatNano,
  pretty
} from '../lib/common'

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

    if (!currentReport) {
      return (<Pane />)
    }

    const maxWidthLabel = 100

    return (
      <Pane>
        <Pane display='flex' alignItems='center' marginTop={6} marginBottom={10}>
          <Icon
            marginRight={16}
            icon={getIconForStatus(currentReport.status)}
            color={getColorForStatus(currentReport.status)} />
          <Heading size={500}>
            {currentReport.name || `REPORT: ${currentReport.id}`}
          </Heading>
        </Pane>
        <Pane display='flex'>
          <Pane flex={1} paddingX={16} paddingY={16}>
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
                  <Pane display='flex'>
                    {formatNano(currentReport.average)} ms
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(currentReport.averageStatus)}
                      color={getColorForStatus(currentReport.averageStatus)} />
                  </Pane>
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Slowest</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(currentReport.slowest)} ms
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(currentReport.slowestStatus)}
                      color={getColorForStatus(currentReport.slowestStatus)} />
                  </Pane>
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Fastest</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(currentReport.fastest)} ms
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(currentReport.fastestStatus)}
                      color={getColorForStatus(currentReport.fastestStatus)} />
                  </Pane>
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>RPS</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {currentReport.rps}
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(currentReport.rpsStatus)}
                      color={getColorForStatus(currentReport.rpsStatus)} />
                  </Pane>
                </Table.TextCell>
              </Table.Row>
            </Pane>
          </Pane>
          <Pane flex={2} paddingY={16}>
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
      </Pane>
    )
  }
}
