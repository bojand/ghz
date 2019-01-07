import React, { Component } from 'react'
import {
  Pane, Heading, Strong, Table, Icon,
  Tooltip, Text, Badge, Button, Link, CornerDialog
} from 'evergreen-ui'
import _ from 'lodash'
import { Provider, Subscribe } from 'unstated'
import { Link as RouterLink } from 'react-router-dom'

import {
  formatNanoUnit,
  formatFloat,
  toLocaleString,
  getAppRoot
} from '../lib/common'

import StatusCodeChart from './ReportDistChart'
import OptionsPane from './OptionsPane'
import LatencyPane from './LatencyPane'
import HistogramPane from './HistogramPane'
import StatusBadge from './StatusBadge'

import HistogramContainer from '../containers/HistogramContainer'
import OptionsContainer from '../containers/OptionsContainer'

export default class ReportDetailPane extends Component {
  constructor (props) {
    super(props)

    this.state = {
      reportId: props.reportId || 0
    }
  }

  componentDidMount () {
    this.props.compareStore.fetchReports(this.props.reportId, 'previous')
  }

  componentDidUpdate (prevProps) {
    if (!this.props.compareStore.state.isFetching &&
      (this.props.reportId !== prevProps.reportId)) {
      this.props.compareStore.fetchReports(this.props.reportId, 'previous')
    }
  }

  render () {
    const currentReport = this.props.compareStore.state.report1
    const prevReport = this.props.compareStore.state.report2

    if (!currentReport || !currentReport.id) {
      return (<Pane />)
    }

    const dateStr = toLocaleString(currentReport.date)

    let statusKey = 0
    let errKey = 0
    let tagKey = 0

    const appRoot = getAppRoot()

    return (
      <Pane>

        <Pane display='flex' marginTop={6} marginBottom={10}>
          <Pane flex={1}>
            <Pane display='flex' textAlign='center' alignItems='center'>
              <StatusBadge status={currentReport.status} marginRight={8} />
              <Heading size={500}>
                {currentReport.name || `REPORT: ${currentReport.id}`}
              </Heading>
            </Pane>
            <Pane marginTop={8}>
              <Text>
                {dateStr}
              </Text>
            </Pane>
            {currentReport.tags && _.keys(currentReport.tags).length
              ? <Pane marginTop={12}>
                {_.map(currentReport.tags, (v, k) => (
                  <Badge color='blue' marginRight={8} marginBottom={8} key={'tag-' + tagKey++}>
                    {`${k}: ${v}`}
                  </Badge>
                ))}
              </Pane>
              : <Pane />
            }
          </Pane>

          <Pane>
            <Pane display='flex'>
              <RouterLink to={`/compare/${currentReport.id}/previous`}>
                <Button iconBefore='comparison' appearance='minimal' intent='none' height={32} marginRight={12}>
                  COMPARE TO PREVIOUS
                </Button>
              </RouterLink>
              <Link href={`${appRoot}/api/reports/${currentReport.id}/export?format=json`} target='_blank'>
                <Button iconBefore='code' appearance='minimal' intent='none' height={32} marginRight={12}>
                  JSON
                </Button>
              </Link>
              <Link href={`${appRoot}/api/reports/${currentReport.id}/export?format=csv`} target='_blank'>
                <Button iconBefore='label' appearance='minimal' intent='none' height={32}>
                  CSV
                </Button>
              </Link>
            </Pane>
          </Pane>
        </Pane>

        <Pane display='flex' paddingY={20}>
          <Pane flex={1} minWidth={260} maxWidth={360}>
            <Heading>
              Summary
            </Heading>
            <Pane borderBottom paddingY={16}>
              <Pane>
                <Text size={300} color='muted' fontWeight='normal'>
                  COUNT
                </Text>
              </Pane>
              <Pane display='flex' paddingTop={8}>
                <Pane flex={1} display='flex'>
                  <Text size={500} fontFamily='mono'>{currentReport.count}</Text>
                </Pane>
              </Pane>
            </Pane>
            <SummaryPropComponent currentReport={currentReport} previousReport={prevReport} propName='total' />
            <SummaryPropComponent currentReport={currentReport} previousReport={prevReport} propName='average' />
            <SummaryPropComponent currentReport={currentReport} previousReport={prevReport} propName='fastest' />
            <SummaryPropComponent currentReport={currentReport} previousReport={prevReport} propName='slowest' />
            <SummaryPropComponent currentReport={currentReport} previousReport={prevReport} propName='rps' />
          </Pane>

          <Pane flex={1} marginLeft={30} marginRight={16}>
            <LatencyComponent
              currentReport={currentReport} previousReport={prevReport}
            />
          </Pane>
          <Pane flex={1} />

        </Pane>

        <Pane display='flex' alignItems='left' marginTop={24} marginBottom={24}>
          <Pane flex={3}>
            <Provider>
              <Subscribe to={[HistogramContainer]}>
                {histogramStore => (
                  <HistogramPane
                    histogramStore={histogramStore}
                    reportId={currentReport.id}
                    count={currentReport.count}
                  />
                )}
              </Subscribe>
            </Provider>
          </Pane>
          <Pane flex={1} />
        </Pane>

        <Pane>
          <Heading>
            Status Code Distribution
          </Heading>
          <Pane display='flex' marginTop={16} alignItems='left'>
            <Pane minWidth={400} maxWidth={500} alignItems='left'>
              <StatusCodeChart
                report={currentReport}
                label='Status Code'
                dataMapKey='statusCodeDistribution'
              />
            </Pane>
            <Pane flex={1} marginLeft={16}>
              {_.map(currentReport.statusCodeDistribution, (v, k) => (
                <Table.Row key={'status-' + statusKey++}>
                  <Table.TextCell textProps={{ size: 400 }} minWidth={200}>
                    {k.toString().toUpperCase()}
                  </Table.TextCell>
                  <Table.TextCell isNumber>
                    {v.toString()}
                    <Tooltip content='Percent of all calls'>
                      <Text marginLeft={8} textDecoration='underline dotted'>
                        {'(' + Number.parseFloat((v / currentReport.count) * 100).toFixed(2) + ' %)'}
                      </Text>
                    </Tooltip>
                  </Table.TextCell>
                </Table.Row>
              ))}
            </Pane>
            <Pane flex={1} />
          </Pane>
        </Pane>

        {currentReport.errorDistribution && _.keys(currentReport.errorDistribution).length
          ? <Pane marginTop={30}>
            <Heading>
              Error Distribution
            </Heading>
            <Pane display='flex' marginTop={16} alignItems='left'>
              <Pane minWidth={400} maxWidth={500} alignItems='left'>
                <StatusCodeChart
                  report={currentReport}
                  label='Error Distribution'
                  dataMapKey='errorDistribution'
                />
              </Pane>
              <Pane flex={1} marginLeft={16}>
                {_.map(currentReport.errorDistribution, (v, k) => (
                  <Table.Row key={'error-' + errKey++}>
                    <Table.TextCell textProps={{ size: 400 }}>
                      {k.toString()}
                    </Table.TextCell>
                    <Table.TextCell isNumber maxWidth={120}>
                      {v.toString()}
                      <Tooltip content='Percent of all calls'>
                        <Text marginLeft={8} textDecoration='underline dotted'>
                          {'(' + Number.parseFloat((v / currentReport.count) * 100).toFixed(2) + ' %)'}
                        </Text>
                      </Tooltip>
                    </Table.TextCell>
                  </Table.Row>
                ))}
              </Pane>
            </Pane>
          </Pane>
          : <Pane />
        }

        <Pane display='flex' alignItems='left' marginTop={32} marginBottom={24}>
          <Pane flex={3} >
            <Provider>
              <Subscribe to={[OptionsContainer]}>
                {optionsStore => (
                  <OptionsPane
                    optionsStore={optionsStore}
                    reportId={currentReport.id}
                  />
                )}
              </Subscribe>
            </Provider>
          </Pane>
          <Pane flex={2} />
        </Pane>
      </Pane>
    )
  }
}

