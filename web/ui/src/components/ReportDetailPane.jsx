import React, { Component } from 'react'
import {
  Pane, Heading, Table, Icon,
  Tooltip, Text, Badge, Button, toaster
} from 'evergreen-ui'
import _ from 'lodash'
import { Provider, Subscribe } from 'unstated'
import { Link as RouterLink, withRouter } from 'react-router-dom'

import {
  formatNanoUnit,
  formatFloat,
  toLocaleString,
  getAppRoot
} from '../lib/common'

import StatusCodeChart from './ReportDistChart'
import OptionsPane from './OptionsPane'
import HistogramPane from './HistogramPane'
import StatusBadge from './StatusBadge'
import DeleteDialog from './DeleteDialog'
import HistogramContainer from '../containers/HistogramContainer'
import OptionsContainer from '../containers/OptionsContainer'

class ReportDetailPane extends Component {
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

  async deleteReport () {
    this.setState({ deleteVisible: false })
    const currentReport = this.props.compareStore.state.report1
    const id = this.props.reportId
    const name = currentReport && currentReport.name ? currentReport.name : id

    const ok = await this.props.reportStore.deleteReport(id)
    if (ok) {
      toaster.success(`Report ${name} deleted.`)
      this.props.history.push('/projects')
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
              <Heading size={600}>
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
              : <Pane />}
          </Pane>

          <Pane>
            <Pane display='flex'>
              <RouterLink to={`/compare/${currentReport.id}/previous`} style={{ textDecoration: 'none' }}>
                {prevReport &&
                  <Button iconBefore='comparison' appearance='minimal' intent='none' height={32} marginRight={12}>
                    COMPARE TO PREVIOUS
                  </Button>}
              </RouterLink>
              <Button is='a' href={`${appRoot}/api/reports/${currentReport.id}/export?format=json`} target='_blank' iconBefore='code' appearance='minimal' intent='none' height={32} marginRight={12}>
                JSON
              </Button>
              <Button is='a' href={`${appRoot}/api/reports/${currentReport.id}/export?format=csv`} target='_blank' iconBefore='label' appearance='minimal' intent='none' height={32} marginRight={12}>
                CSV
              </Button>
              {this.state.deleteVisible
                ? <DeleteDialog
                  dataType='report'
                  name={currentReport.name || currentReport.id}
                  isShown={this.state.deleteVisible}
                  onConfirm={() => this.deleteReport()}
                  onCancel={() => this.setState({ deleteVisible: !this.state.deleteVisible })}
                  /> : null}
              <Button
                iconBefore='trash'
                appearance='minimal'
                intent='danger'
                onClick={() => this.setState({ deleteVisible: !this.state.deleteVisible })}
              >DELETE
              </Button>
            </Pane>
          </Pane>
        </Pane>

        <Pane display='flex' paddingY={8}>
          <Pane flex={1} minWidth={260} maxWidth={360}>
            <Heading size={600}>
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
          <Heading size={600}>
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
            <Heading size={600}>
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
          : <Pane />}

        <Pane display='flex' alignItems='left' marginTop={32} marginBottom={24}>
          <Pane flex={3}>
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
  const prVal = previousReport ? previousReport[propName] : -1
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
        {previousReport
          ? <Pane flex={3} display='flex'>
            <Icon icon={changeIcon} color={changeColor} marginRight={8} />
            <Text fontFamily='mono'>
              {propName === 'rps' ? formatFloat(changeAbs) : formatNanoUnit(changeAbs)} ({formatFloat(changeP)} %)
            </Text>
            </Pane>
          : <Pane />}
      </Pane>
    </Pane>
  )
}

const LatencyPropComponent = ({ currentReportLD, previousReportLD }) => {
  const crVal = currentReportLD.latency
  const prVal = previousReportLD ? previousReportLD.latency : -1
  const change = crVal - prVal
  const changeAbs = Math.abs(change)
  const changeP = change > 0
    ? crVal / prVal * 100 - 100
    : 100 - crVal / prVal * 100
  const changeIcon = change > 0
    ? 'arrow-up'
    : 'arrow-down'
  const changeColor = change > 0
    ? 'danger'
    : 'success'

  const label = currentReportLD.percentage

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
      {previousReportLD
        ? <Pane flex={5} display='flex'>
          <Icon icon={changeIcon} color={changeColor} marginRight={8} />
          <Text fontFamily='mono'>
            {formatNanoUnit(changeAbs)} ({formatFloat(changeP)} %)
          </Text>
          </Pane>
        : <Pane />}
    </Pane>
  )
}

const LatencyComponent = ({ currentReport, previousReport }) => {
  return (
    <Pane>
      <Heading size={600}>
        Latency Distribution
      </Heading>
      <Pane>
        {currentReport.latencyDistribution.map((p, i) => (
          <LatencyPropComponent key={i} currentReportLD={p} previousReportLD={previousReport ? previousReport.latencyDistribution[i] : null} propName={p} />
        ))}
      </Pane>
    </Pane>
  )
}

export default withRouter(ReportDetailPane)
