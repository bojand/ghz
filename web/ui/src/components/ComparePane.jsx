import React, { Component } from 'react'
import { Heading, Pane, Tooltip, Text, Table, Strong, Icon } from 'evergreen-ui'
import { Bar } from 'react-chartjs-2'

import {
  formatNano,
  formatFloat
} from '../lib/common'

import { colors } from '../lib/colors'

import {
  createComparisonChart
} from '../lib/compareBarChart'

export default class ComparePane extends Component {
  componentDidMount () {
    this.props.compareStore.fetchReports(this.props.reportId1, this.props.reportId2)
  }

  render () {
    const { state: { report1, report2 } } = this.props.compareStore

    const color1 = colors.orange
    const color2 = colors.skyBlue

    if (!report1 || !report1.id) {
      return (<Pane />)
    }

    const maxWidthLabel = 100

    let latKey = 0

    const report1Name = report1.name
      ? `${report1.name} [ID:${report1.id}]`
      : `Report: ${report1.id}`

    const report2Name = report2.name
      ? `${report2.name} [ID:${report2.id}]`
      : `Report: ${report2.id}`

    const config = createComparisonChart(report1, report2, color1, color2)

    return (
      <Pane marginTop={6}>
        <Pane>
          <Heading size={500}>REPORT COMPARISON</Heading>
        </Pane>

        <Pane display='flex' marginTop={16}>
          <Pane>
            <Icon icon='full-circle' size={12} color={color1} marginRight={10} /><Text size={500}>{report1Name}</Text>
          </Pane>
          <Pane marginLeft={20} >
            <Icon icon='full-circle' size={12} color={color2} marginRight={10} /><Text size={500}>{report2Name}</Text>
          </Pane>
        </Pane>

        <Pane marginTop={24} maxWidth={840}>
          <Bar data={config.data} options={config.options} />
        </Pane>

        <Pane display='flex' marginTop={24} marginBottom={24}>

          <Pane maxWidth={400}>
            <Heading>
              Summary
            </Heading>
            <Pane>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel} />
                <Table.TextCell>
                  <Tooltip content={report1Name}>
                    <Text size={500} color={color1}>{report1Name}</Text>
                  </Tooltip>
                </Table.TextCell>
                <Table.TextCell>
                  <Tooltip content={report2Name}>
                    <Text size={500} color={color2}>{report2Name}</Text>
                  </Tooltip>
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}>
                  <Strong>Count</Strong>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {report1.count}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {report2.count}
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Total</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report1.total)} ms
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report2.total)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Average</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report1.average)} ms
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report2.average)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Slowest</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report1.slowest)} ms
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report2.slowest)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>Fastest</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report1.fastest)} ms
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(report2.fastest)} ms
                </Table.TextCell>
              </Table.Row>
              <Table.Row>
                <Table.TextCell maxWidth={maxWidthLabel}><Strong>RPS</Strong></Table.TextCell>
                <Table.TextCell isNumber>
                  {formatFloat(report1.rps)}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatFloat(report2.rps)}
                </Table.TextCell>
              </Table.Row>
            </Pane>
          </Pane>

          <Pane flex={1} maxWidth={400} marginLeft={20}>
            <Heading>
              Latency Distribution
            </Heading>
            <Pane>
              <Table.Row>
                <Table.TextCell maxWidth={60} />
                <Table.TextCell>
                  <Tooltip content={report1Name}>
                    <Text size={500} color={color1}>{report1Name}</Text>
                  </Tooltip>
                </Table.TextCell>
                <Table.TextCell>
                  <Tooltip content={report2Name}>
                    <Text size={500} color={color2}>{report2Name}</Text>
                  </Tooltip>
                </Table.TextCell>
              </Table.Row>
              {report1.latencyDistribution.map((p, i) => (
                <Table.Row key={'lat-' + latKey++}>
                  <Table.TextCell maxWidth={60}>
                    <Strong>{p.percentage} %</Strong>
                  </Table.TextCell>
                  <Table.TextCell isNumber>
                    {formatNano(p.latency)} ms
                  </Table.TextCell>
                  <Table.TextCell isNumber>
                    {formatNano(report2.latencyDistribution[i].latency)} ms
                  </Table.TextCell>
                </Table.Row>
              ))}
            </Pane>
          </Pane>
        </Pane>
      </Pane>

    )
  }
}