const SummaryPropComponent = ({ currentReport, previousReport, propName }) => {
  const crVal = currentReport[propName]
  const prVal = previousReport[propName]
  const change = crVal - prVal
  const changeAbs = Math.abs(change)
  const changeP = change > 0
    ? crVal / prVal * 100 - 100
    : 100 - crVal / prVal * 100
  const changeIcon = change > 0
    ? 'arrow-up'
    : 'arrow-down'
  let changeColor = change > 0
    ? 'danger'
    : 'success'

  if (propName === 'rps') {
    changeColor = change > 0
      ? 'success'
      : 'danger'
  }

  let label = propName

  if (propName === 'rps') {
    label = 'Requests Per Second'
  }

  return (
    <Pane borderBottom paddingY={16}>
      <Pane>
        <Text size={300} color='muted' fontWeight='normal'>
          {_.toUpper(label)}
        </Text>
      </Pane>
      <Pane display='flex' paddingTop={8}>
        <Pane flex={2} display='flex'>
          <Text size={500} fontFamily='mono'>
            {propName === 'rps' ? formatFloat(crVal) : formatNanoUnit(crVal)}
          </Text>
        </Pane>
        <Pane flex={3} display='flex'>
          <Icon icon={changeIcon} color={changeColor} marginRight={8} />
          <Text fontFamily='mono'>
            {propName === 'rps' ? formatFloat(changeAbs) : formatNanoUnit(changeAbs)} ({formatFloat(changeP)} %)
          </Text>
        </Pane>
      </Pane>
    </Pane>
  )
}

const LatencyPropComponent = ({ currentReportLD, previousReportLD, propName }) => {
  const crVal = currentReportLD.latency
  const prVal = previousReportLD.latency
  const change = crVal - prVal
  const changeAbs = Math.abs(change)
  const changeP = change > 0
    ? crVal / prVal * 100 - 100
    : 100 - crVal / prVal * 100
  const changeIcon = change > 0
    ? 'arrow-up'
    : 'arrow-down'
  let changeColor = change > 0
    ? 'danger'
    : 'success'

  let label = currentReportLD.percentage

  return (
    <Pane borderBottom paddingY={16} display='flex'>
      <Pane flex={2}>
        <Text fontWeight='bold' size={500}>
          {_.toUpper(label)} %
        </Text>
      </Pane>
      <Pane flex={3} display='flex'>
        <Text fontFamily='mono'>
          {formatNanoUnit(crVal)}
        </Text>
      </Pane>
      <Pane flex={5} display='flex'>
        <Icon icon={changeIcon} color={changeColor} marginRight={8} />
        <Text fontFamily='mono'>
          {formatNanoUnit(changeAbs)} ({formatFloat(changeP)} %)
        </Text>
      </Pane>
    </Pane>
  )
}

const LatencyComponent = ({ currentReport, previousReport }) => {
  return (
    <Pane>
      <Heading>
        Latency Distribution
      </Heading>
      <Pane>
        {currentReport.latencyDistribution.map((p, i) => (
          <LatencyPropComponent currentReportLD={p} previousReportLD={previousReport.latencyDistribution[i]} propName={p} />
        ))}
      </Pane>
    </Pane>
  )
}
