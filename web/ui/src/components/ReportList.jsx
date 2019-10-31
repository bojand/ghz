import React, { Component } from 'react'
import { Table, Heading, IconButton, Pane, Tooltip, Button, Text, Checkbox, toaster } from 'evergreen-ui'
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
      selected: {},
      selectAllChecked: false
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
    const selectedIds = Object.keys(this.state.selected)
    if (selectedIds.length === 2 && selectedIds[0] && selectedIds[1]) {
      this.props.history.push(`/compare/${selectedIds[0]}/${selectedIds[1]}`)
    }
  }

  async deleteBulk () {
    const selectedIds = (Object.keys(this.state.selected)).map(v => Number.parseInt(v))
    const res = await this.props.reportStore.deleteReports(selectedIds)
    if (res && typeof res.deleted === 'number') {
      toaster.success(`Deleted ${res.deleted} reports.`)
      this.props.reportStore.fetchReports(
        this.state.ordering, this.state.sort, this.state.page, this.state.projectId)
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

  onSelectAll (checked) {
    const { selected } = this.state
    const reports = this.props.reportStore.state.reports
    reports.forEach(report => {
      if (checked) {
        selected[report.id] = checked
      } else {
        delete selected[report.id]
      }
    })

    this.setState({ selected, selectAllChecked: checked })
  }

  onCheckChange (id, checked) {
    const { selected } = this.state
    if (checked) {
      selected[id] = checked
    } else {
      delete selected[id]
    }

    this.setState({ selected })
  }

  render () {
    const { state: { reports, total } } = this.props.reportStore

    const totalPerPage = 20

    const nSelected = Object.keys(this.state.selected).length

    return (
      <Pane>
        <Pane display='flex' alignItems='center' marginTop={0}>
          <Pane flex={1}>
            <Heading size={600}>REPORTS</Heading>
          </Pane>
          <Pane>
            <Button
              iconBefore='comparison'
              appearance='minimal'
              intent='none'
              disabled={!(nSelected === 2)}
              onClick={() => this.compare()}
              marginRight={12}
            >
              COMPARE
            </Button>
            <Button
              iconBefore='trash'
              appearance='minimal'
              intent='danger'
              disabled={nSelected === 0}
              onClick={() => this.deleteBulk()}
            >
              DELETE
            </Button>
          </Pane>
        </Pane>

        <Table marginY={20}>
          <Table.Head>
            <Table.TextHeaderCell maxWidth={40}>
              <Pane display='flex'>
                <Checkbox
                  name='select-all-chk'
                  onChange={e => this.onSelectAll(e.target.checked)}
                  checked={this.state.selectAllChecked}
                />
              </Pane>
            </Table.TextHeaderCell>
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
                <Table.TextCell maxWidth={40}>
                  <Checkbox
                    name={`chk-${p.id}`}
                    onChange={e => this.onCheckChange(p.id, e.target.checked)}
                    checked={this.state.selected[p.id]}
                  />
                </Table.TextCell>
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
                  display='flex' textAlign='center' alignItems='center' maxWidth={80}
                >
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
