import React, { Component } from 'react'
import { Table, Heading, IconButton, Pane, Icon } from 'evergreen-ui'
import { Link as RouterLink } from 'react-router-dom'

import {
  Order,
  getIconForOrder,
  getIconForMetricStatus,
  getIconForStatus,
  getColorForStatus,
  formatNano
} from '../lib/common'

export default class ReportList extends Component {
  constructor (props) {
    super(props)

    this.state = {
      projectId: props.projectId || -1,
      ordering: Order.NONE
    }
  }

  componentDidMount () {
    this.props.reportStore.fetchReports()
  }

  componentDidUpdate (prevProps) {
    if ((prevProps.projectId === this.props.projectId) &&
      !this.props.reportStore.state.isFetching) {
      const currentList = this.props.reportStore.state.reports
      const prevList = prevProps.reportStore.state.reports

      if (currentList.length === 0 && prevList.length > 0) {
        this.props.reportStore.fetchReports()
      }
    }
  }

  sort () {
    this.props.reportStore.fetchReports(true)
    const order = this.state.ordering === Order.ASC ? Order.DESC : Order.ASC
    this.setState({ ordering: order })
  }

  render () {
    const { state: { reports } } = this.props.reportStore

    return (
      <Pane>
        <Pane display='flex' alignItems='center' marginTop={6}>
          <Heading size={500}>REPORTS</Heading>
        </Pane>

        <Table marginY={30}>
          <Table.Head>
            <Table.TextHeaderCell minWidth={210} textProps={{ size: 400 }}>
              <Pane display='flex'>
                Date
                <IconButton
                  marginLeft={10}
                  icon={getIconForOrder(this.state.ordering)}
                  appearance='minimal'
                  height={20}
                  onClick={() => this.sort()}
                />
              </Pane>
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }} maxWidth={100}>
              Count
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Total
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Average
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Slowest
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              Fastest
            </Table.TextHeaderCell>
            <Table.TextHeaderCell textProps={{ size: 400 }}>
              RPS
            </Table.TextHeaderCell>
            <Table.TextHeaderCell maxWidth={80}>
              Status
            </Table.TextHeaderCell>
          </Table.Head>
          <Table.Body>
            {reports.map(p => (
              <Table.Row key={p.id}>
                <Table.TextCell minWidth={210} textProps={{ size: 400 }}>
                  <RouterLink to={`/reports/${p.id}`}>
                    {p.date}
                  </RouterLink>
                </Table.TextCell>
                <Table.TextCell isNumber maxWidth={80}>
                  {p.count}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNano(p.total)} ms
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(p.average)} ms
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(p.averageStatus)}
                      color={getColorForStatus(p.averageStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(p.slowest)} ms
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(p.slowestStatus)}
                      color={getColorForStatus(p.slowestStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {formatNano(p.fastest)} ms
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(p.fastestStatus)}
                      color={getColorForStatus(p.fastestStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  <Pane display='flex'>
                    {p.rps}
                    <Icon
                      marginLeft={10}
                      icon={getIconForMetricStatus(p.rpsStatus)}
                      color={getColorForStatus(p.rpsStatus)} />
                  </Pane>
                </Table.TextCell>
                <Table.TextCell
                  display='flex' textAlign='center' maxWidth={100}>
                  <Icon
                    icon={getIconForStatus(p.status)}
                    color={getColorForStatus(p.status)} />
                </Table.TextCell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      </Pane>
    )
  }
}
