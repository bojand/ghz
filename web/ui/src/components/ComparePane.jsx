import React, { Component } from 'react'
import { Heading, Pane, Tooltip, Text, Table, Strong, Icon, Badge } from 'evergreen-ui'
import { Bar } from 'react-chartjs-2'
import { Link as RouterLink } from 'react-router-dom'
import _ from 'lodash'

import {
  formatNanoUnit,
  formatFloat,
  toLocaleString
} from '../lib/common'

import { colors } from '../lib/colors'

import {
  createComparisonChart
} from '../lib/compareBarChart'

import StatusBadge from './StatusBadge'

export default class ComparePane extends Component {
  componentDidMount () {
    this.props.compareStore.fetchReports(this.props.reportId1, this.props.reportId2)
  }

  render () {
    const { state: { report1, report2 } } = this.props.compareStore

    const color1 = colors.orange
    const color2 = colors.skyBlue

    let tagKey = 0

    if (!report1 || !report1.id) {
      return (<Pane />)
    }

    const maxWidthLabel = 100

    let latKey = 0

    const report1Name = report1.name
      ? `${report1.name} [ID: ${report1.id}]`
      : `Report: ${report1.id}`

    const report2Name = report2.name
      ? `${report2.name} [ID: ${report2.id}]`
      : `Report: ${report2.id}`

    const config = createComparisonChart(report1, report2, color1, color2)

    const latency1 = _.sortBy(report1.latencyDistribution, 'percentage')
    const latency2 = _.sortBy(report2.latencyDistribution, 'percentage')

    return (
      <Pane marginTop={6}>
        <Pane>
          <Heading size={600}>REPORT COMPARISON</Heading>
        </Pane>

        <Pane marginTop={16} display='flex'>
          <Pane maxWidth={450}>
            <Icon icon='full-circle' size={12} color={color1} marginRight={10} />
            <RouterLink to={`/reports/${report1.id}`}>
              <Text size={500} marginRight={8}>{report1Name}</Text>
            </RouterLink>
            <StatusBadge status={report1.status} />
            <Pane marginTop={8}>
              <Text>
                {toLocaleString(report1.date)}
              </Text>
            </Pane>
            {report1.tags && _.keys(report1.tags).length
              ? <Pane marginTop={12}>
                {_.map(report1.tags, (v, k) => (
                  <Badge color='blue' marginRight={8} marginBottom={8} key={'tag1-' + tagKey++}>
                    {`${k}: ${v}`}
                  </Badge>
                ))}
              </Pane>
              : <Pane />
            }
          </Pane>

          <Pane marginLeft={32} maxWidth={450}>
            <Icon icon='full-circle' size={12} color={color2} marginRight={10} />
            <RouterLink to={`/reports/${report2.id}`}>
              <Text size={500} marginRight={8}>{report2Name}</Text>
            </RouterLink>
            <StatusBadge status={report2.status} />
            <Pane marginTop={8}>
              <Text>
                {toLocaleString(report2.date)}
              </Text>
            </Pane>
            {report2.tags && _.keys(report2.tags).length
              ? <Pane marginTop={12}>
                {_.map(report2.tags, (v, k) => (
                  <Badge color='blue' marginRight={8} marginBottom={8} key={'tag2-' + tagKey++}>
                    {`${k}: ${v}`}
                  </Badge>
                ))}
              </Pane>
              : <Pane />
            }
          </Pane>
        </Pane>

        <Pane marginTop={32} display='flex'>
          <Pane flex={4}>
            <Bar data={config.data} options={config.options} />
          </Pane>
          <Pane flex={1} />
        </Pane>

        <Pane display='flex' marginTop={24} marginBottom={24}>

          <Pane flex={2}>
            <Heading size={600}>
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
                <Table.TextCell><Text size={500}>Change</Text></Table.TextCell>
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
                <Table.TextCell />
              </Table.Row>
              <LatencyRow
                maxWidth={maxWidthLabel}
                label='Total'
                value1={report1.total}
                value2={report2.total}
              />
              <LatencyRow
                maxWidth={maxWidthLabel}
                label='Average'
                value1={report1.average}
                value2={report2.average}
              />
              <LatencyRow
                maxWidth={maxWidthLabel}
                label='Fastest'
                value1={report1.fastest}
                value2={report2.fastest}
              />
              <LatencyRow
                maxWidth={maxWidthLabel}
                label='Slowest'
                value1={report1.slowest}
                value2={report2.slowest}
              />
              <LatencyRow
                maxWidth={maxWidthLabel}
                label='RPS'
                value1={report1.rps}
                value2={report2.rps}
                invert
                floatFormat
              />
            </Pane>
          </Pane>

          <Pane flex={1} />
        </Pane>

        <Pane display='flex' marginTop={24} marginBottom={24}>
          <Pane flex={2}>
            <Heading size={600}>
              Latency Distribution
            </Heading>
            <Pane>
              <Table.Row>
                <Table.TextCell maxWidth={60} >
                  <Icon icon='percentage' />
                </Table.TextCell>
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
                <Table.TextCell>
                  <Text size={500}>Change</Text>
                </Table.TextCell>
              </Table.Row>
              {latency1.map((p, i) => (
                <LatencyRow
                  key={'lat-' + latKey++}
                  maxWidth={60}
                  label={p.percentage + ' %'}
                  value1={p.latency}
                  value2={latency2[i].latency}
                />
              ))}
            </Pane>
          </Pane>

          <Pane flex={1} />
        </Pane>
      </Pane>

    )
  }
}

const LatencyRow = ({ maxWidth, label, value1, value2, invert, floatFormat }) => {
  const change = value1 - value2
  const changeAbs = Math.abs(change)
  const changeP = change > 0
    ? value1 / value2 * 100 - 100
    : 100 - value1 / value2 * 100
  const changeIcon = change > 0
    ? 'arrow-up'
    : 'arrow-down'
  let changeColor = change > 0
    ? 'danger'
    : 'success'

  if (invert) {
    changeColor = change > 0
      ? 'success'
      : 'danger'
  }
  return (
    <Table.Row>
      <Table.TextCell maxWidth={maxWidth || 60}>
        <Strong>{label}</Strong>
      </Table.TextCell>
      <Table.TextCell isNumber>
        {floatFormat ? formatFloat(value1) : formatNanoUnit(value1)}
      </Table.TextCell>
      <Table.TextCell isNumber>
        {floatFormat ? formatFloat(value2) : formatNanoUnit(value2)}
      </Table.TextCell>
      <Table.TextCell isNumber>
        <Icon icon={changeIcon} color={changeColor} marginRight={8} />
        {floatFormat ? formatFloat(changeAbs) : formatNanoUnit(changeAbs)} ({formatFloat(changeP)} %)
      </Table.TextCell>
    </Table.Row>
  )
}
