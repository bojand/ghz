import React, { Component } from 'react'
import { Table, Heading, IconButton, Pane, Tooltip, TextInput, Button, Text } from 'evergreen-ui'
import { Link as RouterLink, withRouter } from 'react-router-dom'
import { format as formatAgo } from 'timeago.js'

import {
  Order,
  getIconForOrder,
  formatNanoUnit,
  formatFloat,
  toLocaleString
} from '../lib/common'

import StatusBadge from './StatusBadge'

class ReportList extends Component {
  constructor (props) {
    super(props)

    this.state = {
      projectId: props.projectId || 0,
      ordering: Order.DESC,
      sort: 'date',
      page: 0,
      compareId1: '',
      compareId2: ''
    }
  }

  componentDidMount () {
    this.props.reportStore.fetchReports(
      this.state.ordering, this.state.sort, this.state.page, this.state.projectId)
  }

  componentDidUpdate (prevProps) {
    if ((prevProps.projectId === this.props.projectId) &&
      !this.props.reportStore.state.isFetching) {
      const currentList = this.props.reportStore.state.reports
      const prevList = prevProps.reportStore.state.reports

      if (currentList.length === 0 && prevList.length > 0) {
        this.setState({ page: 0 })
        this.props.reportStore.fetchReports(
          this.state.ordering, this.state.sort, this.state.page, this.props.projectId)
      }
    }
  }

  sort () {
    const order = this.state.ordering === Order.ASC ? Order.DESC : Order.ASC
    this.props.reportStore.fetchReports(
      order, this.state.sort, this.state.page, this.state.projectId)
    this.setState({ ordering: order })
  }

  compare () {
    const { compareId1, compareId2 } = this.state
    if (compareId1 && compareId2) {
      this.props.history.push(`/compare/${compareId1}/${compareId2}`)
    }
  }

  fetchPage (page) {
    if (page < 0) {
      page = 0
    }

    this.props.reportStore.fetchReports(
      this.state.ordering, this.state.sort, page, this.state.projectId)

    this.setState({ page })
  }

  render () {
    const { state: { reports, total } } = this.props.reportStore

    const totalPerPage = 20

    return (
      <Pane>
        <Pane display='flex' alignItems='center' marginTop={0}>
          <Pane flex={1}>
            <Heading size={600}>REPORTS</Heading>
          </Pane>
          <Pane>
            <TextInput
              name='text-input-id1'
              placeholder='report id 1'
              marginRight={12}
              width={80}
              value={this.state.compareId1}
              onChange={ev => this.setState({ compareId1: ev.target.value })}
            />
            <TextInput
              name='text-input-id2'
              placeholder='report id 2'
              marginRight={12}
              width={80}
              value={this.state.compareId2}
              onChange={ev => this.setState({ compareId2: ev.target.value })}
            />
            <Button iconBefore='comparison' appearance='minimal' intent='none' onClick={() => this.compare()}>
              COMPARE
            </Button>
          </Pane>
        </Pane>

        <Table marginY={20}>
          <Table.Head>
            <Table.TextHeaderCell maxWidth={80} textProps={{ size: 400 }}>
              <Pane display='flex'>
                ID
              </Pane>
            </Table.TextHeaderCell>
            <Table.TextHeaderCell minWidth={280} textProps={{ size: 400 }}>
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
            <Table.TextHeaderCell maxWidth={80} textProps={{ size: 400 }}>
              Status
            </Table.TextHeaderCell>
          </Table.Head>
          <Table.Body>
            {reports.map(p => (
              <Table.Row key={p.id}>
                <Table.TextCell maxWidth={80} textProps={{ size: 400 }}>
                  {p.name
                    ? (
                      <Tooltip content={p.name}>
                        <Text size={400} textDecoration='underline dotted'>{p.id}</Text>
                      </Tooltip>
                    )
                    : p.id}
                </Table.TextCell>
                <Table.TextCell minWidth={280} textProps={{ size: 400 }}>
                  <RouterLink to={`/reports/${p.id}`}>
                    {toLocaleString(p.date)} ({formatAgo(p.date)})
                  </RouterLink>
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNanoUnit(p.total)}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNanoUnit(p.average)}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNanoUnit(p.slowest)}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatNanoUnit(p.fastest)}
                </Table.TextCell>
                <Table.TextCell isNumber>
                  {formatFloat(p.rps)}
                </Table.TextCell>
                <Table.TextCell
                  display='flex' textAlign='center' alignItems='center' maxWidth={80}>
                  <StatusBadge status={p.status} marginRight={8} />
                </Table.TextCell>
              </Table.Row>
            ))}
          </Table.Body>
          <Pane justifyContent='right' marginTop={10} display='flex'>
            <IconButton
              disabled={total < totalPerPage || this.state.page === 0}
              icon='chevron-left'
              onClick={() => this.fetchPage(this.state.page - 1)}
            />
            <IconButton
              disabled={total < totalPerPage || reports.length < totalPerPage}
              marginLeft={10}
              icon='chevron-right'
              onClick={() => this.fetchPage(this.state.page + 1)}
            />
          </Pane>
        </Table>
      </Pane>
    )
  }
}

export default withRouter(ReportList)
